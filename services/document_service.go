package services

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/internal/storage"
	models "github.com/ojt-tel4vn-project/internal-collab-api/models/document"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/response"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
)

type DocumentService interface {
	Create(employeeID uuid.UUID, doc models.Document) (*models.Document, error)
	List() ([]models.Document, error)
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

func (s *documentServiceImpl) List() ([]models.Document, error) {
	return s.repo.FindAll()
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