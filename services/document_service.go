package services

import (
	"context"
	"io"

	"github.com/google/uuid"
	docDTO "github.com/ojt-tel4vn-project/internal-collab-api/dtos/document"
	"github.com/ojt-tel4vn-project/internal-collab-api/internal/storage"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/response"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
)

type DocumentService interface {
	Create(employeeID uuid.UUID, doc models.Document) (*models.Document, error)
	List(role string, employeeID uuid.UUID) ([]docDTO.DocumentResponse, error)
	Read(docID, employeeID uuid.UUID) error
	UploadFile(ctx context.Context, path string, file io.Reader) (string, error)
	FindByID(docID uuid.UUID) (*models.Document, error)
}

type documentServiceImpl struct {
	repo    repository.DocumentRepository
	storage storage.StorageService
}

func NewDocumentService(repo repository.DocumentRepository, storage storage.StorageService) DocumentService {
	return &documentServiceImpl{
		repo:    repo,
		storage: storage,
	}
}

func (s *documentServiceImpl) Create(employeeID uuid.UUID, doc models.Document) (*models.Document, error) {
	doc.UploadedBy = employeeID
	if err := s.repo.Create(&doc); err != nil {
		return nil, response.InternalServerError("Failed to create document")
	}
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
			FileSize:    d.FileSize,
			MimeType:    d.MimeType,
			Roles:       d.Roles,
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
