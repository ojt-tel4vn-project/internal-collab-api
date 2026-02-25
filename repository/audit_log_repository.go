package repository

import (
	"time"

	auditlog "github.com/ojt-tel4vn-project/internal-collab-api/dtos/audit_log"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type AuditLogRepository interface {
	Create(log *models.AuditLog) error
	FindAll(filter *auditlog.AuditLogFilter) ([]models.AuditLog, int64, error)
}

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditLogRepository) FindAll(filter *auditlog.AuditLogFilter) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.Model(&models.AuditLog{}).Preload("Employee")

	if filter.EmployeeID != nil {
		query = query.Where("employee_id = ?", filter.EmployeeID)
	}
	if filter.Action != nil && *filter.Action != "" {
		query = query.Where("action = ?", *filter.Action)
	}
	if filter.EntityType != nil && *filter.EntityType != "" {
		query = query.Where("entity_type = ?", *filter.EntityType)
	}
	if filter.EntityID != nil {
		query = query.Where("entity_id = ?", filter.EntityID)
	}
	if filter.StartDate != nil && *filter.StartDate != "" {
		startDate, _ := time.Parse("2006-01-02", *filter.StartDate)
		query = query.Where("created_at >= ?", startDate)
	}
	if filter.EndDate != nil && *filter.EndDate != "" {
		endDate, _ := time.Parse("2006-01-02", *filter.EndDate)
		query = query.Where("created_at <= ?", endDate.Add(24*time.Hour))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.IgnorePagination {
		// No limit/offset
		if err := query.Order("created_at desc").Find(&logs).Error; err != nil {
			return nil, 0, err
		}
	} else {
		offset := (filter.Page - 1) * filter.Limit
		if err := query.Order("created_at desc").Offset(offset).Limit(filter.Limit).Find(&logs).Error; err != nil {
			return nil, 0, err
		}
	}

	return logs, total, nil
}
