package models

import (
	"time"

	"github.com/google/uuid"
)

type AttendanceStatus string
type AttendanceCommentStatus string
type DayStatus string

const (
	AttendanceStatusPending       AttendanceStatus = "pending"
	AttendanceStatusConfirmed     AttendanceStatus = "confirmed"
	AttendanceStatusAutoConfirmed AttendanceStatus = "auto_confirmed"

	CommentStatusPending  AttendanceCommentStatus = "pending"
	CommentStatusReviewed AttendanceCommentStatus = "reviewed"
	CommentStatusResolved AttendanceCommentStatus = "resolved"

	DayStatusPresent DayStatus = "present"
	DayStatusAbsent  DayStatus = "absent"
	DayStatusLate    DayStatus = "late"
	DayStatusLeave   DayStatus = "leave"
)

// AttendanceData stores daily status as a map: {"1": "present", "2": "absent", ...}
// GORM stores this as JSONB
type AttendanceData map[string]DayStatus

// Attendance represents a monthly attendance record per employee
type Attendance struct {
	ID               uuid.UUID        `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	EmployeeID       uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:unique_employee_month" json:"employee_id"`
	Employee         *Employee        `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Month            int32            `gorm:"type:integer;not null;uniqueIndex:unique_employee_month" json:"month"`
	Year             int32            `gorm:"type:integer;not null;uniqueIndex:unique_employee_month" json:"year"`
	AttendanceData   AttendanceData   `gorm:"type:jsonb;not null;default:'{}'" json:"attendance_data"`
	TotalDaysPresent int32            `gorm:"type:integer;default:0" json:"total_days_present"`
	TotalDaysAbsent  int32            `gorm:"type:integer;default:0" json:"total_days_absent"`
	TotalDaysLate    int32            `gorm:"type:integer;default:0" json:"total_days_late"`
	Status           AttendanceStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	ConfirmedAt      *time.Time       `gorm:"type:timestamp" json:"confirmed_at,omitempty"`
	UploadedBy       *uuid.UUID       `gorm:"type:uuid" json:"uploaded_by,omitempty"`
	Uploader         *Employee        `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
	UploadedAt       time.Time        `gorm:"autoCreateTime" json:"uploaded_at"`
	CreatedAt        time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
}

// AttendanceComment represents a dispute or clarification comment on an attendance record
type AttendanceComment struct {
	ID           uuid.UUID               `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	AttendanceID uuid.UUID               `gorm:"type:uuid;not null" json:"attendance_id"`
	Attendance   *Attendance             `gorm:"foreignKey:AttendanceID" json:"attendance,omitempty"`
	EmployeeID   uuid.UUID               `gorm:"type:uuid;not null" json:"employee_id"`
	Employee     *Employee               `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Comment      string                  `gorm:"type:text;not null" json:"comment"`
	DayNumber    int32                   `gorm:"type:integer;check:day_number BETWEEN 1 AND 31" json:"day_number"`
	Status       AttendanceCommentStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	ReviewedBy   *uuid.UUID              `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	Reviewer     *Employee               `gorm:"foreignKey:ReviewedBy" json:"reviewer,omitempty"`
	ReviewedAt   *time.Time              `gorm:"type:timestamp" json:"reviewed_at,omitempty"`
	HRResponse   string                  `gorm:"type:text" json:"hr_response,omitempty"`
	CreatedAt    time.Time               `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time               `gorm:"autoUpdateTime" json:"updated_at"`
}
