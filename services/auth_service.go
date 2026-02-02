package services

import (
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos/auth"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/response"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"go.uber.org/zap"
)

type AuthService interface {
	Login(req *auth.LoginRequest) (*auth.LoginResponse, error)
	ChangePassword(employeeID uuid.UUID, req *auth.ChangePasswordRequest) (*auth.ChangePasswordResponse, error)
	FirstTimeSetup(email string, req *auth.FirstTimeSetupRequest) (*auth.FirstTimeSetupResponse, error)
}

type authServiceImpl struct {
	employeeRepo repository.EmployeeRepository
	jwtService   crypto.JWTService
	password     crypto.PasswordService
}

func NewAuthService(
	employeeRepo repository.EmployeeRepository,
	jwtService crypto.JWTService,
	password crypto.PasswordService,
) AuthService {
	return &authServiceImpl{
		employeeRepo: employeeRepo,
		jwtService:   jwtService,
		password:     password,
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

	token, err := s.jwtService.GenerateToken(employee.ID, employee.FullName, employee.Email, 24)
	if err != nil {
		logger.Error("Login failed: token generation error", zap.Error(err))
		return nil, response.InternalServerError("Failed to generate access token")
	}

	return &auth.LoginResponse{
		AccessToken:           token,
		TokenType:             "Bearer",
		RequirePasswordChange: requirePasswordChange,
		User: struct {
			ID           uuid.UUID `json:"id"`
			Email        string    `json:"email"`
			Name         string    `json:"name"`
			EmployeeCode string    `json:"employee_code"`
			Status       string    `json:"status"`
		}{
			ID:           employee.ID,
			Email:        employee.Email,
			Name:         employee.FullName,
			EmployeeCode: employee.EmployeeCode,
			Status:       string(employee.Status),
		},
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
