package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/danielgtaylor/huma/v2"
	auditlog "github.com/ojt-tel4vn-project/internal-collab-api/dtos/audit_log"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

type AuditLogHandler struct {
	service      services.AuditLogService
	jwtService   crypto.JWTService
	employeeRepo repository.EmployeeRepository
}

func NewAuditLogHandler(
	service services.AuditLogService,
	jwtService crypto.JWTService,
	employeeRepo repository.EmployeeRepository,
) *AuditLogHandler {
	return &AuditLogHandler{
		service:      service,
		jwtService:   jwtService,
		employeeRepo: employeeRepo,
	}
}

func (h *AuditLogHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "audit-logs-list",
		Method:      http.MethodGet,
		Path:        "/api/v1/audit-logs",
		Summary:     "List Audit Logs (Admin only)",
		Tags:        []string{"Audit Logs"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.GetLogs)

	huma.Register(api, huma.Operation{
		OperationID: "audit-logs-export",
		Method:      http.MethodGet,
		Path:        "/api/v1/audit-logs/export",
		Summary:     "Export Audit Logs to CSV (Admin only)",
		Tags:        []string{"Audit Logs"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.ExportLogs)
}

type GetLogsRequest struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`

	Page       int    `query:"page" default:"1" doc:"Page number"`
	Limit      int    `query:"limit" default:"20" doc:"Items per page"`
	EmployeeID string `query:"employee_id" doc:"Filter by Employee ID"`
	Action     string `query:"action" doc:"Filter by Action"`
	EntityType string `query:"entity_type" doc:"Filter by Entity Type"`
	EntityID   string `query:"entity_id" doc:"Filter by Entity ID"`
	StartDate  string `query:"start_date" doc:"YYYY-MM-DD"`
	EndDate    string `query:"end_date" doc:"YYYY-MM-DD"`
}

func (h *AuditLogHandler) GetLogs(ctx context.Context, input *GetLogsRequest) (*struct {
	Body auditlog.ListAuditLogResponse
}, error) {
	// 1. Verify Token & Role
	if !strings.HasPrefix(input.Authorization, "Bearer ") {
		return nil, huma.Error401Unauthorized("Invalid authorization format")
	}
	token := strings.TrimPrefix(input.Authorization, "Bearer ")

	if err := h.checkAdmin(token); err != nil {
		return nil, err
	}

	// 2. Prepare Filter
	filter := h.prepareFilter(input)

	// 3. Call Service
	resp, err := h.service.GetLogs(filter)

	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch logs", err)
	}

	return &struct{ Body auditlog.ListAuditLogResponse }{Body: *resp}, nil
}

type ExportLogsResponse struct {
	Body               []byte `doc:"CSV File Content"`
	ContentType        string `header:"Content-Type"`
	ContentDisposition string `header:"Content-Disposition"`
}

func (h *AuditLogHandler) checkAdmin(token string) error {
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		return huma.Error401Unauthorized("Invalid or expired token")
	}

	employee, err := h.employeeRepo.FindByID(claims.UserID)
	if err != nil {
		return huma.Error401Unauthorized("User not found")
	}

	isAdmin := false
	for _, role := range employee.Roles {
		if role.Name == "admin" {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		return huma.Error403Forbidden("Access denied. Admin role required.")
	}
	return nil
}

func (h *AuditLogHandler) prepareFilter(input *GetLogsRequest) *auditlog.AuditLogFilter {
	filter := &auditlog.AuditLogFilter{
		Page:  input.Page,
		Limit: input.Limit,
	}

	if input.EmployeeID != "" {
		if id, err := uuid.Parse(input.EmployeeID); err == nil {
			filter.EmployeeID = &id
		}
	}
	if input.Action != "" {
		filter.Action = &input.Action
	}
	if input.EntityType != "" {
		filter.EntityType = &input.EntityType
	}
	if input.EntityID != "" {
		if id, err := uuid.Parse(input.EntityID); err == nil {
			filter.EntityID = &id
		}
	}
	if input.StartDate != "" {
		filter.StartDate = &input.StartDate
	}
	if input.EndDate != "" {
		filter.EndDate = &input.EndDate
	}
	return filter
}

func (h *AuditLogHandler) ExportLogs(ctx context.Context, input *GetLogsRequest) (*ExportLogsResponse, error) {
	// 1. Verify Token & Role
	if !strings.HasPrefix(input.Authorization, "Bearer ") {
		return nil, huma.Error401Unauthorized("Invalid authorization format")
	}
	token := strings.TrimPrefix(input.Authorization, "Bearer ")

	if err := h.checkAdmin(token); err != nil {
		return nil, err
	}

	// 2. Prepare Filter
	filter := h.prepareFilter(input)

	// 3. Call Service
	data, filename, err := h.service.ExportLogs(filter)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to export logs", err)
	}

	return &ExportLogsResponse{
		Body:               data,
		ContentType:        "text/csv",
		ContentDisposition: "attachment; filename=" + filename,
	}, nil
}
