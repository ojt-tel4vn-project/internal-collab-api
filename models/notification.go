package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationPriority string

const (
	PriorityLow    NotificationPriority = "low"
	PriorityNormal NotificationPriority = "normal"
	PriorityHigh   NotificationPriority = "high"
	PriorityUrgent NotificationPriority = "urgent"
)

type Notification struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	EmployeeID uuid.UUID `gorm:"type:uuid;not null" json:"employee_id"`
	Employee   *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`

	// Notification Content
	Type    string `gorm:"type:varchar(50);not null" json:"type"` // e.g., 'birthday', 'system', 'task'
	Title   string `gorm:"type:varchar(255);not null" json:"title"`
	Message string `gorm:"type:text;not null" json:"message"`

	// Related Entity
	EntityType *string    `gorm:"type:varchar(50)" json:"entity_type,omitempty"`
	EntityID   *uuid.UUID `gorm:"type:uuid" json:"entity_id,omitempty"`
	ActionURL  *string    `gorm:"type:varchar(500)" json:"action_url,omitempty"`

	// Status
	IsRead bool       `gorm:"default:false;index" json:"is_read"`
	ReadAt *time.Time `gorm:"type:timestamp" json:"read_at,omitempty"`

	// Priority
	Priority NotificationPriority `gorm:"type:varchar(20);default:'normal'" json:"priority"`

	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	ExpiresAt *time.Time `gorm:"type:timestamp" json:"expires_at,omitempty"`
}
