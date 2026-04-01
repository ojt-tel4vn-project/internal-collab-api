package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	leavedto "github.com/ojt-tel4vn-project/internal-collab-api/dtos/leave"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
	"github.com/ojt-tel4vn-project/internal-collab-api/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ── helpers ──────────────────────────────────────────────────

func newLeaveService(leaveRepo *mocks.MockLeaveRepository, empRepo *mocks.MockEmployeeRepository, jwt *mocks.MockJWTService) services.LeaveService {
	return services.NewLeaveService(leaveRepo, empRepo, jwt)
}

func sampleLeaveType() *models.LeaveType {
	return &models.LeaveType{
		ID:          uuid.New(),
		Name:        "Annual Leave",
		Description: "Paid annual leave",
	}
}

func sampleLeaveEmployee() *models.Employee {
	id := uuid.New()
	mgr := uuid.New()
	return &models.Employee{
		ID:        id,
		Email:     "emp@company.com",
		FullName:  "Test Employee",
		ManagerID: &mgr,
		Status:    models.StatusActive,
	}
}

// ═══════════════════════════════════════════════════════════════
// GetLeaveTypes Tests
// ═══════════════════════════════════════════════════════════════

func TestGetLeaveTypes_Success(t *testing.T) {
	leaveRepo := &mocks.MockLeaveRepository{}
	empRepo := &mocks.MockEmployeeRepository{}
	jwt := &mocks.MockJWTService{}

	leaveTypes := []models.LeaveType{
		{ID: uuid.New(), Name: "Annual Leave", Description: "Paid annual leave"},
		{ID: uuid.New(), Name: "Sick Leave", Description: "Medical leave"},
	}
	leaveRepo.On("FindLeaveTypes").Return(leaveTypes, nil)

	svc := newLeaveService(leaveRepo, empRepo, jwt)
	result, err := svc.GetLeaveTypes()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Annual Leave", result[0].Name)
	assert.Equal(t, "Sick Leave", result[1].Name)
}

func TestGetLeaveTypes_RepoError(t *testing.T) {
	leaveRepo := &mocks.MockLeaveRepository{}
	empRepo := &mocks.MockEmployeeRepository{}
	jwt := &mocks.MockJWTService{}

	leaveRepo.On("FindLeaveTypes").Return([]models.LeaveType{}, errors.New("db error"))

	svc := newLeaveService(leaveRepo, empRepo, jwt)
	result, err := svc.GetLeaveTypes()

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ═══════════════════════════════════════════════════════════════
// CreateLeaveRequest Tests
// ═══════════════════════════════════════════════════════════════

func TestCreateLeaveRequest_Success(t *testing.T) {
	leaveRepo := &mocks.MockLeaveRepository{}
	empRepo := &mocks.MockEmployeeRepository{}
	jwt := &mocks.MockJWTService{}

	emp := sampleLeaveEmployee()
	leaveType := sampleLeaveType()
	quota := &models.LeaveQuota{
		EmployeeID:  emp.ID,
		LeaveTypeID: leaveType.ID,
		Year:        2026,
		TotalDays:   12,
		UsedDays:    2,
	}
	quota.RemainingDays = quota.TotalDays - quota.UsedDays // 10 days

	leaveRepo.On("FindLeaveTypeByID", leaveType.ID).Return(leaveType, nil)
	empRepo.On("FindByID", emp.ID).Return(emp, nil)
	leaveRepo.On("FindLeaveQuota", emp.ID, leaveType.ID, 2026).Return(quota, nil)
	leaveRepo.On("CreateLeaveRequest", mock.AnythingOfType("*models.LeaveRequest")).Return(nil)

	svc := newLeaveService(leaveRepo, empRepo, jwt)
	resp, warning, err := svc.CreateLeaveRequest(emp.ID, leavedto.CreateLeaveRequest{
		LeaveTypeID: leaveType.ID,
		FromDate:    "2026-05-10",
		ToDate:      "2026-05-12", // 3 days
		Reason:      "Personal trip",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, warning) // 3 days <= 10 remaining, no warning
	assert.Equal(t, models.LeaveRequestStatusPending, resp.Status)
}

func TestCreateLeaveRequest_ExceedsQuota_Fails(t *testing.T) {
	leaveRepo := &mocks.MockLeaveRepository{}
	empRepo := &mocks.MockEmployeeRepository{}
	jwt := &mocks.MockJWTService{}

	emp := sampleLeaveEmployee()
	leaveType := sampleLeaveType()
	quota := &models.LeaveQuota{
		EmployeeID:    emp.ID,
		LeaveTypeID:   leaveType.ID,
		Year:          2026,
		TotalDays:     5,
		UsedDays:      4,
		RemainingDays: 1, // only 1 day left
	}

	leaveRepo.On("FindLeaveTypeByID", leaveType.ID).Return(leaveType, nil)
	empRepo.On("FindByID", emp.ID).Return(emp, nil)
	leaveRepo.On("FindLeaveQuota", emp.ID, leaveType.ID, 2026).Return(quota, nil)

	svc := newLeaveService(leaveRepo, empRepo, jwt)
	resp, warning, err := svc.CreateLeaveRequest(emp.ID, leavedto.CreateLeaveRequest{
		LeaveTypeID: leaveType.ID,
		FromDate:    "2026-05-10",
		ToDate:      "2026-05-15", // 6 days but only 1 remaining
		Reason:      "Holiday",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, warning)
	assert.Contains(t, err.Error(), "exceeds the limit")
}

func TestCreateLeaveRequest_InvalidDateFormat(t *testing.T) {
	leaveRepo := &mocks.MockLeaveRepository{}
	empRepo := &mocks.MockEmployeeRepository{}
	jwt := &mocks.MockJWTService{}

	svc := newLeaveService(leaveRepo, empRepo, jwt)
	resp, _, err := svc.CreateLeaveRequest(uuid.New(), leavedto.CreateLeaveRequest{
		LeaveTypeID: uuid.New(),
		FromDate:    "not-a-date",
		ToDate:      "2026-03-15",
		Reason:      "Test",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid from_date")
}

func TestCreateLeaveRequest_LeaveTypeNotFound(t *testing.T) {
	leaveRepo := &mocks.MockLeaveRepository{}
	empRepo := &mocks.MockEmployeeRepository{}
	jwt := &mocks.MockJWTService{}

	fakeTypeID := uuid.New()
	leaveRepo.On("FindLeaveTypeByID", fakeTypeID).Return(nil, nil) // nil = not found

	svc := newLeaveService(leaveRepo, empRepo, jwt)
	resp, _, err := svc.CreateLeaveRequest(uuid.New(), leavedto.CreateLeaveRequest{
		LeaveTypeID: fakeTypeID,
		FromDate:    "2026-05-10",
		ToDate:      "2026-05-12",
		Reason:      "Test",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "leave type not found")
}

// ═══════════════════════════════════════════════════════════════
// CancelLeaveRequest Tests
// ═══════════════════════════════════════════════════════════════

func TestCancelLeaveRequest_Success(t *testing.T) {
	leaveRepo := &mocks.MockLeaveRepository{}
	empRepo := &mocks.MockEmployeeRepository{}
	jwt := &mocks.MockJWTService{}

	emp := sampleLeaveEmployee()
	reqID := uuid.New()
	leaveReq := &models.LeaveRequest{
		ID:         reqID,
		EmployeeID: emp.ID,
		Status:     models.LeaveRequestStatusPending,
		FromDate:   time.Now().Add(48 * time.Hour), // future date
	}

	leaveRepo.On("FindLeaveRequestByID", reqID).Return(leaveReq, nil)
	leaveRepo.On("UpdateLeaveRequest", leaveReq).Return(nil)

	svc := newLeaveService(leaveRepo, empRepo, jwt)
	err := svc.CancelLeaveRequest(emp.ID, reqID)

	assert.NoError(t, err)
	assert.Equal(t, models.LeaveRequestStatusCanceled, leaveReq.Status)
}

func TestCancelLeaveRequest_NotOwner(t *testing.T) {
	leaveRepo := &mocks.MockLeaveRepository{}
	empRepo := &mocks.MockEmployeeRepository{}
	jwt := &mocks.MockJWTService{}

	actualOwnerID := uuid.New()
	differentEmpID := uuid.New()
	reqID := uuid.New()
	leaveReq := &models.LeaveRequest{
		ID:         reqID,
		EmployeeID: actualOwnerID,
		Status:     models.LeaveRequestStatusPending,
		FromDate:   time.Now().Add(48 * time.Hour),
	}

	leaveRepo.On("FindLeaveRequestByID", reqID).Return(leaveReq, nil)

	svc := newLeaveService(leaveRepo, empRepo, jwt)
	err := svc.CancelLeaveRequest(differentEmpID, reqID) // different employee trying to cancel

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestCancelLeaveRequest_AlreadyApproved(t *testing.T) {
	leaveRepo := &mocks.MockLeaveRepository{}
	empRepo := &mocks.MockEmployeeRepository{}
	jwt := &mocks.MockJWTService{}

	emp := sampleLeaveEmployee()
	reqID := uuid.New()
	leaveReq := &models.LeaveRequest{
		ID:         reqID,
		EmployeeID: emp.ID,
		Status:     models.LeaveRequestStatusApproved, // already approved
		FromDate:   time.Now().Add(48 * time.Hour),
	}

	leaveRepo.On("FindLeaveRequestByID", reqID).Return(leaveReq, nil)

	svc := newLeaveService(leaveRepo, empRepo, jwt)
	err := svc.CancelLeaveRequest(emp.ID, reqID)

	assert.Error(t, err)
	assert.Equal(t, services.ErrInvalidStatusTransition, err)
}

// ═══════════════════════════════════════════════════════════════
// UpdateLeaveQuota Tests
// ═══════════════════════════════════════════════════════════════

func TestUpdateLeaveQuota_Success(t *testing.T) {
	leaveRepo := &mocks.MockLeaveRepository{}
	empRepo := &mocks.MockEmployeeRepository{}
	jwt := &mocks.MockJWTService{}

	quotaID := uuid.New()
	quota := &models.LeaveQuota{
		ID:        quotaID,
		TotalDays: 12,
		UsedDays:  2,
	}

	leaveRepo.On("FindLeaveQuotaByID", quotaID).Return(quota, nil)
	leaveRepo.On("UpdateLeaveQuotaByID", quotaID, float64(15)).Return(nil)

	svc := newLeaveService(leaveRepo, empRepo, jwt)
	err := svc.UpdateLeaveQuota(quotaID, leavedto.UpdateLeaveQuotaRequest{TotalDays: 15})

	assert.NoError(t, err)
	leaveRepo.AssertExpectations(t)
}

func TestUpdateLeaveQuota_NotFound(t *testing.T) {
	leaveRepo := &mocks.MockLeaveRepository{}
	empRepo := &mocks.MockEmployeeRepository{}
	jwt := &mocks.MockJWTService{}

	quotaID := uuid.New()
	leaveRepo.On("FindLeaveQuotaByID", quotaID).Return(nil, nil)

	svc := newLeaveService(leaveRepo, empRepo, jwt)
	err := svc.UpdateLeaveQuota(quotaID, leavedto.UpdateLeaveQuotaRequest{TotalDays: 15})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUpdateLeaveQuota_NegativeDays(t *testing.T) {
	leaveRepo := &mocks.MockLeaveRepository{}
	empRepo := &mocks.MockEmployeeRepository{}
	jwt := &mocks.MockJWTService{}

	svc := newLeaveService(leaveRepo, empRepo, jwt)
	err := svc.UpdateLeaveQuota(uuid.New(), leavedto.UpdateLeaveQuotaRequest{TotalDays: -5})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-negative")
	leaveRepo.AssertNotCalled(t, "FindLeaveQuotaByID")
}
