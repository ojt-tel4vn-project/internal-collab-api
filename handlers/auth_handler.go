package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos/auth"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

type AuthHandler struct {
	service    services.AuthService
	jwtService crypto.JWTService
}

func NewAuthHandler(service services.AuthService, jwtService crypto.JWTService) *AuthHandler {
	return &AuthHandler{
		service:    service,
		jwtService: jwtService,
	}
}

func (h *AuthHandler) RegisterRoutes(api huma.API) {
	// Public routes
	huma.Register(api, huma.Operation{
		OperationID: "auth-login",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/login",
		Summary:     "User Login",
		Tags:        []string{"Auth"},
	}, h.Login)

	huma.Register(api, huma.Operation{
		OperationID: "auth-first-time-setup",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/first-time-setup",
		Summary:     "First-time Password Setup (for new employees)",
		Tags:        []string{"Auth"},
	}, h.FirstTimeSetup)

	// Protected routes
	huma.Register(api, huma.Operation{
		OperationID: "auth-change-password",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/change-password",
		Summary:     "Change Password (Requires Authentication)",
		Tags:        []string{"Auth"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.ChangePassword)
}

func (h *AuthHandler) Login(ctx context.Context, input *struct {
	Body auth.LoginRequest
}) (*struct {
	Body auth.LoginResponse
}, error) {
	resp, err := h.service.Login(&input.Body)
	if err != nil {
		return nil, huma.Error401Unauthorized("Login failed", err)
	}
	return &struct{ Body auth.LoginResponse }{Body: *resp}, nil
}

func (h *AuthHandler) FirstTimeSetup(ctx context.Context, input *struct {
	Body  auth.FirstTimeSetupRequest
	Email string `query:"email" required:"true" doc:"Employee email address"`
}) (*struct {
	Body auth.FirstTimeSetupResponse
}, error) {
	resp, err := h.service.FirstTimeSetup(input.Email, &input.Body)
	if err != nil {
		return nil, huma.Error400BadRequest("First-time setup failed", err)
	}
	return &struct{ Body auth.FirstTimeSetupResponse }{Body: *resp}, nil
}

func (h *AuthHandler) ChangePassword(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
	Body          auth.ChangePasswordRequest
}) (*struct {
	Body auth.ChangePasswordResponse
}, error) {
	// Validate JWT token
	if !strings.HasPrefix(input.Authorization, "Bearer ") {
		return nil, huma.Error401Unauthorized("Invalid authorization format. Use: Bearer <token>")
	}

	token := strings.TrimPrefix(input.Authorization, "Bearer ")
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid or expired token")
	}

	// Use employee ID from JWT claims
	resp, err := h.service.ChangePassword(claims.UserID, &input.Body)
	if err != nil {
		return nil, huma.Error400BadRequest("Password change failed", err)
	}
	return &struct{ Body auth.ChangePasswordResponse }{Body: *resp}, nil
}
