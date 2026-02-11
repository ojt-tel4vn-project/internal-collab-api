package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	BaseRepository[models.Notification]
	Create(notification *models.Notification) error
	GetNotifications(employeeID uuid.UUID, page, limit int) ([]models.Notification, int64, error)
	GetUnreadCount(employeeID uuid.UUID) (int64, error)
	MarkAsRead(id, employeeID uuid.UUID) error
	MarkAllAsRead(employeeID uuid.UUID) error
}

type notificationRepository struct {
	BaseRepository[models.Notification]
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{
		BaseRepository: NewBaseRepository[models.Notification](db),
		db:             db,
	}
}

func (r *notificationRepository) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetNotifications(employeeID uuid.UUID, page, limit int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	// Filter by employee and active notifications (not expired)
	query := r.db.Model(&models.Notification{}).
		Where("employee_id = ?", employeeID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now())

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (r *notificationRepository) GetUnreadCount(employeeID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("employee_id = ? AND is_read = ?", employeeID, false).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Count(&count).Error
	return count, err
}

func (r *notificationRepository) MarkAsRead(id, employeeID uuid.UUID) error {
	now := time.Now()
	// Update using ID and EmployeeID to ensure ownership
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND employee_id = ?", id, employeeID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

func (r *notificationRepository) MarkAllAsRead(employeeID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.Notification{}).
		Where("employee_id = ? AND is_read = ?", employeeID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}
