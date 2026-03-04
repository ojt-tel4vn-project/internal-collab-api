package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos/leave"
	"github.com/ojt-tel4vn-project/internal-collab-api/middleware"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

type LeaveHandler struct {
	service      services.LeaveService
	jwtService   crypto.JWTService
	employeeRepo repository.EmployeeRepository
}

func NewLeaveHandler(service services.LeaveService, jwtService crypto.JWTService, employeeRepo repository.EmployeeRepository) *LeaveHandler {
	return &LeaveHandler{
		service:      service,
		jwtService:   jwtService,
		employeeRepo: employeeRepo,
	}
}

func (h *LeaveHandler) RegisterRoutes(api huma.API) {
	// 4. Leave Management

	// List Leave Types
	huma.Register(api, huma.Operation{
		OperationID: "get-leave-types",
		Method:      http.MethodGet,
		Path:        "/api/v1/leave-types",
		Summary:     "List Leave Types",
		Tags:        []string{"Leave"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleGetLeaveTypes)

	// Get Leave Quotas
	huma.Register(api, huma.Operation{
		OperationID: "get-leave-quotas",
		Method:      http.MethodGet,
		Path:        "/api/v1/leave-quotas",
		Summary:     "Get Leave Quotas",
		Tags:        []string{"Leave"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleGetLeaveQuotas)

	// Update Leave Quota (HR only)
	huma.Register(api, huma.Operation{
		OperationID: "update-leave-quota",
		Method:      http.MethodPut,
		Path:        "/api/v1/leave-quotas/{id}",
		Summary:     "Update Leave Quota (HR only)",
		Tags:        []string{"Leave"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleUpdateLeaveQuota)

	// Get My Leave Requests (for employee)
	huma.Register(api, huma.Operation{
		OperationID: "get-my-leave-requests",
		Method:      http.MethodGet,
		Path:        "/api/v1/leave-requests",
		Summary:     "Get My Leave Requests",
		Tags:        []string{"Leave"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleGetMyLeaveRequests)

	// Create Leave Request
	huma.Register(api, huma.Operation{
		OperationID: "create-leave-request",
		Method:      http.MethodPost,
		Path:        "/api/v1/leave-requests",
		Summary:     "Create Leave Request",
		Tags:        []string{"Leave"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleCreateLeaveRequest)

	// Approve/Reject Leave Request (Manager only)
	huma.Register(api, huma.Operation{
		OperationID: "approve-leave-request",
		Method:      http.MethodPost,
		Path:        "/api/v1/leave-requests/{id}/approve",
		Summary:     "Approve/Reject Leave Request (Manager only)",
		Tags:        []string{"Leave"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleApproveLeaveRequest)

	// Cancel Leave Request
	huma.Register(api, huma.Operation{
		OperationID: "cancel-leave-request",
		Method:      http.MethodDelete,
		Path:        "/api/v1/leave-requests/{id}",
		Summary:     "Cancel Leave Request",
		Tags:        []string{"Leave"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleCancelLeaveRequest)

	// Approve Leave via Email Link
	huma.Register(api, huma.Operation{
		OperationID: "email-action-leave-request",
		Method:      http.MethodPost,
		Path:        "/api/v1/leave-requests/{id}/email-action",
		Summary:     "Approve Leave via Email Link",
		Tags:        []string{"Leave"},
	}, h.handleEmailActionLeaveRequest)

	// Get Leave Requests for Manager
	huma.Register(api, huma.Operation{
		OperationID: "get-manager-leave-requests",
		Method:      http.MethodGet,
		Path:        "/api/v1/leave-requests/pending-approval",
		Summary:     "Get Leave Requests for Manager",
		Tags:        []string{"Leave"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleGetPendingLeaveRequests)

	// Get Company Leave Overview (HR only)
	huma.Register(api, huma.Operation{
		OperationID: "get-leave-overview",
		Method:      http.MethodGet,
		Path:        "/api/v1/leave-requests/overview",
		Summary:     "Get Company Leave Overview (HR only)",
		Tags:        []string{"Leave"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleGetLeaveOverview)
}

func (h *LeaveHandler) handleGetLeaveTypes(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
}) (*leave.LeaveTypeListResponse, error) {
	_, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	types, err := h.service.GetLeaveTypes()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch leave types", err)
	}

	res := &leave.LeaveTypeListResponse{}
	res.Body.Data = types
	return res, nil
}

func (h *LeaveHandler) handleGetLeaveQuotas(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
	Year          int    `query:"year" minimum:"2000" maximum:"2100"`
}) (*struct {
	Body struct {
		Data []leave.LeaveQuotaResponse `json:"data"`
	}
}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}
	employeeID := claims.UserID

	year := input.Year
	if year == 0 {
		year = time.Now().Year()
	}

	quotas, err := h.service.GetLeaveQuotas(employeeID, year)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch leave quotas", err)
	}

	res := &struct {
		Body struct {
			Data []leave.LeaveQuotaResponse `json:"data"`
		}
	}{}
	res.Body.Data = quotas
	return res, nil
}

func (h *LeaveHandler) handleUpdateLeaveQuota(ctx context.Context, input *struct {
	Authorization string                        `header:"Authorization" required:"true"`
	ID            string                        `path:"id"`
	Body          leave.UpdateLeaveQuotaRequest `json:"body"`
}) (*struct{}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	// HR-only: check role
	if roleErr := middleware.CheckUserRole(claims.UserID, h.employeeRepo, "hr", "admin"); roleErr != nil {
		return nil, huma.Error403Forbidden("Only HR or Admin can update leave quotas")
	}

	quotaID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid quota ID", err)
	}

	err = h.service.UpdateLeaveQuota(quotaID, input.Body)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error(), err)
	}

	return nil, nil
}

func (h *LeaveHandler) handleGetMyLeaveRequests(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
	Page          string `query:"page" default:"1"`
	Limit         string `query:"limit" default:"20"`
}) (*struct {
	Body struct {
		Data  []leave.LeaveRequestResponse `json:"data"`
		Total int64                        `json:"total"`
	}
}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}
	employeeID := claims.UserID

	page, _ := strconv.Atoi(input.Page)
	limit, _ := strconv.Atoi(input.Limit)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	reqs, total, err := h.service.GetMyLeaveRequests(employeeID, page, limit)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get leave requests", err)
	}

	res := &struct {
		Body struct {
			Data  []leave.LeaveRequestResponse `json:"data"`
			Total int64                        `json:"total"`
		}
	}{}
	res.Body.Data = reqs
	res.Body.Total = total
	return res, nil
}

func (h *LeaveHandler) handleCreateLeaveRequest(ctx context.Context, input *struct {
	Authorization string                   `header:"Authorization" required:"true"`
	Body          leave.CreateLeaveRequest `json:"body"`
}) (*leave.CreateLeaveResponse, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}
	employeeID := claims.UserID

	resp, warning, err := h.service.CreateLeaveRequest(employeeID, input.Body)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error(), err)
	}

	res := &leave.CreateLeaveResponse{}
	res.Body.Data = *resp
	if warning != nil {
		res.Body.Warning = warning
	}
	return res, nil
}

func (h *LeaveHandler) handleApproveLeaveRequest(ctx context.Context, input *struct {
	Authorization string                    `header:"Authorization" required:"true"`
	ID            string                    `path:"id"`
	Body          leave.ApproveLeaveRequest `json:"body"`
}) (*struct{}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}
	managerID := claims.UserID

	reqID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid request ID", err)
	}

	err = h.service.ApproveLeaveRequest(managerID, reqID, input.Body.Action, input.Body.Comment)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error(), err)
	}

	return nil, nil
}

func (h *LeaveHandler) handleCancelLeaveRequest(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
	ID            string `path:"id"`
}) (*struct{}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}
	employeeID := claims.UserID

	reqID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid request ID", err)
	}

	err = h.service.CancelLeaveRequest(employeeID, reqID)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error(), err)
	}

	return nil, nil
}

func (h *LeaveHandler) handleEmailActionLeaveRequest(ctx context.Context, input *struct {
	ID   string                        `path:"id"`
	Body leave.EmailActionLeaveRequest `json:"body"`
}) (*struct{}, error) {
	err := h.service.EmailActionLeaveRequest(input.Body.Token, input.Body.Action)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error(), err)
	}

	return nil, nil
}

func (h *LeaveHandler) handleGetPendingLeaveRequests(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
	Page          string `query:"page" default:"1"`
	Limit         string `query:"limit" default:"20"`
}) (*leave.LeaveRequestListResponse, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}
	managerID := claims.UserID

	page, _ := strconv.Atoi(input.Page)
	limit, _ := strconv.Atoi(input.Limit)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	reqs, _, err := h.service.GetPendingLeaveRequests(managerID, page, limit)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get pending leave requests", err)
	}

	res := &leave.LeaveRequestListResponse{}
	res.Body.Data = reqs
	return res, nil
}

func (h *LeaveHandler) handleGetLeaveOverview(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
	Year          int    `query:"year"`
	Month         int    `query:"month"`
}) (*leave.LeaveOverviewResponse, error) {
	_, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}
	year := input.Year
	month := input.Month
	if year == 0 {
		year = time.Now().Year()
	}
	if month == 0 {
		month = int(time.Now().Month())
	}

	overview, err := h.service.GetCompanyLeaveOverview(year, month)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get overview", err)
	}

	res := &leave.LeaveOverviewResponse{}
	res.Body.Data = *overview
	return res, nil
}
