package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type LeaveRepository interface {
	// Leave Types
	CreateLeaveType(t *models.LeaveType) error
	FindLeaveTypes() ([]models.LeaveType, error)
	FindLeaveTypeByID(id uuid.UUID) (*models.LeaveType, error)
	UpdateLeaveType(t *models.LeaveType) error

	// Leave Quotas
	CreateLeaveQuota(q *models.LeaveQuota) error
	FindLeaveQuotasByEmployeeAndYear(employeeID uuid.UUID, year int) ([]models.LeaveQuota, error)
	FindLeaveQuota(employeeID uuid.UUID, leaveTypeID uuid.UUID, year int) (*models.LeaveQuota, error)
	FindLeaveQuotaByID(id uuid.UUID) (*models.LeaveQuota, error)
	UpdateLeaveQuota(q *models.LeaveQuota) error
	UpdateLeaveQuotaByID(id uuid.UUID, totalDays float64) error

	// Leave Requests
	CreateLeaveRequest(req *models.LeaveRequest) error
	FindLeaveRequestByID(id uuid.UUID) (*models.LeaveRequest, error)
	UpdateLeaveRequest(req *models.LeaveRequest) error
	DeleteLeaveRequest(id uuid.UUID) error
	FindLeaveRequestsByEmployee(employeeID uuid.UUID, page, limit int) ([]models.LeaveRequest, int64, error)
	FindPendingLeaveRequestsByManager(managerID uuid.UUID, page, limit int) ([]models.LeaveRequest, int64, error)
	FindAllLeaveRequestsOverview(year, month int) ([]models.LeaveRequest, error)
	FindLeaveRequestByActionToken(token string) (*models.LeaveRequest, error)

	// Transaction wrapper
	Transaction(fn func(txRepo LeaveRepository) error) error
}

type leaveRepository struct {
	db *gorm.DB
}

func NewLeaveRepository(db *gorm.DB) LeaveRepository {
	return &leaveRepository{db: db}
}

// Transaction wrapper
func (r *leaveRepository) Transaction(fn func(txRepo LeaveRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &leaveRepository{db: tx}
		return fn(txRepo)
	})
}

// --- Leave Types ---

func (r *leaveRepository) CreateLeaveType(t *models.LeaveType) error {
	return r.db.Create(t).Error
}

func (r *leaveRepository) FindLeaveTypes() ([]models.LeaveType, error) {
	var types []models.LeaveType
	err := r.db.Find(&types).Error
	return types, err
}

func (r *leaveRepository) FindLeaveTypeByID(id uuid.UUID) (*models.LeaveType, error) {
	var t models.LeaveType
	err := r.db.First(&t, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *leaveRepository) UpdateLeaveType(t *models.LeaveType) error {
	return r.db.Save(t).Error
}

// --- Leave Quotas ---

func (r *leaveRepository) CreateLeaveQuota(q *models.LeaveQuota) error {
	return r.db.Create(q).Error
}

func (r *leaveRepository) FindLeaveQuotasByEmployeeAndYear(employeeID uuid.UUID, year int) ([]models.LeaveQuota, error) {
	var quotas []models.LeaveQuota
	err := r.db.Preload("LeaveType").Where("employee_id = ? AND year = ?", employeeID, year).Find(&quotas).Error
	return quotas, err
}

func (r *leaveRepository) FindLeaveQuota(employeeID uuid.UUID, leaveTypeID uuid.UUID, year int) (*models.LeaveQuota, error) {
	var q models.LeaveQuota
	err := r.db.Where("employee_id = ? AND leave_type_id = ? AND year = ?", employeeID, leaveTypeID, year).First(&q).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &q, nil
}

func (r *leaveRepository) FindLeaveQuotaByID(id uuid.UUID) (*models.LeaveQuota, error) {
	var q models.LeaveQuota
	err := r.db.Preload("LeaveType").First(&q, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &q, nil
}

func (r *leaveRepository) UpdateLeaveQuota(q *models.LeaveQuota) error {
	return r.db.Save(q).Error
}

func (r *leaveRepository) UpdateLeaveQuotaByID(id uuid.UUID, totalDays float64) error {
	return r.db.Model(&models.LeaveQuota{}).Where("id = ?", id).Update("total_days", totalDays).Error
}

// --- Leave Requests ---

func (r *leaveRepository) CreateLeaveRequest(req *models.LeaveRequest) error {
	return r.db.Create(req).Error
}

func (r *leaveRepository) FindLeaveRequestByID(id uuid.UUID) (*models.LeaveRequest, error) {
	var req models.LeaveRequest
	err := r.db.Preload("Employee").Preload("LeaveType").Preload("Approver").First(&req, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

func (r *leaveRepository) UpdateLeaveRequest(req *models.LeaveRequest) error {
	return r.db.Save(req).Error
}

func (r *leaveRepository) DeleteLeaveRequest(id uuid.UUID) error {
	return r.db.Delete(&models.LeaveRequest{}, id).Error
}

func (r *leaveRepository) FindLeaveRequestsByEmployee(employeeID uuid.UUID, page, limit int) ([]models.LeaveRequest, int64, error) {
	var reqs []models.LeaveRequest
	var total int64
	offset := (page - 1) * limit

	query := r.db.Model(&models.LeaveRequest{}).Where("employee_id = ?", employeeID)
	query.Count(&total)

	err := query.Preload("LeaveType").Preload("Approver").Order("submitted_at desc").Offset(offset).Limit(limit).Find(&reqs).Error
	return reqs, total, err
}

func (r *leaveRepository) FindPendingLeaveRequestsByManager(managerID uuid.UUID, page, limit int) ([]models.LeaveRequest, int64, error) {
	var reqs []models.LeaveRequest
	var total int64
	offset := (page - 1) * limit

	// Manager must be the manager of the requester
	query := r.db.Model(&models.LeaveRequest{}).
		Joins("JOIN employees ON employees.id = leave_requests.employee_id").
		Where("employees.manager_id = ? AND leave_requests.status = ?", managerID, models.LeaveRequestStatusPending)

	query.Count(&total)

	err := query.Preload("Employee").Preload("LeaveType").Order("submitted_at asc").Offset(offset).Limit(limit).Find(&reqs).Error
	return reqs, total, err
}

func (r *leaveRepository) FindAllLeaveRequestsOverview(year, month int) ([]models.LeaveRequest, error) {
	var reqs []models.LeaveRequest
	// Find all requests that overlap with the targeted month/year
	// Date logic depends on DB; we can use simple extract or just load all for the year/month
	query := r.db.Model(&models.LeaveRequest{}).Preload("Employee").Preload("LeaveType").
		Where("EXTRACT(YEAR FROM from_date) = ? AND EXTRACT(MONTH FROM from_date) = ?", year, month)
	err := query.Find(&reqs).Error
	return reqs, err
}

func (r *leaveRepository) FindLeaveRequestByActionToken(token string) (*models.LeaveRequest, error) {
	var req models.LeaveRequest
	err := r.db.Preload("Employee").Preload("LeaveType").Where("action_token = ?", token).First(&req).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}
