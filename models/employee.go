package models

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusInActive Status = "inactive"
	StatusPending  Status = "pending"
)

type Employee struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	EmployeeCode string    `gorm:"type:varchar(20);not null;unique" json:"employee_code"`
	Email        string    `gorm:"type:varchar(255);not null;unique" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`

	// Personal info
	FirstName   string    `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName    string    `gorm:"type:varchar(100);not null" json:"last_name"`
	FullName    string    `gorm:"type:varchar(255);->;generated" json:"full_name"`
	DateOfBirth time.Time `gorm:"type:date;not null" json:"date_of_birth"`
	Phone       string    `gorm:"type:varchar(20)" json:"phone"`
	Address     string    `gorm:"type:text" json:"address"`
	AvatarUrl   string    `gorm:"type:varchar(500);column:avatar_url" json:"avatar_url"`

	// Work info
	DepartmentID *uuid.UUID  `gorm:"type:uuid" json:"department_id"`
	Department   *Department `gorm:"foreignKey:DepartmentID;references:ID" json:"department,omitempty"`
	Position     string      `gorm:"type:varchar(100);not null" json:"position"`
	ManagerID    *uuid.UUID  `gorm:"type:uuid" json:"manager_id"`
	Manager      *Employee   `gorm:"foreignKey:ManagerID;references:ID" json:"manager,omitempty"`
	JoinDate     time.Time   `gorm:"type:date;not null" json:"join_date"`
	LeaveDate    *time.Time  `gorm:"type:date" json:"leave_date"`

	// System
	Status      Status     `gorm:"type:varchar(20);default:'active'" json:"status"`
	LastLoginAt *time.Time `gorm:"type:timestamp" json:"last_login_at"`

	// Password Reset
	PasswordResetToken     *string    `gorm:"type:varchar(255)" json:"-"`
	PasswordResetExpiresAt *time.Time `gorm:"type:timestamp" json:"-"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Roles []Role `gorm:"many2many:employee_roles;" json:"roles,omitempty"`
}
