package handlers

import (
	"context"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	models "github.com/ojt-tel4vn-project/internal-collab-api/models/document"
	authPkg "github.com/ojt-tel4vn-project/internal-collab-api/pkg/auth"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

const (
	MaxFileSize = 10 << 20 // 10 MB
)

var AllowedMimeTypes = map[string]bool{
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"image/png":  true,
	"image/jpeg": true,
}

type DocumentHandler struct {
	service      services.DocumentService
	jwtService   crypto.JWTService
	employeeRepo repository.EmployeeRepository
}

func NewDocumentHandler(service services.DocumentService, jwtService crypto.JWTService, employeeRepo repository.EmployeeRepository) *DocumentHandler {
	return &DocumentHandler{
		service:      service,
		jwtService:   jwtService,
		employeeRepo: employeeRepo,
	}
}

func (h *DocumentHandler) RegisterRoutes(api huma.API) {
	// Document endpoints (require authentication)
	huma.Register(api, huma.Operation{
		OperationID: "create-document",
		Method:      http.MethodPost,
		Path:        "/api/v1/hr/documents",
		Summary:     "Create a document (HR only)",
		Tags:        []string{"HR", "Documents"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.CreateDocument)

	huma.Register(api, huma.Operation{
		OperationID: "list-documents",
		Method:      http.MethodGet,
		Path:        "/api/v1/documents",
		Summary:     "List all documents",
		Tags:        []string{"Documents"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.ListDocuments)

	huma.Register(api, huma.Operation{
		OperationID: "read-document",
		Method:      http.MethodPost,
		Path:        "/api/v1/documents/{id}/read",
		Summary:     "Mark document as read",
		Tags:        []string{"Documents"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.ReadDocument)
}

func (h *DocumentHandler) CreateDocument(
	ctx context.Context,
	input *struct {
		Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
		RawBody       multipart.Form
	},
) (*struct{ Body models.Document }, error) {
	// Validate HR access
	claims, err := authPkg.Authorize(
		input.Authorization,
		h.jwtService,
		h.employeeRepo,
		authPkg.AuthOptions{
			Roles: []string{"hr", "admin"},
		},
	)
	if err != nil {
		return nil, err
	}

	// Get form values
	title := ""
	if titles, ok := input.RawBody.Value["title"]; ok && len(titles) > 0 {
		title = titles[0]
	}
	if title == "" {
		return nil, huma.Error400BadRequest("Title is required")
	}

	categoryIDStr := ""
	if categoryIDs, ok := input.RawBody.Value["category_id"]; ok && len(categoryIDs) > 0 {
		categoryIDStr = categoryIDs[0]
	}
	if categoryIDStr == "" {
		return nil, huma.Error400BadRequest("Category ID is required")
	}
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid category_id format")
	}

	// Get file
	files, ok := input.RawBody.File["file"]
	if !ok || len(files) == 0 {
		return nil, huma.Error400BadRequest("File is required")
	}
	fileHeader := files[0]

	if fileHeader.Size > MaxFileSize {
		return nil, huma.Error400BadRequest("File size exceeds the maximum limit (10 MB)")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to open uploaded file", err)
	}
	defer file.Close()

	// Validate MIME type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to read uploaded file", err)
	}
	mimeType := http.DetectContentType(buffer)
	if !AllowedMimeTypes[mimeType] {
		return nil, huma.Error400BadRequest("Unsupported file type", nil)
	}
	// Reset file read pointer
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to reset file reader", err)
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExts := map[string]bool{
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".png":  true,
		".jpg":  true,
		".jpeg": true,
	}
	if !allowedExts[ext] {
		return nil, huma.Error400BadRequest("Unsupported file extension", nil)
	}

	filename := uuid.New().String() + ext
	path := "documents/" + filename

	// Upload file to storage
	fileURL, err := h.service.UploadFile(ctx, path, file)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to upload file", err)
	}

	doc := models.Document{
		Title:      title,
		CategoryID: categoryID,
		FilePath:   fileURL,
	}

	result, err := h.service.Create(claims.UserID, doc)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to create document", err)
	}

	return &struct{ Body models.Document }{Body: *result}, nil

}

func (h *DocumentHandler) ListDocuments(
	ctx context.Context,
	input *struct {
		Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
	},
) (*struct{ Body []models.Document }, error) {
	// Validate login
	_, err := authPkg.Authorize(
		input.Authorization,
		h.jwtService,
		h.employeeRepo,
		authPkg.AuthOptions{},
	)
	if err != nil {
		return nil, err
	}
	docs, err := h.service.List()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list documents", err)
	}
	return &struct{ Body []models.Document }{Body: docs}, nil
}

func (h *DocumentHandler) ReadDocument(
	ctx context.Context,
	input *struct {
		Authorization string    `header:"Authorization" required:"true" doc:"Bearer token"`
		ID            uuid.UUID `path:"id" required:"true" doc:"Document ID"`
	},
) (*struct {
	Body struct {
		Message string `json:"message"`
	}
}, error) {
	// Validate login
	claims, err := authPkg.Authorize(
		input.Authorization,
		h.jwtService,
		h.employeeRepo,
		authPkg.AuthOptions{
			RequireActive: true,
		},
	)
	if err != nil {
		return nil, err
	}
	err = h.service.Read(input.ID, claims.UserID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to mark document as read", err)
	}
	return &struct {
		Body struct {
			Message string `json:"message"`
		}
	}{
		Body: struct {
			Message string `json:"message"`
		}{
			Message: "Document marked as read",
		},
	}, nil
}
