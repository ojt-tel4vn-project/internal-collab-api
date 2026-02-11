package auditlog

import (
	"time"

	"github.com/google/uuid"
)

type AuditLogFilter struct {
	EmployeeID       *uuid.UUID `json:"employee_id,omitempty" doc:"Filter by employee ID"`
	Action           *string    `json:"action,omitempty" doc:"Filter by action (e.g., create, update)"`
	EntityType       *string    `json:"entity_type,omitempty" doc:"Filter by entity type"`
	EntityID         *uuid.UUID `json:"entity_id,omitempty" doc:"Filter by entity ID"`
	StartDate        *string    `json:"start_date,omitempty" doc:"Start date (YYYY-MM-DD)"`
	EndDate          *string    `json:"end_date,omitempty" doc:"End date (YYYY-MM-DD)"`
	Page             int        `json:"page,omitempty" default:"1" doc:"Page number"`
	Limit            int        `json:"limit,omitempty" default:"20" doc:"Items per page"`
	IgnorePagination bool       `json:"ignore_pagination,omitempty" doc:"If true, returns all matching records"`
}

type AuditLogResponse struct {
	ID           uuid.UUID              `json:"id"`
	EmployeeID   *uuid.UUID             `json:"employee_id"`
	EmployeeName string                 `json:"employee_name,omitempty"`
	Action       string                 `json:"action"`
	EntityType   string                 `json:"entity_type"`
	EntityID     uuid.UUID              `json:"entity_id"`
	OldValues    map[string]interface{} `json:"old_values,omitempty"`
	NewValues    map[string]interface{} `json:"new_values,omitempty"`
	Description  string                 `json:"description,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

type ListAuditLogResponse struct {
	Logs  []AuditLogResponse `json:"logs"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
}
