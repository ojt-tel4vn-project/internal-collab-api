package mocks

import (
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/stretchr/testify/mock"
)

type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(token *models.RefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) FindByToken(token string) (*models.RefreshToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) Revoke(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeAllUserTokens(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}
