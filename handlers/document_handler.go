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
	docDTO "github.com/ojt-tel4vn-project/internal-collab-api/dtos/document"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	authPkg "github.com/ojt-tel4vn-project/internal-collab-api/pkg/auth"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/utils"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

type DocumentHandler struct {
	service         services.DocumentService
	jwtService      crypto.JWTService
	employeeRepo    repository.EmployeeRepository
	categoryService services.DocumentCategoryService
}

var ContentTypeMap = map[string]string{
	".pdf":  "application/pdf",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
}

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
		RequestBody: &huma.RequestBody{
			Content: map[string]*huma.MediaType{
				"multipart/form-data": {
					Schema: &huma.Schema{
						Type: "object",
						Properties: map[string]*huma.Schema{
							"file":        {Type: "string", Format: "binary", Description: "File to upload (key: file)"},
							"category_id": {Type: "string", Format: "uuid", Description: "Document category ID"},
							"roles":       {Type: "string", Description: "Comma-separated roles that can access the document (default: employee)"},
							"description": {Type: "string", Description: "Description of the document"},
						},
						Required: []string{"file", "category_id"},
					},
				},
			},
		},
	}, h.CreateDocument)

	// List documents (HR sees all, employees see public)
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

	// List documents - HR alias
	huma.Register(api, huma.Operation{
		OperationID: "list-hr-documents",
		Method:      http.MethodGet,
		Path:        "/api/v1/hr/documents",
		Summary:     "List all documents (HR)",
		Tags:        []string{"HR", "Documents"},
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

	// List document categories (all authenticated users)
	huma.Register(api, huma.Operation{
		OperationID: "list-document-categories",
		Method:      http.MethodGet,
		Path:        "/api/v1/hr/document-category",
		Summary:     "List document categories (HR)",
		Tags:        []string{"Documents", "HR"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.ListDocumentCategories)

	// List document categories - public alias
	huma.Register(api, huma.Operation{
		OperationID: "list-document-categories-public",
		Method:      http.MethodGet,
		Path:        "/api/v1/documents/categories",
		Summary:     "List document categories (all employees)",
		Tags:        []string{"Documents"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.ListDocumentCategories)

	//Update Document - HR
	huma.Register(api, huma.Operation{
		OperationID: "update-document",
		Method:      http.MethodPut,
		Path:        "/api/v1/hr/documents/{id}",
		Summary:     "Update document metadata (HR only)",
		Tags:        []string{"HR", "Documents"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.UpdateDocument)

	// Delete Document - HR
	huma.Register(api, huma.Operation{
		OperationID: "delete-document",
		Method:      http.MethodDelete,
		Path:        "/api/v1/hr/documents/{id}",
		Summary:     "Delete document (HR only)",
		Tags:        []string{"HR", "Documents"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.DeleteDocument)
}

func (h *DocumentHandler) CreateDocument(
	ctx context.Context,
	input *struct {
		Authorization string         `header:"Authorization" required:"true" doc:"Bearer token"`
		RawBody       multipart.Form `contentType:"multipart/form-data"`
	},
) (*struct{ Body docDTO.DocumentResponse }, error) {
	// 1. AUTHORIZATION: Validate HR access
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

	files, ok := input.RawBody.File["file"]
	if !ok || len(files) == 0 {
		return nil, huma.Error400BadRequest("File is required (key: file)")
	}
	fileHeader := files[0]

	categoryIDStr := ""
	if categoryIDs, ok := input.RawBody.Value["category_id"]; ok && len(categoryIDs) > 0 {
		categoryIDStr = categoryIDs[0]
	}

	roles := "employee"
	if r, ok := input.RawBody.Value["roles"]; ok && len(r) > 0 {
		roles = r[0]
	}

	description := ""
	if d, ok := input.RawBody.Value["description"]; ok && len(d) > 0 {
		description = strings.TrimSpace(d[0])
	}

	// Extract title from filename
	title := strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename))
	title = strings.TrimSpace(title)

	if err := h.service.ValidateCreateRequest(title, categoryIDStr, roles, description); err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to open uploaded file", err)
	}
	defer file.Close()

	// Detect MIME type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to read uploaded file", err)
	}

	mimeType, err := h.service.DetectAndValidateMimeType(buffer, fileHeader.Filename)
	if err != nil {
		return nil, err
	}

	// Validate file using service
	if err := h.service.ValidateFile(fileHeader.Filename, fileHeader.Size, mimeType); err != nil {
		return nil, err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to reset file reader", err)
	}

	filename, path := h.service.GenerateStoragePath(title, fileHeader.Filename)

	fileURL, err := h.service.UploadFile(ctx, path, file)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to upload file", err)
	}

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid categoryID")
	}

	doc := models.Document{
		Title:       title,
		FileName:    filename,
		Description: description,
		CategoryID:  categoryID,
		FilePath:    fileURL,
		MimeType:    mimeType,
		FileSize:    fileHeader.Size,
		Roles:       roles,
		UploadedBy:  claims.UserID,
	}

	result, err := h.service.Create(claims.UserID, doc)
	if err != nil {
		return nil, err
	}

	return &struct{ Body docDTO.DocumentResponse }{
		Body: docDTO.DocumentResponse{
			ID:          result.ID,
			Title:       result.Title,
			Description: result.Description,
			CategoryID:  result.CategoryID,
			FileName:    result.FileName,
			FileSize:    utils.FormatFileSize(result.FileSize),
			MimeType:    result.MimeType,
			Roles:       result.Roles,
			UploadedBy:  result.UploadedBy,
			CreatedAt:   result.CreatedAt,
		},
	}, nil
}

// ListDocuments function
func (h *DocumentHandler) ListDocuments(
	ctx context.Context,
	input *struct {
		Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
	},
) (*docDTO.ListDocumentResponse, error) {
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
	if employee.Role != nil {
		userRole = employee.Role.Name
	}

	docs, err := h.service.List(userRole, claims.UserID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list documents", err)
	}
	return &docDTO.ListDocumentResponse{
		Body: struct {
			Data []docDTO.DocumentResponse `json:"data"`
		}{
			Data: docs,
		},
	}, nil
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

// ListDocumentCategories lists all document categories
func (h *DocumentHandler) ListDocumentCategories(
	ctx context.Context,
	input *struct {
		Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
	},
) (*struct {
	Body struct {
		Data []models.DocumentCategory `json:"data"`
	}
}, error) {
	_, err := authPkg.Authorize(input.Authorization, h.jwtService, h.employeeRepo, authPkg.AuthOptions{RequireActive: true})
	if err != nil {
		return nil, err
	}
	cats, err := h.categoryService.List()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list categories", err)
	}
	return &struct {
		Body struct {
			Data []models.DocumentCategory `json:"data"`
		}
	}{Body: struct {
		Data []models.DocumentCategory `json:"data"`
	}{Data: cats}}, nil
}

// serveDocumentFile is a helper function to serve document files for both view and download endpoints
func (h *DocumentHandler) serveDocumentFile(docID uuid.UUID, role string, inline bool) (*huma.StreamResponse, error) {
	doc, err := h.service.FindByID(docID)
	if err != nil {
		return nil, huma.Error404NotFound("Document not found")
	}

	// Check if user has permission to access the document using service
	if !h.service.HasPermission(doc.Roles, role) {
		return nil, huma.Error403Forbidden("You do not have permission to access this document")
	}

	disposition := "attachment"
	if inline {
		disposition = "inline"
	}
	filename := doc.Title + filepath.Ext(doc.FilePath)

	// Detect content type from file extension
	ext := strings.ToLower(filepath.Ext(doc.FilePath))
	contentType := ContentTypeMap[ext]
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
	if employee.Role != nil {
		userRole = employee.Role.Name
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
	if employee.Role != nil {
		userRole = employee.Role.Name
	}

	return h.serveDocumentFile(input.ID, userRole, false)
}

func (h *DocumentHandler) UpdateDocument(
	ctx context.Context,
	input *struct {
		Authorization string    `header:"Authorization" required:"true"`
		ID            uuid.UUID `path:"id" required:"true"`
		Body          struct {
			Title       string     `json:"title,omitempty"`
			Description string     `json:"description,omitempty"`
			CategoryID  *uuid.UUID `json:"category_id,omitempty"`
			Roles       string     `json:"roles,omitempty"`
		}
	},
) (*struct{ Body docDTO.DocumentResponse }, error) {

	// HR/Admin only
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

	doc, err := h.service.FindByID(input.ID)
	if err != nil {
		return nil, huma.Error404NotFound("Document not found")
	}

	if input.Body.Title != "" {
		doc.Title = input.Body.Title
	}

	if input.Body.Description != "" {
		doc.Description = input.Body.Description
	}

	if input.Body.CategoryID != nil {
		doc.CategoryID = *input.Body.CategoryID
	}

	if input.Body.Roles != "" {
		doc.Roles = input.Body.Roles
	}

	updated, err := h.service.Update(*doc)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to update document", err)
	}

	return &struct{ Body docDTO.DocumentResponse }{
		Body: docDTO.DocumentResponse{
			ID:          updated.ID,
			Title:       updated.Title,
			Description: updated.Description,
			CategoryID:  updated.CategoryID,
			FileName:    updated.FileName,
			Roles:       updated.Roles,
			FileSize:    utils.FormatFileSize(updated.FileSize),
			MimeType:    updated.MimeType,
			UploadedBy:  updated.UploadedBy,
			IsRead:      false,
			CreatedAt:   updated.CreatedAt,
			UpdatedAt:   updated.UpdatedAt,
		},
	}, nil
}

func (h *DocumentHandler) DeleteDocument(
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
	// HR/Admin only
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

	// Check if document exists
	doc, err := h.service.FindByID(input.ID)
	if err != nil {
		return nil, huma.Error404NotFound("Document not found")
	}

	// Delete the document
	if err := h.service.Delete(input.ID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete document", err)
	}

	return &struct {
		Body struct {
			Message string `json:"message"`
		}
	}{
		Body: struct {
			Message string `json:"message"`
		}{
			Message: fmt.Sprintf("Document '%s' deleted successfully", doc.Title),
		},
	}, nil
}
