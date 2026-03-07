package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos/sticker"
	"github.com/ojt-tel4vn-project/internal-collab-api/middleware"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

type StickerHandler struct {
	service      services.StickerService
	jwtService   crypto.JWTService
	employeeRepo repository.EmployeeRepository
}

func NewStickerHandler(service services.StickerService, jwtService crypto.JWTService, employeeRepo repository.EmployeeRepository) *StickerHandler {
	return &StickerHandler{
		service:      service,
		jwtService:   jwtService,
		employeeRepo: employeeRepo,
	}
}

func (h *StickerHandler) RegisterRoutes(api huma.API) {
	// Send Sticker
	huma.Register(api, huma.Operation{
		OperationID: "send-sticker",
		Method:      http.MethodPost,
		Path:        "/api/v1/stickers/send",
		Summary:     "Send a sticker to another employee",
		Description: "Sends a sticker to a colleague. Deducts points from sender.",
		Tags:        []string{"Sticker"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleSendSticker)

	// Get My Point Balance
	huma.Register(api, huma.Operation{
		OperationID: "get-point-balance",
		Method:      http.MethodGet,
		Path:        "/api/v1/stickers/balance",
		Summary:     "Get current user's point balance",
		Description: "Returns the current point balance for the authenticated user.",
		Tags:        []string{"Sticker"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleGetPointBalance)

	// Get Leaderboard
	huma.Register(api, huma.Operation{
		OperationID: "get-sticker-leaderboard",
		Method:      http.MethodGet,
		Path:        "/api/v1/stickers/leaderboard",
		Summary:     "Get sticker leaderboard",
		Description: "Returns top employees who received the most stickers. Supports filtering by time range and department.",
		Tags:        []string{"Sticker"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleGetLeaderboard)

	// Update Global Point Config (HR/Admin only)
	huma.Register(api, huma.Operation{
		OperationID: "update-point-config",
		Method:      http.MethodPut,
		Path:        "/api/v1/stickers/config",
		Summary:     "Update global point configuration (HR/Admin only)",
		Description: "Updates the yearly point allocation and reset date. Only HR or Admin can access this.",
		Tags:        []string{"Sticker"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleUpdateGlobalConfig)
}

// handleSendSticker handles POST /api/v1/stickers/send
func (h *StickerHandler) handleSendSticker(ctx context.Context, input *struct {
	Authorization string                     `header:"Authorization" required:"true"`
	Body          sticker.SendStickerRequest `json:"body"`
}) (*sticker.SendStickerResponse, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	senderID := claims.UserID

	err = h.service.SendSticker(senderID, input.Body.ReceiverID, input.Body.StickerTypeID, input.Body.Message)
	if err != nil {
		switch err.Error() {
		case "cannot send sticker to yourself":
			return nil, huma.Error400BadRequest("Cannot send sticker to yourself")
		case "sticker type not found":
			return nil, huma.Error404NotFound("Sticker type not found")
		case "sender point balance not found":
			return nil, huma.Error404NotFound("Point balance not found. Please contact HR.")
		case "not enough points to send sticker":
			return nil, huma.Error400BadRequest("Not enough points to send sticker")
		default:
			return nil, huma.Error500InternalServerError("Failed to send sticker", err)
		}
	}

	res := &sticker.SendStickerResponse{}
	res.Body.Success = true
	res.Body.Message = "Sticker sent successfully"
	return res, nil
}

// handleGetPointBalance handles GET /api/v1/stickers/balance
func (h *StickerHandler) handleGetPointBalance(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
}) (*sticker.PointBalanceAPIResponse, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	balance, err := h.service.GetPointBalance(claims.UserID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch point balance", err)
	}

	res := &sticker.PointBalanceAPIResponse{}
	res.Body.Data = sticker.PointBalanceResponse{
		EmployeeID:    balance.EmployeeID,
		Year:          balance.Year,
		InitialPoints: balance.InitialPoints,
		CurrentPoints: balance.CurrentPoints,
	}
	return res, nil
}

// handleGetLeaderboard handles GET /api/v1/stickers/leaderboard
func (h *StickerHandler) handleGetLeaderboard(ctx context.Context, input *struct {
	Authorization string    `header:"Authorization" required:"true"`
	Limit         int       `query:"limit" minimum:"1" maximum:"100" default:"10"`
	StartDate     string    `query:"start_date"`
	EndDate       string    `query:"end_date"`
	DepartmentID  uuid.UUID `query:"department_id"`
}) (*sticker.LeaderboardResponse, error) {
	_, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	filter := repository.LeaderboardFilter{
		Limit: input.Limit,
	}

	// Parse date strings if provided
	if input.StartDate != "" {
		t, err := parseTime(input.StartDate)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid start_date format. Use RFC3339 format.")
		}
		filter.StartDate = &t
	}
	if input.EndDate != "" {
		t, err := parseTime(input.EndDate)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid end_date format. Use RFC3339 format.")
		}
		filter.EndDate = &t
	}
	if input.DepartmentID != uuid.Nil {
		deptID := input.DepartmentID
		filter.DepartmentID = &deptID
	}

	results, err := h.service.GetLeaderboard(filter)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch leaderboard", err)
	}

	res := &sticker.LeaderboardResponse{}
	res.Body.Data = results
	return res, nil
}

// handleUpdateGlobalConfig handles PUT /api/v1/stickers/config
func (h *StickerHandler) handleUpdateGlobalConfig(ctx context.Context, input *struct {
	Authorization string                            `header:"Authorization" required:"true"`
	Body          sticker.UpdateGlobalConfigRequest `json:"body"`
}) (*sticker.UpdateConfigResponse, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	// HR/Admin only
	if roleErr := middleware.CheckUserRole(claims.UserID, h.employeeRepo, "hr", "admin"); roleErr != nil {
		return nil, huma.Error403Forbidden("Only HR or Admin can update point configuration")
	}

	err = h.service.UpdateGlobalConfig(input.Body.YearlyPoints, input.Body.ResetMonth, input.Body.ResetDay)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to update configuration", err)
	}

	res := &sticker.UpdateConfigResponse{}
	res.Body.Success = true
	res.Body.Message = "Point configuration updated successfully"
	return res, nil
}

// parseTime parses a time string in RFC3339 format
func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
