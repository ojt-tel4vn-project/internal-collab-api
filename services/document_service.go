package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	docDTO "github.com/ojt-tel4vn-project/internal-collab-api/dtos/document"
	"github.com/ojt-tel4vn-project/internal-collab-api/internal/storage"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/response"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/utils"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
)

const MaxFileSize = 10 << 20 // 10MB

var AllowedMimeTypes = map[string]bool{
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"image/png":  true,
	"image/jpeg": true,
}

var AllowedExtensions = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
}

type DocumentService interface {
	Create(employeeID uuid.UUID, doc models.Document) (*models.Document, error)
	List(role string, employeeID uuid.UUID) ([]docDTO.DocumentResponse, error)
	Read(docID, employeeID uuid.UUID) error
	UploadFile(ctx context.Context, path string, file io.Reader) (string, error)
	FindByID(docID uuid.UUID) (*models.Document, error)
	ExistsByTitle(title string) (bool, error)
	Update(doc models.Document) (*models.Document, error)
	Delete(docID uuid.UUID) error

	// Validation methods
	ValidateFile(filename string, fileSize int64, mimeType string) error
	ValidateCreateRequest(title, categoryIDStr, roles, description string) error
	DetectAndValidateMimeType(buffer []byte, filename string) (string, error)

	// Permission methods
	HasPermission(docRoles, userRole string) bool

	GenerateStoragePath(title, originalFilename string) (string, string)
}

type documentServiceImpl struct {
	repo         repository.DocumentRepository
	employeeRepo repository.EmployeeRepository
	storage      storage.StorageService
	notifService NotificationService
}

func NewDocumentService(repo repository.DocumentRepository, employeeRepo repository.EmployeeRepository, storage storage.StorageService, notifService NotificationService) DocumentService {
	return &documentServiceImpl{
		repo:         repo,
		employeeRepo: employeeRepo,
		storage:      storage,
		notifService: notifService,
	}
}

func (s *documentServiceImpl) Create(employeeID uuid.UUID, doc models.Document) (*models.Document, error) {
	doc.UploadedBy = employeeID
	if err := s.repo.Create(&doc); err != nil {
		return nil, response.InternalServerError("Failed to create document")
	}

	// Send notifications to allowed roles
	go func() {
		var targetRoles []string
		if doc.Roles == "all" {
			targetRoles = []string{"employee", "manager", "hr"}
		} else {
			roles := strings.Split(doc.Roles, ",")
			for _, r := range roles {
				targetRoles = append(targetRoles, strings.TrimSpace(r))
			}
		}

		employees, err := s.employeeRepo.FindEmployeesByRoles(targetRoles)
		if err == nil {
			title := "New Document Uploaded"
			message := fmt.Sprintf("A new document '%s' has been uploaded and is available for your role.", doc.Title)
			entityType := "document"
			actionURL := "/documents" // Adjust if needed

			for _, emp := range employees {
				// Don't notify the uploader
				if emp.ID == employeeID {
					continue
				}
				_ = s.notifService.SendNotification(emp.ID, "document", title, message, &entityType, &doc.ID, models.PriorityNormal)
				if actionURL != "" {
					// We might want to update the last notification with the action URL
					// but SendNotification currently takes entityType and ID which is used for building links
				}
			}
		}
	}()

	return &doc, nil
}

func (s *documentServiceImpl) List(role string, employeeID uuid.UUID) ([]docDTO.DocumentResponse, error) {
	// Lấy tất cả document mà Role này xem được
	docs, err := s.repo.FindByRole(role)
	if err != nil {
		return nil, response.InternalServerError("Failed to list documents")
	}

	// Lấy danh sách ID document mà User này đã click xem/đọc
	readIDs, err := s.repo.GetReadDocumentIDs(employeeID)
	if err != nil {
		return nil, response.InternalServerError("Failed to fetch reading statuses")
	}

	// Tạo Hashset (Map) tra cứu ID đã đọc O(1)
	readMap := make(map[uuid.UUID]bool, len(readIDs))
	for _, id := range readIDs {
		readMap[id] = true
	}

	// Ánh xạ ra DTO với trường IsRead boolean
	out := make([]docDTO.DocumentResponse, len(docs))
	for i, d := range docs {
		out[i] = docDTO.DocumentResponse{
			ID:          d.ID,
			Title:       d.Title,
			Description: d.Description,
			CategoryID:  d.CategoryID,
			FileName:    d.FileName,
			Roles:       d.Roles,
			FileSize:    utils.FormatFileSize(d.FileSize),
			MimeType:    d.MimeType,
			UploadedBy:  d.UploadedBy,
			IsRead:      readMap[d.ID],
			CreatedAt:   d.CreatedAt,
		}
	}

	return out, nil
}

func (s *documentServiceImpl) Read(docID, employeeID uuid.UUID) error {
	exists, err := s.repo.Exists(docID)
	if err != nil {
		return response.InternalServerError("Failed to check document existence")
	}
	if !exists {
		return response.NotFound("Document not found")
	}
	return s.repo.MarkAsRead(docID, employeeID)
}

func (s *documentServiceImpl) UploadFile(
	ctx context.Context,
	path string,
	file io.Reader,
) (string, error) {
	return s.storage.UploadFile(ctx, path, file)
}

func (s *documentServiceImpl) FindByID(docID uuid.UUID) (*models.Document, error) {
	return s.repo.FindByID(docID)
}

func (s *documentServiceImpl) ExistsByTitle(title string) (bool, error) {
	return s.repo.ExistsByTitle(title)
}

func (s *documentServiceImpl) Update(doc models.Document) (*models.Document, error) {
	if err := s.repo.Update(&doc); err != nil {
		return nil, response.InternalServerError("Failed to update document")
	}
	return &doc, nil
}

// ValidateFile checks file size, MIME type, and extension
func (s *documentServiceImpl) ValidateFile(filename string, fileSize int64, mimeType string) error {
	// Check file size
	if fileSize > MaxFileSize {
		return response.BadRequest("File size exceeds the maximum limit (10 MB)")
	}
	// Check MIME type
	if !AllowedMimeTypes[mimeType] {
		return response.BadRequest("Unsupported file type")
	}
	// Check file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if !AllowedExtensions[ext] {
		return response.BadRequest("Unsupported file extension")
	}
	return nil
}

// ValidateCreateRequest validates the create request parameters
func (s *documentServiceImpl) ValidateCreateRequest(title, categoryIDStr, roles, description string) error {
	// Validate title
	if strings.TrimSpace(title) == "" {
		return response.BadRequest("File name cannot be empty")
	}
	// Check title uniqueness
	exists, err := s.repo.ExistsByTitle(title)
	if err != nil {
		return response.InternalServerError("Failed to check document title uniqueness")
	}
	if exists {
		return response.BadRequest("A document with the same title already exists")
	}
	// Validate category ID
	if categoryIDStr == "" {
		return response.BadRequest("Category ID is required")
	}
	_, err = uuid.Parse(categoryIDStr)
	if err != nil {
		return response.BadRequest("Invalid category_id format")
	}
	// Validate roles
	if roles == "" {
		return response.BadRequest("Roles field is required")
	}

	allowedRoles := map[string]bool{
		"employee": true,
		"manager":  true,
		"hr":       true,
		"all":      true,
	}

	for _, r := range strings.Split(roles, ",") {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if !allowedRoles[r] {
			return response.BadRequest(fmt.Sprintf("Invalid role: %s", r))
		}
	}
	return nil
}

// HasPermission checks if a user with the given role can access a document
func (s *documentServiceImpl) HasPermission(docRoles, userRole string) bool {
	if userRole == "hr" {
		return true
	}

	if docRoles == "all" {
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

// DetectAndValidateMimeType reads file buffer and validates MIME type
func (s *documentServiceImpl) DetectAndValidateMimeType(buffer []byte, filename string) (string, error) {
	mimeType := http.DetectContentType(buffer)
	ext := strings.ToLower(filepath.Ext(filename))

	// DOCX files are essentially ZIP archives, so DetectContentType returns "application/zip"
	if mimeType == "application/zip" && ext == ".docx" {
		mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	if !AllowedMimeTypes[mimeType] {
		return "", response.BadRequest(fmt.Sprintf("Unsupported file type: %s", mimeType))
	}

	return mimeType, nil
}

func (s *documentServiceImpl) GenerateStoragePath(title, originalFilename string) (string, string) {
	ext := strings.ToLower(filepath.Ext(originalFilename))
	filename := uuid.New().String() + ext
	path := "documents/" + filename

	return path, filename
}

func (s *documentServiceImpl) Delete(docID uuid.UUID) error {
	return s.repo.Delete(docID)
}
