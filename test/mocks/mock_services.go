package mocks

import (
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/stretchr/testify/mock"
)

// MockJWTService mocks crypto.JWTService
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(userID uuid.UUID, username, email string, expirationHours int) (string, error) {
	args := m.Called(userID, username, email, expirationHours)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateToken(tokenString string) (*crypto.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*crypto.Claims), args.Error(1)
}

func (m *MockJWTService) RefreshToken(tokenString string, expirationHours int) (string, error) {
	args := m.Called(tokenString, expirationHours)
	return args.String(0), args.Error(1)
}

// MockPasswordService mocks crypto.PasswordService
type MockPasswordService struct {
	mock.Mock
}

func (m *MockPasswordService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordService) VerifyPassword(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

// MockEmailService mocks email.EmailService
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendWelcomeEmail(to, name, tempPassword string) error {
	args := m.Called(to, name, tempPassword)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordResetEmail(to, name, resetLink string) error {
	args := m.Called(to, name, resetLink)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordChangedEmail(to, name string) error {
	args := m.Called(to, name)
	return args.Error(0)
}

func (m *MockEmailService) SendBirthdayWish(to, name string) error {
	args := m.Called(to, name)
	return args.Error(0)
}
