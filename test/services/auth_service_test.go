package services_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	authdto "github.com/ojt-tel4vn-project/internal-collab-api/dtos/auth"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
	"github.com/ojt-tel4vn-project/internal-collab-api/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestMain initializes shared resources (logger) once for the whole test suite
func TestMain(m *testing.M) {
	_ = logger.InitDefaultLogger()
	defer logger.Sync()
	os.Exit(m.Run())
}

// ── helpers ──────────────────────────────────────────────────

func newAuthService(empRepo *mocks.MockEmployeeRepository, rtRepo *mocks.MockRefreshTokenRepository, jwt *mocks.MockJWTService, pw *mocks.MockPasswordService, email *mocks.MockEmailService) services.AuthService {
	// Using mock JWT — no need to call crypto.InitJWT
	return services.NewAuthService(empRepo, rtRepo, jwt, pw, email)
}

func sampleEmployee() *models.Employee {
	id := uuid.New()
	return &models.Employee{
		ID:           id,
		Email:        "john.doe@company.com",
		PasswordHash: "$2a$10$hashedpassword",
		FirstName:    "John",
		LastName:     "Doe",
		FullName:     "John Doe",
		EmployeeCode: "EMP001",
		Position:     "Developer",
		Status:       models.StatusActive,
		RoleID:       func() *uuid.UUID { id := uuid.New(); return &id }(),
		Role:         &models.Role{Name: models.RoleEmployee},
	}
}

// ═══════════════════════════════════════════════════════════════
// Login Tests
// ═══════════════════════════════════════════════════════════════

func TestLogin_Success(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	rtRepo := &mocks.MockRefreshTokenRepository{}
	jwtSvc := &mocks.MockJWTService{}
	pwSvc := &mocks.MockPasswordService{}
	emailSvc := &mocks.MockEmailService{}

	emp := sampleEmployee()

	empRepo.On("FindByEmail", emp.Email).Return(emp, nil)
	pwSvc.On("VerifyPassword", emp.PasswordHash, "correctPassword").Return(nil)
	jwtSvc.On("GenerateToken", emp.ID, emp.FullName, emp.Email, 1).Return("access-token-xyz", nil)
	rtRepo.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	svc := newAuthService(empRepo, rtRepo, jwtSvc, pwSvc, emailSvc)
	resp, err := svc.Login(&authdto.LoginRequest{Email: emp.Email, Password: "correctPassword"})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "access-token-xyz", resp.AccessToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, emp.Email, resp.User.Email)

	empRepo.AssertExpectations(t)
	pwSvc.AssertExpectations(t)
	jwtSvc.AssertExpectations(t)
	rtRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	rtRepo := &mocks.MockRefreshTokenRepository{}
	jwtSvc := &mocks.MockJWTService{}
	pwSvc := &mocks.MockPasswordService{}
	emailSvc := &mocks.MockEmailService{}

	empRepo.On("FindByEmail", "unknown@company.com").Return(nil, errors.New("not found"))

	svc := newAuthService(empRepo, rtRepo, jwtSvc, pwSvc, emailSvc)
	resp, err := svc.Login(&authdto.LoginRequest{Email: "unknown@company.com", Password: "anything"})

	assert.Error(t, err)
	assert.Nil(t, resp)
	empRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	rtRepo := &mocks.MockRefreshTokenRepository{}
	jwtSvc := &mocks.MockJWTService{}
	pwSvc := &mocks.MockPasswordService{}
	emailSvc := &mocks.MockEmailService{}

	emp := sampleEmployee()
	empRepo.On("FindByEmail", emp.Email).Return(emp, nil)
	pwSvc.On("VerifyPassword", emp.PasswordHash, "wrongPassword").Return(errors.New("mismatch"))

	svc := newAuthService(empRepo, rtRepo, jwtSvc, pwSvc, emailSvc)
	resp, err := svc.Login(&authdto.LoginRequest{Email: emp.Email, Password: "wrongPassword"})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// ═══════════════════════════════════════════════════════════════
// ChangePassword Tests
// ═══════════════════════════════════════════════════════════════

func TestChangePassword_Success(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	rtRepo := &mocks.MockRefreshTokenRepository{}
	jwtSvc := &mocks.MockJWTService{}
	pwSvc := &mocks.MockPasswordService{}
	emailSvc := &mocks.MockEmailService{}

	emp := sampleEmployee()
	empRepo.On("FindByID", emp.ID).Return(emp, nil)
	pwSvc.On("VerifyPassword", emp.PasswordHash, "oldPass123").Return(nil)
	pwSvc.On("HashPassword", "newPass456").Return("$2a$10$newhash", nil)
	empRepo.On("Update", emp).Return(nil)

	svc := newAuthService(empRepo, rtRepo, jwtSvc, pwSvc, emailSvc)
	resp, err := svc.ChangePassword(emp.ID, &authdto.ChangePasswordRequest{
		OldPassword: "oldPass123",
		NewPassword: "newPass456",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Password changed successfully", resp.Message)
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	rtRepo := &mocks.MockRefreshTokenRepository{}
	jwtSvc := &mocks.MockJWTService{}
	pwSvc := &mocks.MockPasswordService{}
	emailSvc := &mocks.MockEmailService{}

	emp := sampleEmployee()
	empRepo.On("FindByID", emp.ID).Return(emp, nil)
	pwSvc.On("VerifyPassword", emp.PasswordHash, "wrongOld").Return(errors.New("mismatch"))

	svc := newAuthService(empRepo, rtRepo, jwtSvc, pwSvc, emailSvc)
	resp, err := svc.ChangePassword(emp.ID, &authdto.ChangePasswordRequest{
		OldPassword: "wrongOld",
		NewPassword: "newPass",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// ═══════════════════════════════════════════════════════════════
// ForgotPassword Tests
// ═══════════════════════════════════════════════════════════════

func TestForgotPassword_UserExists_SendsEmail(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	rtRepo := &mocks.MockRefreshTokenRepository{}
	jwtSvc := &mocks.MockJWTService{}
	pwSvc := &mocks.MockPasswordService{}
	emailSvc := &mocks.MockEmailService{}

	emp := sampleEmployee()
	empRepo.On("FindByEmail", emp.Email).Return(emp, nil)
	empRepo.On("Update", emp).Return(nil)
	emailSvc.On("SendPasswordResetEmail", emp.Email, emp.FullName, mock.AnythingOfType("string")).Return(nil)

	svc := newAuthService(empRepo, rtRepo, jwtSvc, pwSvc, emailSvc)
	resp, err := svc.ForgotPassword(&authdto.ForgotPasswordRequest{Email: emp.Email})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.Message, "If your email is registered")
	emailSvc.AssertCalled(t, "SendPasswordResetEmail", emp.Email, emp.FullName, mock.AnythingOfType("string"))
}

func TestForgotPassword_UserNotFound_NeverReveals(t *testing.T) {
	// Security: should return same message even if user doesn't exist
	empRepo := &mocks.MockEmployeeRepository{}
	rtRepo := &mocks.MockRefreshTokenRepository{}
	jwtSvc := &mocks.MockJWTService{}
	pwSvc := &mocks.MockPasswordService{}
	emailSvc := &mocks.MockEmailService{}

	empRepo.On("FindByEmail", "ghost@company.com").Return(nil, errors.New("not found"))

	svc := newAuthService(empRepo, rtRepo, jwtSvc, pwSvc, emailSvc)
	resp, err := svc.ForgotPassword(&authdto.ForgotPasswordRequest{Email: "ghost@company.com"})

	assert.NoError(t, err) // Must NOT return error to prevent user enumeration
	assert.Contains(t, resp.Message, "If your email is registered")
	emailSvc.AssertNotCalled(t, "SendPasswordResetEmail")
}

// ═══════════════════════════════════════════════════════════════
// ResetPassword Tests
// ═══════════════════════════════════════════════════════════════

func TestResetPassword_Success(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	rtRepo := &mocks.MockRefreshTokenRepository{}
	jwtSvc := &mocks.MockJWTService{}
	pwSvc := &mocks.MockPasswordService{}
	emailSvc := &mocks.MockEmailService{}

	token := "valid-reset-token"
	expiry := time.Now().Add(1 * time.Hour)
	emp := sampleEmployee()
	emp.PasswordResetToken = &token
	emp.PasswordResetExpiresAt = &expiry

	empRepo.On("FindByPasswordResetToken", token).Return(emp, nil)
	pwSvc.On("HashPassword", "newSecurePass!").Return("$2a$10$newhash", nil)
	empRepo.On("Update", emp).Return(nil)

	svc := newAuthService(empRepo, rtRepo, jwtSvc, pwSvc, emailSvc)
	resp, err := svc.ResetPassword(&authdto.ResetPasswordRequest{
		Token:       token,
		NewPassword: "newSecurePass!",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.Message, "reset successfully")
	// Token should be cleared
	assert.Nil(t, emp.PasswordResetToken)
}

func TestResetPassword_ExpiredToken(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	rtRepo := &mocks.MockRefreshTokenRepository{}
	jwtSvc := &mocks.MockJWTService{}
	pwSvc := &mocks.MockPasswordService{}
	emailSvc := &mocks.MockEmailService{}

	token := "expired-token"
	expiry := time.Now().Add(-1 * time.Hour) // already expired
	emp := sampleEmployee()
	emp.PasswordResetToken = &token
	emp.PasswordResetExpiresAt = &expiry

	empRepo.On("FindByPasswordResetToken", token).Return(emp, nil)

	svc := newAuthService(empRepo, rtRepo, jwtSvc, pwSvc, emailSvc)
	resp, err := svc.ResetPassword(&authdto.ResetPasswordRequest{
		Token:       token,
		NewPassword: "newPass",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// ═══════════════════════════════════════════════════════════════
// RefreshToken Tests
// ═══════════════════════════════════════════════════════════════

func TestRefreshToken_Success(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	rtRepo := &mocks.MockRefreshTokenRepository{}
	jwtSvc := &mocks.MockJWTService{}
	pwSvc := &mocks.MockPasswordService{}
	emailSvc := &mocks.MockEmailService{}

	emp := sampleEmployee()
	storedToken := &models.RefreshToken{
		UserID:    emp.ID,
		Token:     "old-refresh-token",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	rtRepo.On("FindByToken", "old-refresh-token").Return(storedToken, nil)
	rtRepo.On("Revoke", "old-refresh-token").Return(nil)
	empRepo.On("FindByID", emp.ID).Return(emp, nil)
	jwtSvc.On("GenerateToken", emp.ID, emp.FullName, emp.Email, 1).Return("new-access-token", nil)
	rtRepo.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	svc := newAuthService(empRepo, rtRepo, jwtSvc, pwSvc, emailSvc)
	resp, err := svc.RefreshToken(&authdto.RefreshTokenRequest{RefreshToken: "old-refresh-token"})

	assert.NoError(t, err)
	assert.Equal(t, "new-access-token", resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	rtRepo := &mocks.MockRefreshTokenRepository{}
	jwtSvc := &mocks.MockJWTService{}
	pwSvc := &mocks.MockPasswordService{}
	emailSvc := &mocks.MockEmailService{}

	rtRepo.On("FindByToken", "bogus-token").Return(nil, errors.New("not found"))

	svc := newAuthService(empRepo, rtRepo, jwtSvc, pwSvc, emailSvc)
	resp, err := svc.RefreshToken(&authdto.RefreshTokenRequest{RefreshToken: "bogus-token"})

	assert.Error(t, err)
	assert.Nil(t, resp)
}
