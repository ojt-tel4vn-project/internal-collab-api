package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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

	//Get all stickers
	huma.Register(api, huma.Operation{
		OperationID: "list-sticker-types",
		Method:      http.MethodGet,
		Path:        "/api/v1/stickers/types",
		Summary:     "List sticker types",
		Description: "Returns a list of all available sticker types with their details.",
		Tags:        []string{"Sticker"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleGetAllStickerTypes)

	//Create Sticker
	huma.Register(api, huma.Operation{
		OperationID: "create-sticker",
		Method:      http.MethodPost,
		Path:        "/api/v1/hr/stickers",
		Summary:     "Create a new sticker type (HR)",
		Description: "Creates a new sticker type with specified attributes. Only HR can access this.",
		Tags:        []string{"Sticker", "HR"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		RequestBody: &huma.RequestBody{
			Content: map[string]*huma.MediaType{
				"multipart/form-data": {
					Schema: &huma.Schema{
						Type: "object",
						Properties: map[string]*huma.Schema{
							"name":          {Type: "string", Description: "Sticker Name"},
							"description":   {Type: "string", Description: "Sticker Description"},
							"point_cost":    {Type: "integer", Description: "Points needed to send"},
							"category":      {Type: "string", Description: "Category (e.g. Work, Fun)"},
							"display_order": {Type: "integer", Description: "Order in list"},
							"icon":          {Type: "string", Format: "binary", Description: "Sticker icon image"},
						},
						Required: []string{"name", "point_cost", "icon"},
					},
				},
			},
		},
	}, h.handleCreateSticker)

	// Update Global Point Config (HR only)
	huma.Register(api, huma.Operation{
		OperationID: "update-point-config",
		Method:      http.MethodPut,
		Path:        "/api/v1/hr/stickers/config",
		Summary:     "Update global point configuration (HR)",
		Description: "Updates the yearly point allocation and reset date. Only HR can access this.",
		Tags:        []string{"Sticker", "HR"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleUpdateGlobalConfig)
}

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
		switch err {
		case services.ErrSendToYourself:
			return nil, huma.Error400BadRequest("Cannot send sticker to yourself")
		case services.ErrStickerNotFound:
			return nil, huma.Error404NotFound("Sticker type not found")
		case services.ErrPointNotFound:
			return nil, huma.Error404NotFound("Point balance not found. Please contact HR.")
		case services.ErrNotEnoughPoints:
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

	// Parse date strings if providedT00:00:00Z
	if input.StartDate != "" {
		t, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid start_date format (YYYY-MM-DD)")
		}
		filter.StartDate = &t
	}
	if input.EndDate != "" {
		t, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid end_date format (YYYY-MM-DD)")
		}
		filter.EndDate = &t
	}

	if filter.StartDate != nil && filter.EndDate != nil {
		if filter.StartDate.After(*filter.EndDate) {
			return nil, huma.Error400BadRequest("start_date cannot be after end_date")
		}
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

func (h *StickerHandler) handleGetAllStickerTypes(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
}) (*sticker.GetStickerTypesResponse, error) {
	_, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	stickers, err := h.service.ListStickerTypes()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch sticker types", err)
	}

	res := &sticker.GetStickerTypesResponse{}
	res.Body.Data = stickers
	return res, nil
}

func (h *StickerHandler) handleCreateSticker(ctx context.Context, input *struct {
	Authorization string         `header:"Authorization" required:"true"`
	RawBody       multipart.Form `contentType:"multipart/form-data"`
}) (*sticker.CreateStickerResponse, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing authentication")
	}

	// HR/Admin only
	if roleErr := middleware.CheckUserRole(claims.UserID, h.employeeRepo, "hr", "admin"); roleErr != nil {
		return nil, huma.Error403Forbidden("Only HR or Admin can create stickers")
	}

	files, ok := input.RawBody.File["icon"]
	if !ok || len(files) == 0 {
		return nil, huma.Error400BadRequest("Icon file is required")
	}
	fileHeader := files[0]
	name := ""
	if names, ok := input.RawBody.Value["name"]; ok && len(names) > 0 {
		name = names[0]
	}

	description := ""
	if descs, ok := input.RawBody.Value["description"]; ok && len(descs) > 0 {
		description = descs[0]
	}

	pointCost := 0
	if costs, ok := input.RawBody.Value["point_cost"]; ok && len(costs) > 0 {
		fmt.Sscanf(costs[0], "%d", &pointCost)
	}

	category := ""
	if cats, ok := input.RawBody.Value["category"]; ok && len(cats) > 0 {
		category = cats[0]
	}

	displayOrder := 0
	if orders, ok := input.RawBody.Value["display_order"]; ok && len(orders) > 0 {
		fmt.Sscanf(orders[0], "%d", &displayOrder)
	}
	if fileHeader.Size > 1024*1024 {
		return nil, huma.Error400BadRequest("Icon must be less than 1MB")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to open icon file", err)
	}
	defer file.Close()
	ext := filepath.Ext(fileHeader.Filename)
	filename := uuid.New().String() + ext

	iconURL, err := UploadSticker(file, filename)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to upload icon", err)
	}

	req := sticker.CreateStickerRequest{
		Name:         name,
		Description:  description,
		PointCost:    pointCost,
		Category:     category,
		IconURL:      iconURL,
		DisplayOrder: displayOrder,
	}

	stickerType, err := h.service.CreateSticker(req)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to create sticker", err)
	}
	res := &sticker.CreateStickerResponse{}
	res.Body.Success = true
	res.Body.Data = stickerType
	return res, nil
}

func UploadSticker(file multipart.File, filename string) (string, error) {
	supabase := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_API_KEY")
	uploadURL := fmt.Sprintf("%s/storage/v1/object/stickers/%s", supabase, filename)
	file.Seek(0, 0)

	body, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(body)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed: %s", string(respBody))
	}
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/stickers/%s", supabase, filename)
	return publicURL, nil
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
