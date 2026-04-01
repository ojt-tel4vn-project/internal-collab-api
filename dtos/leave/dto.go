package leave

import (
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
)

// ========= Types =========

type EmployeeResponse struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Position  string    `json:"position"`
	AvatarUrl string    `json:"avatar_url"`
}

type LeaveTypeResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TotalDays   float64   `json:"total_days"`
}

type LeaveTypeListResponse struct {
	Body struct {
		Data []LeaveTypeResponse `json:"data"`
	}
}

// ========= Quotas =========

type LeaveQuotaResponse struct {
	ID            uuid.UUID `json:"id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	LeaveTypeID   uuid.UUID `json:"leave_type_id"`
	LeaveTypeName string    `json:"leave_type_name"`
	Year          int       `json:"year"`
	TotalDays     float64   `json:"total_days"`
	UsedDays      float64   `json:"used_days"`
	RemainingDays float64   `json:"remaining_days"`
}

type UpdateLeaveQuotaRequest struct {
	TotalDays float64 `json:"total_days" validate:"required,min=0"`
	Reason    string  `json:"reason"`
}

// ========= Requests =========

type CreateLeaveRequest struct {
	LeaveTypeID        uuid.UUID `json:"leave_type_id" validate:"required"`
	FromDate           string    `json:"from_date" validate:"required"` // Format: YYYY-MM-DD
	ToDate             string    `json:"to_date" validate:"required"`   // Format: YYYY-MM-DD
	Reason             string    `json:"reason" validate:"required"`
	ContactDuringLeave string    `json:"contact_during_leave" validate:"required" pattern:"^(0[0-9]{9}|[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,})$" doc:"Valid 10-digit phone number or email address" example:"0912345678"`
}

type ApproveLeaveRequest struct {
	Action  string `json:"action" validate:"required,oneof=approve reject"`
	Comment string `json:"comment"`
}

type EmailActionLeaveRequest struct {
	Token  string `json:"token" validate:"required"`
	Action string `json:"action" validate:"required,oneof=approve reject"`
}

type LeaveRequestWarning struct {
	Type          string  `json:"type"`
	Message       string  `json:"message"`
	RemainingDays float64 `json:"remaining_days"`
	RequestedDays float64 `json:"requested_days"`
}

type LeaveRequestResponse struct {
	ID                 uuid.UUID                 `json:"id"`
	Employee           EmployeeResponse          `json:"employee"`
	LeaveType          LeaveTypeResponse         `json:"leave_type"`
	FromDate           string                    `json:"from_date"`
	ToDate             string                    `json:"to_date"`
	TotalDays          float64                   `json:"total_days"`
	Reason             string                    `json:"reason"`
	ContactDuringLeave string                    `json:"contact_during_leave"`
	Status             models.LeaveRequestStatus `json:"status"`
	IsOverdue          bool                      `json:"is_overdue"`
	Approver           *EmployeeResponse         `json:"approver,omitempty"`
	ApproverComment    string                    `json:"approver_comment,omitempty"`
	SubmittedAt        time.Time                 `json:"submitted_at"`
}

type CreateLeaveResponse struct {
	Body struct {
		Data    LeaveRequestResponse `json:"data"`
		Warning *LeaveRequestWarning `json:"warning,omitempty"`
	}
}

type LeaveRequestListResponse struct {
	Body struct {
		Data []LeaveRequestResponse `json:"data"`
	}
}

type LeaveOverview struct {
	TotalRequests         int `json:"total_requests"`
	Pending               int `json:"pending"`
	Approved              int `json:"approved"`
	Rejected              int `json:"rejected"`
	EmployeesOnLeaveToday int `json:"employees_on_leave_today"`
	UpcomingLeaves        []struct {
		Employee string `json:"employee"`
		FromDate string `json:"from_date"`
		ToDate   string `json:"to_date"`
	} `json:"upcoming_leaves"`
}

type LeaveOverviewResponse struct {
	Body struct {
		Data LeaveOverview `json:"data"`
	}
}
