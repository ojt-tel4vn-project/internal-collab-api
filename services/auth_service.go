package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos/auth"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/email"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/response"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"go.uber.org/zap"
)

type AuthService interface {
	Login(req *auth.LoginRequest) (*auth.LoginResponse, error)
	RefreshToken(req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error)
	ChangePassword(employeeID uuid.UUID, req *auth.ChangePasswordRequest) (*auth.ChangePasswordResponse, error)
	FirstTimeSetup(email string, req *auth.FirstTimeSetupRequest) (*auth.FirstTimeSetupResponse, error)
	ForgotPassword(req *auth.ForgotPasswordRequest) (*auth.ForgotPasswordResponse, error)
	ResetPassword(req *auth.ResetPasswordRequest) (*auth.ResetPasswordResponse, error)
}

type authServiceImpl struct {
	employeeRepo     repository.EmployeeRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtService       crypto.JWTService
	password         crypto.PasswordService
	emailService     email.EmailService
	frontendURL      string
}

func NewAuthService(
	employeeRepo repository.EmployeeRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtService crypto.JWTService,
	password crypto.PasswordService,
	emailService email.EmailService,
	frontendURL string,
) AuthService {
	return &authServiceImpl{
		employeeRepo:     employeeRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		password:         password,
		emailService:     emailService,
		frontendURL:      frontendURL,
	}
}

func (s *authServiceImpl) Login(req *auth.LoginRequest) (*auth.LoginResponse, error) {
	employee, err := s.employeeRepo.FindByEmail(req.Email)
	if err != nil {
		logger.Warn("Login failed: user not found", zap.String("email", req.Email))
		return nil, response.Unauthorized("Invalid credentials")
	}

	err = s.password.VerifyPassword(employee.PasswordHash, req.Password)
	if err != nil {
		logger.Warn("Login failed: invalid password", zap.String("email", req.Email))
		return nil, response.Unauthorized("Invalid credentials")
	}

	// Check if this is first-time login (password starts with "TEMP_")
	requirePasswordChange := len(employee.PasswordHash) > 5 && employee.PasswordHash[:5] == "TEMP_"

	token, err := s.jwtService.GenerateToken(employee.ID, employee.FullName, employee.Email, 1) // 1 hour expiration
	if err != nil {
		logger.Error("Login failed: token generation error", zap.Error(err))
		return nil, response.InternalServerError("Failed to generate access token")
	}

	refreshTokenString, err := crypto.GenerateRandomToken(32)
	if err != nil {
		logger.Error("Login failed: refresh token generation error", zap.Error(err))
		return nil, response.InternalServerError("Failed to generate refresh token")
	}

	refreshToken := &models.RefreshToken{
		UserID:    employee.ID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.refreshTokenRepo.Create(refreshToken); err != nil {
		logger.Error("Login failed: refresh token save error", zap.Error(err))
		return nil, response.InternalServerError("Failed to save refresh token")
	}

	var roles []string
	if employee.Role != nil {
		roles = []string{employee.Role.Name}
	}

	return &auth.LoginResponse{
		AccessToken:           token,
		RefreshToken:          refreshTokenString,
		TokenType:             "Bearer",
		RequirePasswordChange: requirePasswordChange,
		User: struct {
			ID           uuid.UUID `json:"id"`
			Email        string    `json:"email"`
			Name         string    `json:"name"`
			EmployeeCode string    `json:"employee_code"`
			Status       string    `json:"status"`
			Roles        []string  `json:"roles"`
		}{
			ID:           employee.ID,
			Email:        employee.Email,
			Name:         employee.FullName,
			EmployeeCode: employee.EmployeeCode,
			Status:       string(employee.Status),
			Roles:        roles,
		},
	}, nil
}

func (s *authServiceImpl) RefreshToken(req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error) {
	// Find refresh token
	storedToken, err := s.refreshTokenRepo.FindByToken(req.RefreshToken)
	if err != nil {
		logger.Warn("RefreshToken failed: invalid or expired token")
		return nil, response.Unauthorized("Invalid or expired refresh token")
	}

	// Revoke the used token (Rotation)
	if err := s.refreshTokenRepo.Revoke(storedToken.Token); err != nil {
		logger.Error("RefreshToken failed: revocation error", zap.Error(err))
		// We could continue, but it's safer to error out or just log
	}

	// Get user to generate new claims
	employee, err := s.employeeRepo.FindByID(storedToken.UserID)
	if err != nil {
		logger.Warn("RefreshToken failed: user not found", zap.String("id", storedToken.UserID.String()))
		return nil, response.Unauthorized("User not found")
	}

	// Generate new access token
	newAccessToken, err := s.jwtService.GenerateToken(employee.ID, employee.FullName, employee.Email, 1)
	if err != nil {
		logger.Error("RefreshToken failed: token generation error", zap.Error(err))
		return nil, response.InternalServerError("Failed to generate access token")
	}

	// Generate new refresh token
	newRefreshTokenString, err := crypto.GenerateRandomToken(32)
	if err != nil {
		logger.Error("RefreshToken failed: refresh token generation error", zap.Error(err))
		return nil, response.InternalServerError("Failed to generate refresh token")
	}

	newRefreshToken := &models.RefreshToken{
		UserID:    employee.ID,
		Token:     newRefreshTokenString,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.refreshTokenRepo.Create(newRefreshToken); err != nil {
		logger.Error("RefreshToken failed: refresh token save error", zap.Error(err))
		return nil, response.InternalServerError("Failed to save refresh token")
	}

	return &auth.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshTokenString,
		TokenType:    "Bearer",
	}, nil
}

func (s *authServiceImpl) ChangePassword(employeeID uuid.UUID, req *auth.ChangePasswordRequest) (*auth.ChangePasswordResponse, error) {
	// Find employee
	employee, err := s.employeeRepo.FindByID(employeeID)
	if err != nil {
		logger.Warn("ChangePassword failed: employee not found", zap.String("employee_id", employeeID.String()))
		return nil, response.NotFound("Employee not found")
	}

	// Verify old password
	err = s.password.VerifyPassword(employee.PasswordHash, req.OldPassword)
	if err != nil {
		logger.Warn("ChangePassword failed: invalid old password", zap.String("employee_id", employeeID.String()))
		return nil, response.Unauthorized("Invalid old password")
	}

	// Hash new password
	hashedPassword, err := s.password.HashPassword(req.NewPassword)
	if err != nil {
		logger.Error("ChangePassword failed: password hashing error", zap.Error(err))
		return nil, response.InternalServerError("Failed to change password")
	}

	// Update password
	employee.PasswordHash = hashedPassword
	err = s.employeeRepo.Update(employee)
	if err != nil {
		logger.Error("ChangePassword failed: database update error", zap.Error(err))
		return nil, response.InternalServerError("Failed to update password")
	}

	logger.Info("Password changed successfully", zap.String("employee_id", employeeID.String()))

	return &auth.ChangePasswordResponse{
		Message: "Password changed successfully",
	}, nil
}

func (s *authServiceImpl) FirstTimeSetup(email string, req *auth.FirstTimeSetupRequest) (*auth.FirstTimeSetupResponse, error) {
	// Find employee by email
	employee, err := s.employeeRepo.FindByEmail(email)
	if err != nil {
		logger.Warn("FirstTimeSetup failed: employee not found", zap.String("email", email))
		return nil, response.NotFound("Employee not found")
	}

	// Verify temporary password
	err = s.password.VerifyPassword(employee.PasswordHash, req.TemporaryPassword)
	if err != nil {
		logger.Warn("FirstTimeSetup failed: invalid temporary password", zap.String("email", email))
		return nil, response.Unauthorized("Invalid temporary password")
	}

	// Hash new password
	hashedPassword, err := s.password.HashPassword(req.NewPassword)
	if err != nil {
		logger.Error("FirstTimeSetup failed: password hashing error", zap.Error(err))
		return nil, response.InternalServerError("Failed to setup password")
	}

	// Update password and activate account
	employee.PasswordHash = hashedPassword
	employee.Status = models.StatusActive
	err = s.employeeRepo.Update(employee)
	if err != nil {
		logger.Error("FirstTimeSetup failed: database update error", zap.Error(err))
		return nil, response.InternalServerError("Failed to complete setup")
	}

	// Generate token
	token, err := s.jwtService.GenerateToken(employee.ID, employee.FullName, employee.Email, 24)
	if err != nil {
		logger.Error("FirstTimeSetup failed: token generation error", zap.Error(err))
		return nil, response.InternalServerError("Failed to generate access token")
	}

	logger.Info("First-time setup completed successfully", zap.String("employee_id", employee.ID.String()))

	return &auth.FirstTimeSetupResponse{
		Message:     "Account setup completed successfully",
		AccessToken: token,
		TokenType:   "Bearer",
	}, nil
}

func (s *authServiceImpl) ForgotPassword(req *auth.ForgotPasswordRequest) (*auth.ForgotPasswordResponse, error) {
	// Find employee
	employee, err := s.employeeRepo.FindByEmail(req.Email)
	if err != nil {
		// Don't reveal if user exists
		logger.Warn("ForgotPassword: user not found", zap.String("email", req.Email))
		return &auth.ForgotPasswordResponse{
			Message: "If your email is registered, you will receive a password reset link",
		}, nil
	}

	// Generate reset token
	token, err := crypto.GenerateRandomToken(32)
	if err != nil {
		logger.Error("ForgotPassword failed: token generation error", zap.Error(err))
		return nil, response.InternalServerError("Failed to process request")
	}

	// Save token to database
	expiration := time.Now().Add(1 * time.Hour)
	employee.PasswordResetToken = &token
	employee.PasswordResetExpiresAt = &expiration

	if err := s.employeeRepo.Update(employee); err != nil {
		logger.Error("ForgotPassword failed: database update error", zap.Error(err))
		return nil, response.InternalServerError("Failed to process request")
	}

	// Send email
	resetLink := s.frontendURL + "/reset-password?token=" + token
	if s.emailService != nil {
		err := s.emailService.SendPasswordResetEmail(employee.Email, employee.FullName, resetLink)
		if err != nil {
			logger.Error("ForgotPassword failed: email sending error", zap.Error(err))
		}
	}

	logger.Info("Password reset email sent", zap.String("email", req.Email))

	return &auth.ForgotPasswordResponse{
		Message: "If your email is registered, you will receive a password reset link",
	}, nil
}

func (s *authServiceImpl) ResetPassword(req *auth.ResetPasswordRequest) (*auth.ResetPasswordResponse, error) {
	employee, err := s.employeeRepo.FindByPasswordResetToken(req.Token)
	if err != nil {
		logger.Warn("ResetPassword failed: invalid token", zap.String("token", req.Token))
		return nil, response.BadRequest("Invalid or expired reset token")
	}

	if employee.PasswordResetExpiresAt == nil || employee.PasswordResetExpiresAt.Before(time.Now()) {
		logger.Warn("ResetPassword failed: token expired")
		return nil, response.BadRequest("Invalid or expired reset token")
	}

	hashedPassword, err := s.password.HashPassword(req.NewPassword)
	if err != nil {
		logger.Error("ResetPassword failed: password hashing error", zap.Error(err))
		return nil, response.InternalServerError("Failed to reset password")
	}

	employee.PasswordHash = hashedPassword
	employee.PasswordResetToken = nil
	employee.PasswordResetExpiresAt = nil

	if err := s.employeeRepo.Update(employee); err != nil {
		logger.Error("ResetPassword failed: database update error", zap.Error(err))
		return nil, response.InternalServerError("Failed to reset password")
	}

	logger.Info("Password reset successfully", zap.String("email", employee.Email))

	return &auth.ResetPasswordResponse{
		Message: "Password has been reset successfully",
	}, nil
}
