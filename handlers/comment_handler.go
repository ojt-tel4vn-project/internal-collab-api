package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	commentDTO "github.com/ojt-tel4vn-project/internal-collab-api/dtos/comment"
	authPkg "github.com/ojt-tel4vn-project/internal-collab-api/pkg/auth"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

type CommentHandler struct {
	service      services.CommentService
	jwtService   crypto.JWTService
	employeeRepo repository.EmployeeRepository
}

func NewCommentHandler(service services.CommentService, jwtService crypto.JWTService, employeeRepo repository.EmployeeRepository) *CommentHandler {
	return &CommentHandler{
		service:      service,
		jwtService:   jwtService,
		employeeRepo: employeeRepo,
	}
}

func (h *CommentHandler) RegisterRoutes(api huma.API) {
	// GET /api/v1/documents/{id}/comments  — list comments + total
	huma.Register(api, huma.Operation{
		OperationID: "list-document-comments",
		Method:      http.MethodGet,
		Path:        "/api/v1/documents/{id}/comments",
		Summary:     "List comments for a document",
		Tags:        []string{"Documents", "Comments"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.ListComments)

	// POST /api/v1/documents/{id}/comments  — add a comment
	huma.Register(api, huma.Operation{
		OperationID: "create-document-comment",
		Method:      http.MethodPost,
		Path:        "/api/v1/documents/{id}/comments",
		Summary:     "Add a comment to a document",
		Tags:        []string{"Documents", "Comments"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.CreateComment)

	// DELETE /api/v1/comments/{commentId}  — delete a comment (author or HR)
	huma.Register(api, huma.Operation{
		OperationID: "delete-document-comment",
		Method:      http.MethodDelete,
		Path:        "/api/v1/comments/{commentId}",
		Summary:     "Delete a comment (author or HR/admin only)",
		Tags:        []string{"Comments"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.DeleteComment)
}

// ListComments GET /api/v1/documents/{id}/comments
func (h *CommentHandler) ListComments(ctx context.Context, input *struct {
	Authorization string    `header:"Authorization" required:"true" doc:"Bearer token"`
	ID            uuid.UUID `path:"id" doc:"Document ID"`
}) (*struct {
	Body commentDTO.ListCommentsResponse
}, error) {
	_, err := authPkg.Authorize(input.Authorization, h.jwtService, h.employeeRepo, authPkg.AuthOptions{RequireActive: true})
	if err != nil {
		return nil, err
	}

	resp, err := h.service.GetCommentsByDocument(input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch comments", err)
	}

	return &struct {
		Body commentDTO.ListCommentsResponse
	}{Body: *resp}, nil
}

// CreateComment POST /api/v1/documents/{id}/comments
func (h *CommentHandler) CreateComment(ctx context.Context, input *struct {
	Authorization string    `header:"Authorization" required:"true" doc:"Bearer token"`
	ID            uuid.UUID `path:"id" doc:"Document ID"`
	Body          commentDTO.CreateCommentRequest
}) (*struct {
	Body commentDTO.CreateCommentResponse
}, error) {
	claims, err := authPkg.Authorize(input.Authorization, h.jwtService, h.employeeRepo, authPkg.AuthOptions{RequireActive: true})
	if err != nil {
		return nil, err
	}

	resp, err := h.service.CreateComment(input.ID, claims.UserID, &input.Body)
	if err != nil {
		return nil, err
	}

	return &struct {
		Body commentDTO.CreateCommentResponse
	}{Body: *resp}, nil
}

// DeleteComment DELETE /api/v1/comments/{commentId}
func (h *CommentHandler) DeleteComment(ctx context.Context, input *struct {
	Authorization string    `header:"Authorization" required:"true" doc:"Bearer token"`
	CommentID     uuid.UUID `path:"commentId" doc:"Comment ID"`
}) (*struct {
	Body commentDTO.DeleteCommentResponse
}, error) {
	claims, err := authPkg.Authorize(input.Authorization, h.jwtService, h.employeeRepo, authPkg.AuthOptions{RequireActive: true})
	if err != nil {
		return nil, err
	}

	// Check if the user has HR/admin role (they can delete anyone's comment)
	isHR := false
	emp, lookupErr := h.employeeRepo.FindByID(claims.UserID)
	if lookupErr == nil && emp.Role != nil {
		r := emp.Role.Name
		isHR = r == "hr" || r == "admin"
	}

	if err := h.service.DeleteComment(input.CommentID, claims.UserID, isHR); err != nil {
		return nil, err
	}

	return &struct {
		Body commentDTO.DeleteCommentResponse
	}{Body: commentDTO.DeleteCommentResponse{Message: "Comment deleted successfully"}}, nil
}
