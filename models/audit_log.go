package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AuditLog struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	// Actor
	EmployeeID *uuid.UUID `gorm:"type:uuid" json:"employee_id"`
	Employee   *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	IPAddress  string     `gorm:"type:inet" json:"ip_address"`
	UserAgent  string     `gorm:"type:text" json:"user_agent"`

	// Action
	Action     string    `gorm:"type:varchar(100);not null" json:"action"`     // e.g. 'create', 'update', 'login'
	EntityType string    `gorm:"type:varchar(50);not null" json:"entity_type"` // e.g. 'employee', 'auth'
	EntityID   uuid.UUID `gorm:"type:uuid;not null" json:"entity_id"`

	// Changes
	OldValues datatypes.JSON `gorm:"type:jsonb" json:"old_values"`
	NewValues datatypes.JSON `gorm:"type:jsonb" json:"new_values"`

	// Metadata
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName overrides the table name used by User to `profiles`
func (AuditLog) TableName() string {
	return "audit_logs"
}
