package models

import (
	"time"

	"github.com/google/uuid"
)

// Role name constants
const (
	RoleAdmin    = "admin"
	RoleManager  = "manager"
	RoleHR       = "hr"
	RoleEmployee = "employee"
)

type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"type:varchar(50);not null;unique" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}
