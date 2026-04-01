package mocks

import (
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"github.com/stretchr/testify/mock"
)

type MockLeaveRepository struct {
	mock.Mock
}

func (m *MockLeaveRepository) CreateLeaveType(t *models.LeaveType) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockLeaveRepository) FindLeaveTypes() ([]models.LeaveType, error) {
	args := m.Called()
	return args.Get(0).([]models.LeaveType), args.Error(1)
}

func (m *MockLeaveRepository) FindLeaveTypeByID(id uuid.UUID) (*models.LeaveType, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LeaveType), args.Error(1)
}

func (m *MockLeaveRepository) UpdateLeaveType(t *models.LeaveType) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockLeaveRepository) CreateLeaveQuota(q *models.LeaveQuota) error {
	args := m.Called(q)
	return args.Error(0)
}

func (m *MockLeaveRepository) FindLeaveQuotasByEmployeeAndYear(employeeID uuid.UUID, year int) ([]models.LeaveQuota, error) {
	args := m.Called(employeeID, year)
	return args.Get(0).([]models.LeaveQuota), args.Error(1)
}

func (m *MockLeaveRepository) FindLeaveQuota(employeeID uuid.UUID, leaveTypeID uuid.UUID, year int) (*models.LeaveQuota, error) {
	args := m.Called(employeeID, leaveTypeID, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LeaveQuota), args.Error(1)
}

func (m *MockLeaveRepository) FindLeaveQuotaByID(id uuid.UUID) (*models.LeaveQuota, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LeaveQuota), args.Error(1)
}

func (m *MockLeaveRepository) UpdateLeaveQuota(q *models.LeaveQuota) error {
	args := m.Called(q)
	return args.Error(0)
}

func (m *MockLeaveRepository) UpdateLeaveQuotaByID(id uuid.UUID, totalDays float64) error {
	args := m.Called(id, totalDays)
	return args.Error(0)
}

func (m *MockLeaveRepository) CreateLeaveRequest(req *models.LeaveRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockLeaveRepository) FindLeaveRequestByID(id uuid.UUID) (*models.LeaveRequest, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LeaveRequest), args.Error(1)
}

func (m *MockLeaveRepository) UpdateLeaveRequest(req *models.LeaveRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockLeaveRepository) DeleteLeaveRequest(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockLeaveRepository) FindLeaveRequestsByEmployee(employeeID uuid.UUID, page, limit int) ([]models.LeaveRequest, int64, error) {
	args := m.Called(employeeID, page, limit)
	return args.Get(0).([]models.LeaveRequest), args.Get(1).(int64), args.Error(2)
}

func (m *MockLeaveRepository) FindPendingLeaveRequestsByManager(managerID uuid.UUID, page, limit int) ([]models.LeaveRequest, int64, error) {
	args := m.Called(managerID, page, limit)
	return args.Get(0).([]models.LeaveRequest), args.Get(1).(int64), args.Error(2)
}

func (m *MockLeaveRepository) FindLeaveRequestsByManager(managerID uuid.UUID, status string, page, limit int) ([]models.LeaveRequest, int64, error) {
	args := m.Called(managerID, status, page, limit)
	return args.Get(0).([]models.LeaveRequest), args.Get(1).(int64), args.Error(2)
}

func (m *MockLeaveRepository) FindAllLeaveRequestsOverview(year, month int) ([]models.LeaveRequest, error) {
	args := m.Called(year, month)
	return args.Get(0).([]models.LeaveRequest), args.Error(1)
}

func (m *MockLeaveRepository) FindLeaveRequestByActionToken(token string) (*models.LeaveRequest, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LeaveRequest), args.Error(1)
}

func (m *MockLeaveRepository) Transaction(fn func(txRepo repository.LeaveRepository) error) error {
	args := m.Called(fn)
	// Execute the function with self as the tx repo for in-test use
	if args.Error(0) == nil {
		return fn(m)
	}
	return args.Error(0)
}
