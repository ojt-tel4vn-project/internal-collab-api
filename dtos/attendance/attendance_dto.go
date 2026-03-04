package attendance

import (
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
)

// ---- Request DTOs ----

// ConfirmAttendanceRequest is the body for confirming/disputing an attendance record
type ConfirmAttendanceRequest struct {
	Status  string `json:"status" enum:"confirmed,auto_confirmed" doc:"Status to set: confirmed or auto_confirmed"`
	Comment string `json:"comment,omitempty" doc:"Optional comment"`
}

// AddCommentRequest is the body for adding a dispute comment
type AddCommentRequest struct {
	Comment   string `json:"comment" required:"true" minLength:"1"`
	DayNumber int    `json:"day_number" required:"true" minimum:"1" maximum:"31"`
}

// ReviewCommentRequest is the body for HR to review a dispute comment
type ReviewCommentRequest struct {
	HRResponse string `json:"hr_response" required:"true" minLength:"1"`
	Status     string `json:"status" required:"true" enum:"reviewed,resolved"`
}

// UpdateAttendanceConfigRequest is the body for updating attendance config
type UpdateAttendanceConfigRequest struct {
	ConfirmationDeadlineDays   int  `json:"confirmation_deadline_days" minimum:"1" maximum:"30"`
	AutoConfirmEnabled         bool `json:"auto_confirm_enabled"`
	ReminderBeforeDeadlineDays int  `json:"reminder_before_deadline_days" minimum:"0" maximum:"10"`
}

// ---- Response DTOs ----

type EmployeeRef struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	Email    string    `json:"email,omitempty"`
	Position string    `json:"position,omitempty"`
}

type AttendanceResponse struct {
	ID               uuid.UUID               `json:"id"`
	Employee         EmployeeRef             `json:"employee"`
	Month            int32                   `json:"month"`
	Year             int32                   `json:"year"`
	AttendanceData   models.AttendanceData   `json:"attendance_data"`
	TotalDaysPresent int32                   `json:"total_days_present"`
	TotalDaysAbsent  int32                   `json:"total_days_absent"`
	TotalDaysLate    int32                   `json:"total_days_late"`
	Status           models.AttendanceStatus `json:"status"`
	ConfirmedAt      *time.Time              `json:"confirmed_at,omitempty"`
	UploadedAt       time.Time               `json:"uploaded_at"`
}

type AttendanceListResponse struct {
	Body struct {
		Data  []AttendanceResponse `json:"data"`
		Total int64                `json:"total"`
	}
}

type AttendanceDetailResponse struct {
	Body struct {
		Data AttendanceResponse `json:"data"`
	}
}

type AttendanceCommentResponse struct {
	ID           uuid.UUID                      `json:"id"`
	AttendanceID uuid.UUID                      `json:"attendance_id"`
	Employee     EmployeeRef                    `json:"employee"`
	Comment      string                         `json:"comment"`
	DayNumber    int32                          `json:"day_number"`
	Status       models.AttendanceCommentStatus `json:"status"`
	HRResponse   string                         `json:"hr_response,omitempty"`
	ReviewedAt   *time.Time                     `json:"reviewed_at,omitempty"`
	CreatedAt    time.Time                      `json:"created_at"`
}

type AttendanceSummaryResponse struct {
	Body struct {
		Data AttendanceSummary `json:"data"`
	}
}

type AttendanceSummary struct {
	TotalEmployees int `json:"total_employees"`
	Confirmed      int `json:"confirmed"`
	Pending        int `json:"pending"`
	AutoConfirmed  int `json:"auto_confirmed"`
}

type AttendanceConfig struct {
	ConfirmationDeadlineDays   int  `json:"confirmation_deadline_days"`
	AutoConfirmEnabled         bool `json:"auto_confirm_enabled"`
	ReminderBeforeDeadlineDays int  `json:"reminder_before_deadline_days"`
}

type AttendanceConfigResponse struct {
	Body struct {
		Data AttendanceConfig `json:"data"`
	}
}
