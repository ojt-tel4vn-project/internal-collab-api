package notification

import (
	"time"

	"github.com/google/uuid"
)

type CreateNotificationRequest struct {
	EmployeeID uuid.UUID  `json:"employee_id" doc:"Target employee ID" required:"true"`
	Type       string     `json:"type" doc:"Notification type (e.g., birthday)" required:"true"`
	Title      string     `json:"title" doc:"Notification title" required:"true"`
	Message    string     `json:"message" doc:"Notification message" required:"true"`
	Priority   string     `json:"priority" enum:"low,normal,high,urgent" default:"normal"`
	EntityType *string    `json:"entity_type,omitempty" doc:"Related entity type"`
	EntityID   *uuid.UUID `json:"entity_id,omitempty" doc:"Related entity ID"`
	ActionURL  *string    `json:"action_url,omitempty" doc:"Frontend action URL"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" doc:"Optional expiration time"`
}

type NotificationResponse struct {
	ID         uuid.UUID  `json:"id"`
	EmployeeID uuid.UUID  `json:"employee_id"`
	Type       string     `json:"type"`
	Title      string     `json:"title"`
	Message    string     `json:"message"`
	Priority   string     `json:"priority"`
	IsRead     bool       `json:"is_read"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ActionURL  *string    `json:"action_url,omitempty"`
}

type ListNotificationResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int64                  `json:"total"`
	UnreadCount   int64                  `json:"unread_count"`
	Page          int                    `json:"page"`
	Limit         int                    `json:"limit"`
}

type UnreadCountResponse struct {
	Count int64 `json:"count"`
}
