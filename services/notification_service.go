package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"go.uber.org/zap"
)

type NotificationService interface {
	SendNotification(employeeID uuid.UUID, nType, title, message string, entityType *string, entityID *uuid.UUID, priority models.NotificationPriority) error
	GetNotifications(employeeID uuid.UUID, page, limit int) ([]models.Notification, int64, error)
	GetUnreadCount(employeeID uuid.UUID) (int64, error)
	MarkAsRead(id, employeeID uuid.UUID) error
	MarkAllAsRead(employeeID uuid.UUID) error
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) SendNotification(employeeID uuid.UUID, nType, title, message string, entityType *string, entityID *uuid.UUID, priority models.NotificationPriority) error {
	notification := &models.Notification{
		EmployeeID: employeeID,
		Type:       nType,
		Title:      title,
		Message:    message,
		EntityType: entityType,
		EntityID:   entityID,
		Priority:   priority,
		CreatedAt:  time.Now(),
		IsRead:     false,
	}

	// Default expiration: 30 days
	expiry := time.Now().Add(30 * 24 * time.Hour)
	notification.ExpiresAt = &expiry

	if err := s.repo.Create(notification); err != nil {
		logger.Error("Failed to create notification", zap.Error(err), zap.String("email_id", employeeID.String()))
		return err
	}

	// TODO: Trigger WebSocket or Push Notification here if implemented
	return nil
}

func (s *notificationService) GetNotifications(employeeID uuid.UUID, page, limit int) ([]models.Notification, int64, error) {
	return s.repo.GetNotifications(employeeID, page, limit)
}

func (s *notificationService) GetUnreadCount(employeeID uuid.UUID) (int64, error) {
	return s.repo.GetUnreadCount(employeeID)
}

func (s *notificationService) MarkAsRead(id, employeeID uuid.UUID) error {
	return s.repo.MarkAsRead(id, employeeID)
}

func (s *notificationService) MarkAllAsRead(employeeID uuid.UUID) error {
	return s.repo.MarkAllAsRead(employeeID)
}
