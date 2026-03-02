package models

import (
	"time"

	"github.com/google/uuid"
)

type LeaveType struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type LeaveQuota struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	EmployeeID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_emp_type_year" json:"employee_id"`
	Employee      *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	LeaveTypeID   uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_emp_type_year" json:"leave_type_id"`
	LeaveType     *LeaveType `gorm:"foreignKey:LeaveTypeID" json:"leave_type,omitempty"`
	Year          int        `gorm:"not null;uniqueIndex:idx_emp_type_year" json:"year"`
	TotalDays     float64    `gorm:"not null" json:"total_days"`
	UsedDays      float64    `gorm:"not null;default:0" json:"used_days"`
	RemainingDays float64    `gorm:"-" json:"remaining_days"` // Calculated field
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// AfterFind hook to calculate RemainingDays
func (l *LeaveQuota) AfterFind() (err error) {
	l.RemainingDays = l.TotalDays - l.UsedDays
	return
}

type LeaveRequestStatus string

const (
	LeaveRequestStatusPending  LeaveRequestStatus = "pending"
	LeaveRequestStatusApproved LeaveRequestStatus = "approved"
	LeaveRequestStatusRejected LeaveRequestStatus = "rejected"
	LeaveRequestStatusCanceled LeaveRequestStatus = "canceled"
)

type LeaveRequest struct {
	ID                 uuid.UUID          `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	EmployeeID         uuid.UUID          `gorm:"type:uuid;not null" json:"employee_id"`
	Employee           *Employee          `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	LeaveTypeID        uuid.UUID          `gorm:"type:uuid;not null" json:"leave_type_id"`
	LeaveType          *LeaveType         `gorm:"foreignKey:LeaveTypeID" json:"leave_type,omitempty"`
	FromDate           time.Time          `gorm:"type:date;not null" json:"from_date"`
	ToDate             time.Time          `gorm:"type:date;not null" json:"to_date"`
	TotalDays          float64            `gorm:"not null" json:"total_days"`
	Reason             string             `gorm:"type:text;not null" json:"reason"`
	ContactDuringLeave string             `gorm:"type:varchar(255)" json:"contact_during_leave"`
	Status             LeaveRequestStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	ApproverID         *uuid.UUID         `gorm:"type:uuid" json:"approver_id"`
	Approver           *Employee          `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
	ApproverComment    string             `gorm:"type:text" json:"approver_comment"`
	ActionToken        *string            `gorm:"type:varchar(255);uniqueIndex" json:"-"` // token for email approval
	SubmittedAt        time.Time          `gorm:"autoCreateTime" json:"submitted_at"`
	UpdatedAt          time.Time          `gorm:"autoUpdateTime" json:"updated_at"`
}
