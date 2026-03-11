package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	attendancedto "github.com/ojt-tel4vn-project/internal-collab-api/dtos/attendance"
	"github.com/ojt-tel4vn-project/internal-collab-api/middleware"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

type AttendanceHandler struct {
	service      services.AttendanceService
	jwtService   crypto.JWTService
	employeeRepo repository.EmployeeRepository
}

func NewAttendanceHandler(service services.AttendanceService, jwtService crypto.JWTService, employeeRepo repository.EmployeeRepository) *AttendanceHandler {
	return &AttendanceHandler{
		service:      service,
		jwtService:   jwtService,
		employeeRepo: employeeRepo,
	}
}

func (h *AttendanceHandler) RegisterRoutes(api huma.API) {
	// List Attendances (HR: all; Employee: own)
	huma.Register(api, huma.Operation{
		OperationID: "list-attendances",
		Method:      http.MethodGet,
		Path:        "/api/v1/attendances",
		Summary:     "List Attendances",
		Tags:        []string{"Attendance"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleListAttendances)

	// Upload Attendance CSV (HR only) — raw body upload
	huma.Register(api, huma.Operation{
		OperationID:   "upload-attendance",
		Method:        http.MethodPost,
		Path:          "/api/v1/attendances",
		Summary:       "Upload Attendance (HR only) — CSV body",
		Tags:          []string{"Attendance"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: 201,
	}, h.handleUploadAttendance)

	// Get Attendance by ID
	huma.Register(api, huma.Operation{
		OperationID: "get-attendance",
		Method:      http.MethodGet,
		Path:        "/api/v1/attendances/{id}",
		Summary:     "Get Attendance Details",
		Tags:        []string{"Attendance"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleGetAttendance)

	// Confirm/Dispute own attendance
	huma.Register(api, huma.Operation{
		OperationID: "confirm-attendance",
		Method:      http.MethodPost,
		Path:        "/api/v1/attendances/{id}/confirm",
		Summary:     "Confirm or Dispute Attendance",
		Tags:        []string{"Attendance"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleConfirmAttendance)

	// Add comment/dispute to attendance
	huma.Register(api, huma.Operation{
		OperationID:   "add-attendance-dispute",
		Method:        http.MethodPost,
		Path:          "/api/v1/attendances/{id}/disputes",
		Summary:       "Add Dispute to Attendance",
		Tags:          []string{"Attendance"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: 201,
	}, h.handleAddComment)

	// Review comment (HR only)
	huma.Register(api, huma.Operation{
		OperationID: "review-attendance-dispute",
		Method:      http.MethodPost,
		Path:        "/api/v1/attendances/disputes/{comment_id}/review",
		Summary:     "Review Attendance Dispute (HR only)",
		Tags:        []string{"Attendance"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleReviewComment)

	// Get attendance config (HR only)
	huma.Register(api, huma.Operation{
		OperationID: "get-attendance-config",
		Method:      http.MethodGet,
		Path:        "/api/v1/attendances/config",
		Summary:     "Get Attendance Config (HR only)",
		Tags:        []string{"Attendance"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleGetConfig)

	// Update attendance config (HR only)
	huma.Register(api, huma.Operation{
		OperationID: "update-attendance-config",
		Method:      http.MethodPut,
		Path:        "/api/v1/attendances/config",
		Summary:     "Update Attendance Config (HR only)",
		Tags:        []string{"Attendance"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleUpdateConfig)

	// Get attendance summary (HR only)
	huma.Register(api, huma.Operation{
		OperationID: "get-attendance-summary",
		Method:      http.MethodGet,
		Path:        "/api/v1/attendances/summary",
		Summary:     "Get Attendance Summary (HR only)",
		Tags:        []string{"Attendance"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleGetSummary)
}

// handleListAttendances — HR sees all, Employee sees only self
func (h *AttendanceHandler) handleListAttendances(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
	Month         int    `query:"month" minimum:"1" maximum:"12"`
	Year          int    `query:"year" minimum:"2020" maximum:"2100"`
	Status        string `query:"status" enum:"pending,confirmed,auto_confirmed,"`
	Page          string `query:"page" default:"1"`
	Limit         string `query:"limit" default:"20"`
}) (*attendancedto.AttendanceListResponse, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	page, _ := strconv.Atoi(input.Page)
	limit, _ := strconv.Atoi(input.Limit)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	year := input.Year
	month := input.Month
	if year == 0 {
		year = time.Now().Year()
	}
	if month == 0 {
		month = int(time.Now().Month())
	}

	// HR/Admin can see all; Employee sees only themselves
	var filterEmployeeID *uuid.UUID
	if roleErr := middleware.CheckUserRole(claims.UserID, h.employeeRepo, "hr", "admin"); roleErr != nil {
		// Not HR/Admin → filter to self
		filterEmployeeID = &claims.UserID
	}

	records, total, err := h.service.ListAttendances(filterEmployeeID, month, year, input.Status, page, limit)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list attendances", err)
	}

	res := &attendancedto.AttendanceListResponse{}
	res.Body.Data = records
	res.Body.Total = total
	return res, nil
}

// handleUploadAttendance — parses CSV body (HR only)
func (h *AttendanceHandler) handleUploadAttendance(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
	Month         int    `query:"month" required:"true" minimum:"1" maximum:"12"`
	Year          int    `query:"year" required:"true" minimum:"2020" maximum:"2100"`
	RawBody       []byte
}) (*struct {
	Body struct {
		Data    []attendancedto.AttendanceResponse `json:"data"`
		Message string                             `json:"message"`
	}
}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	if roleErr := middleware.CheckUserRole(claims.UserID, h.employeeRepo, "hr", "admin"); roleErr != nil {
		return nil, huma.Error403Forbidden("Only HR or Admin can upload attendance")
	}

	csvContent := string(input.RawBody)
	if csvContent == "" {
		return nil, huma.Error400BadRequest("CSV body is required")
	}

	results, err := h.service.UploadAttendance(claims.UserID, input.Month, input.Year, csvContent)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error(), err)
	}

	res := &struct {
		Body struct {
			Data    []attendancedto.AttendanceResponse `json:"data"`
			Message string                             `json:"message"`
		}
	}{}
	res.Body.Data = results
	res.Body.Message = "Attendance uploaded successfully"
	return res, nil
}

// handleGetAttendance — get details by ID
func (h *AttendanceHandler) handleGetAttendance(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
	ID            string `path:"id"`
}) (*attendancedto.AttendanceDetailResponse, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid attendance ID")
	}

	a, err := h.service.GetAttendanceByID(id)
	if err != nil {
		return nil, huma.Error404NotFound("Attendance record not found")
	}

	// Non-HR can only view their own
	if roleErr := middleware.CheckUserRole(claims.UserID, h.employeeRepo, "hr", "admin"); roleErr != nil {
		if a.Employee.ID != claims.UserID {
			return nil, huma.Error403Forbidden("Access denied")
		}
	}

	res := &attendancedto.AttendanceDetailResponse{}
	res.Body.Data = *a
	return res, nil
}

// handleConfirmAttendance — employee confirms or auto-confirms own record
func (h *AttendanceHandler) handleConfirmAttendance(ctx context.Context, input *struct {
	Authorization string                                 `header:"Authorization" required:"true"`
	ID            string                                 `path:"id"`
	Body          attendancedto.ConfirmAttendanceRequest `json:"body"`
}) (*struct{}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid attendance ID")
	}

	if err := h.service.ConfirmAttendance(claims.UserID, id, input.Body); err != nil {
		switch err {
		case services.ErrAttendanceNotFound:
			return nil, huma.Error404NotFound("Attendance record not found")
		case services.ErrAlreadyConfirmed:
			return nil, huma.Error409Conflict("Attendance already confirmed")
		case services.ErrUnauthorizedAccess:
			return nil, huma.Error403Forbidden("Access denied")
		default:
			return nil, huma.Error500InternalServerError("Failed to confirm attendance", err)
		}
	}

	return nil, nil
}

// handleAddComment — employee adds dispute/comment
func (h *AttendanceHandler) handleAddComment(ctx context.Context, input *struct {
	Authorization string                          `header:"Authorization" required:"true"`
	ID            string                          `path:"id"`
	Body          attendancedto.AddCommentRequest `json:"body"`
}) (*struct {
	Body struct {
		Data attendancedto.AttendanceCommentResponse `json:"data"`
	}
}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid attendance ID")
	}

	comment, err := h.service.AddComment(claims.UserID, id, input.Body)
	if err != nil {
		switch err {
		case services.ErrAttendanceNotFound:
			return nil, huma.Error404NotFound("Attendance record not found")
		case services.ErrUnauthorizedAccess:
			return nil, huma.Error403Forbidden("Access denied")
		default:
			return nil, huma.Error500InternalServerError("Failed to add comment", err)
		}
	}

	res := &struct {
		Body struct {
			Data attendancedto.AttendanceCommentResponse `json:"data"`
		}
	}{}
	res.Body.Data = *comment
	return res, nil
}

// handleReviewComment — HR reviews a dispute comment
func (h *AttendanceHandler) handleReviewComment(ctx context.Context, input *struct {
	Authorization string                             `header:"Authorization" required:"true"`
	CommentID     string                             `path:"comment_id"`
	Body          attendancedto.ReviewCommentRequest `json:"body"`
}) (*struct{}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	if roleErr := middleware.CheckUserRole(claims.UserID, h.employeeRepo, "hr", "admin"); roleErr != nil {
		return nil, huma.Error403Forbidden("Only HR or Admin can review comments")
	}

	commentID, err := uuid.Parse(input.CommentID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid comment ID")
	}

	if err := h.service.ReviewComment(claims.UserID, commentID, input.Body); err != nil {
		if err == services.ErrCommentNotFound {
			return nil, huma.Error404NotFound("Comment not found")
		}
		return nil, huma.Error500InternalServerError("Failed to review comment", err)
	}

	return nil, nil
}

// handleGetConfig — HR gets attendance config
func (h *AttendanceHandler) handleGetConfig(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
}) (*attendancedto.AttendanceConfigResponse, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	if roleErr := middleware.CheckUserRole(claims.UserID, h.employeeRepo, "hr", "admin"); roleErr != nil {
		return nil, huma.Error403Forbidden("Only HR or Admin can view attendance config")
	}

	cfg, err := h.service.GetAttendanceConfig()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get config", err)
	}

	res := &attendancedto.AttendanceConfigResponse{}
	res.Body.Data = *cfg
	return res, nil
}

// handleUpdateConfig — HR updates attendance config
func (h *AttendanceHandler) handleUpdateConfig(ctx context.Context, input *struct {
	Authorization string                                      `header:"Authorization" required:"true"`
	Body          attendancedto.UpdateAttendanceConfigRequest `json:"body"`
}) (*struct{}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	if roleErr := middleware.CheckUserRole(claims.UserID, h.employeeRepo, "hr", "admin"); roleErr != nil {
		return nil, huma.Error403Forbidden("Only HR or Admin can update attendance config")
	}

	if err := h.service.UpdateAttendanceConfig(input.Body); err != nil {
		return nil, huma.Error500InternalServerError("Failed to update config", err)
	}

	return nil, nil
}

// handleGetSummary — HR gets attendance summary
func (h *AttendanceHandler) handleGetSummary(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
	Month         int    `query:"month" minimum:"1" maximum:"12"`
	Year          int    `query:"year" minimum:"2020" maximum:"2100"`
}) (*attendancedto.AttendanceSummaryResponse, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	if roleErr := middleware.CheckUserRole(claims.UserID, h.employeeRepo, "hr", "admin"); roleErr != nil {
		return nil, huma.Error403Forbidden("Only HR or Admin can view attendance summary")
	}

	month := input.Month
	year := input.Year
	if month == 0 {
		month = int(time.Now().Month())
	}
	if year == 0 {
		year = time.Now().Year()
	}

	summary, err := h.service.GetAttendanceSummary(month, year)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get summary", err)
	}

	res := &attendancedto.AttendanceSummaryResponse{}
	res.Body.Data = *summary
	return res, nil
}
