package mocks

import (
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/stretchr/testify/mock"
)

type MockEmployeeRepository struct {
	mock.Mock
}

func (m *MockEmployeeRepository) Create(entity *models.Employee) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockEmployeeRepository) FindAll() ([]models.Employee, error) {
	args := m.Called()
	return args.Get(0).([]models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) FindByID(id uuid.UUID) (*models.Employee, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) Update(entity *models.Employee) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockEmployeeRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockEmployeeRepository) FindByEmail(email string) (*models.Employee, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) FindByPasswordResetToken(token string) (*models.Employee, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) FindEmployeesByBirthday(month, day int) ([]models.Employee, error) {
	args := m.Called(month, day)
	return args.Get(0).([]models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) FindSubordinates(managerID uuid.UUID) ([]models.Employee, error) {
	args := m.Called(managerID)
	return args.Get(0).([]models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) FindByEmployeeCode(code string) (*models.Employee, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) FindAllBirthdays() ([]models.Employee, error) {
	args := m.Called()
	return args.Get(0).([]models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) FindByStatus(status models.Status) ([]models.Employee, error) {
	args := m.Called(status)
	return args.Get(0).([]models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) FindRoleByName(name string) (*models.Role, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockEmployeeRepository) SearchEmployees(query string) ([]models.Employee, error) {
	args := m.Called(query)
	return args.Get(0).([]models.Employee), args.Error(1)
}
