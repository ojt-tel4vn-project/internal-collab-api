package handlers

import (
	"context"
	"fmt"
	"io"
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
	service         services.DocumentService
	jwtService      crypto.JWTService
	employeeRepo    repository.EmployeeRepository
	categoryService services.DocumentCategoryService
}

// FileOutput không cần thiết nữa - dùng huma.StreamResponse thay thế

func NewDocumentHandler(service services.DocumentService, jwtService crypto.JWTService, employeeRepo repository.EmployeeRepository, categoryService services.DocumentCategoryService) *DocumentHandler {
	return &DocumentHandler{
		service:         service,
		jwtService:      jwtService,
		employeeRepo:    employeeRepo,
		categoryService: categoryService,
	}
}

func (h *DocumentHandler) RegisterRoutes(api huma.API) {
	//Create document (HR only)
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

	// List documents
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

	// Mark document as read
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

	// View document (inline)
	huma.Register(api, huma.Operation{
		OperationID: "view-document",
		Method:      http.MethodGet,
		Path:        "/api/v1/documents/{id}/view",
		Summary:     "View document",
		Tags:        []string{"Documents"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Responses: map[string]*huma.Response{
			"200": {
				Description: "File stream",
				Content: map[string]*huma.MediaType{
					"application/pdf":          {},
					"image/png":                {},
					"image/jpeg":               {},
					"application/octet-stream": {},
				},
			},
		},
	}, h.ViewDocument)

	// Download document (attachment)
	huma.Register(api, huma.Operation{
		OperationID: "download-document",
		Method:      http.MethodGet,
		Path:        "/api/v1/documents/{id}/download",
		Summary:     "Download document",
		Tags:        []string{"Documents"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Responses: map[string]*huma.Response{
			"200": {
				Description: "File stream",
				Content: map[string]*huma.MediaType{
					"application/pdf":          {},
					"image/png":                {},
					"image/jpeg":               {},
					"application/octet-stream": {},
				},
			},
		},
	}, h.DownloadDocument)

	// Create document category (HR only)
	huma.Register(api, huma.Operation{
		OperationID: "create-document-category",
		Method:      http.MethodPost,
		Path:        "/api/v1/hr/document-category",
		Summary:     "Create document category (HR only)",
		Tags:        []string{"Documents", "HR"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.CreateDocumentCategory)
}

// CreateDocument function
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

	roles := "employee" // Default role
	if r, ok := input.RawBody.Value["roles"]; ok && len(r) > 0 {
		roles = r[0]
	}
	allowed := map[string]bool{
		"employee": true,
		"manager":  true,
		"hr":       true,
	}

	splitRoles := strings.Split(roles, ",")
	for _, r := range splitRoles {
		if !allowed[strings.TrimSpace(r)] {
			return nil, huma.Error400BadRequest("invalid role")
		}
	}

	doc := models.Document{
		Title:      title,
		CategoryID: categoryID,
		FilePath:   fileURL,
		Roles:      roles,
		UploadedBy: claims.UserID,
	}

	result, err := h.service.Create(claims.UserID, doc)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to create document", err)
	}

	return &struct{ Body models.Document }{Body: *result}, nil

}

// ListDocuments function
func (h *DocumentHandler) ListDocuments(
	ctx context.Context,
	input *struct {
		Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
	},
) (*struct{ Body []models.Document }, error) {
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
	employee, err := h.employeeRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, huma.Error404NotFound("Employee not found")
	}

	userRole := "employee" // default
	if len(employee.Roles) > 0 {
		userRole = employee.Roles[0].Name
	}

	docs, err := h.service.List(userRole)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list documents", err)
	}
	return &struct{ Body []models.Document }{Body: docs}, nil
}

// ReadDocument function
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

// CreateDocumentCategory function
func (h *DocumentHandler) CreateDocumentCategory(
	ctx context.Context,
	input *struct {
		Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
		Body          struct {
			Name     string     `json:"name" required:"true"`
			ParentID *uuid.UUID `json:"parent_id,omitempty"`
		}
	},
) (*struct{ Body models.DocumentCategory }, error) {
	// Validate HR access
	_, err := authPkg.Authorize(
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
	category := models.DocumentCategory{
		Name:     input.Body.Name,
		ParentID: input.Body.ParentID,
	}
	result, err := h.categoryService.Create(category.Name, category.ParentID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to create document category", err)
	}
	return &struct{ Body models.DocumentCategory }{Body: *result}, nil
}

func hasPermission(docRoles, userRole string) bool {
	if userRole == "admin" {
		return true
	}
	roles := strings.Split(docRoles, ",")
	for _, role := range roles {
		if strings.TrimSpace(role) == userRole {
			return true
		}
	}
	return false
}

// serveDocumentFile is a helper function to serve document files for both view and download endpoints
func (h *DocumentHandler) serveDocumentFile(docID uuid.UUID, role string, inline bool) (*huma.StreamResponse, error) {
	doc, err := h.service.FindByID(docID)
	if err != nil {
		return nil, huma.Error404NotFound("Document not found")
	}

	// Check if user has permission to access the document
	if !hasPermission(doc.Roles, role) {
		return nil, huma.Error403Forbidden("You do not have permission to access this document")
	}

	disposition := "attachment"
	if inline {
		disposition = "inline"
	}
	filename := doc.Title + filepath.Ext(doc.FilePath)

	// Detect content type from file extension
	ext := strings.ToLower(filepath.Ext(doc.FilePath))
	contentTypeMap := map[string]string{
		".pdf":  "application/pdf",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
	contentType := contentTypeMap[ext]
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return &huma.StreamResponse{
		Body: func(ctx huma.Context) {
			resp, err := http.Get(doc.FilePath)
			if err != nil {
				ctx.SetStatus(http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			ctx.SetHeader("Content-Type", contentType)
			ctx.SetHeader("Content-Disposition", fmt.Sprintf(`%s; filename="%s"`, disposition, filename))

			// Stream file to response
			io.Copy(ctx.BodyWriter(), resp.Body)
		},
	}, nil
}

// ViewDocument function
func (h *DocumentHandler) ViewDocument(
	ctx context.Context,
	input *struct {
		Authorization string    `header:"Authorization" required:"true" doc:"Bearer token"`
		ID            uuid.UUID `path:"id" required:"true"`
	},
) (*huma.StreamResponse, error) {
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

	employee, err := h.employeeRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, huma.Error404NotFound("Employee not found")
	}

	userRole := "employee" // default
	if len(employee.Roles) > 0 {
		userRole = employee.Roles[0].Name
	}

	resp, err := h.serveDocumentFile(input.ID, userRole, true)
	if err != nil {
		return nil, err
	}
	// Auto mark as read
	_ = h.service.Read(input.ID, claims.UserID)
	return resp, nil
}

func (h *DocumentHandler) DownloadDocument(
	ctx context.Context,
	input *struct {
		Authorization string    `header:"Authorization" required:"true" doc:"Bearer token"`
		ID            uuid.UUID `path:"id" required:"true"`
	},
) (*huma.StreamResponse, error) {
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

	employee, err := h.employeeRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, huma.Error404NotFound("Employee not found")
	}

	userRole := "employee" // default
	if len(employee.Roles) > 0 {
		userRole = employee.Roles[0].Name
	}

	return h.serveDocumentFile(input.ID, userRole, false)
}
