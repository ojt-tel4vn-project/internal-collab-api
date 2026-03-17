package mocks

import "github.com/stretchr/testify/mock"

// MockAppConfigRepository mocks repository.AppConfigRepository
type MockAppConfigRepository struct {
	mock.Mock
}

func (m *MockAppConfigRepository) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockAppConfigRepository) Set(key, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}
