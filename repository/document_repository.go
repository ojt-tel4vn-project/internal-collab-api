package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type DocumentRepository interface {
	Create(doc *models.Document) error
	FindByRole(role string) ([]models.Document, error)
	MarkAsRead(documentID, employeeID uuid.UUID) error
	GetReaders(docID uuid.UUID) ([]uuid.UUID, error)
	Exists(docID uuid.UUID) (bool, error)
	FindByID(docID uuid.UUID) (*models.Document, error)
}

type documentRepositoryImpl struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepositoryImpl{db: db}
}

func (r *documentRepositoryImpl) Create(doc *models.Document) error {
	return r.db.Create(doc).Error
}

func (r *documentRepositoryImpl) FindByRole(role string) ([]models.Document, error) {
	var documents []models.Document
	query := r.db.Order("created_at desc")
	// All employees can see public documents; admin/hr can see all
	if role != "admin" && role != "hr" {
		query = query.Where("is_public = ?", true)
	}
	err := query.Find(&documents).Error
	return documents, err
}

func (r *documentRepositoryImpl) MarkAsRead(documentID, employeeID uuid.UUID) error {
	read := models.DocumentRead{
		DocumentID: documentID,
		EmployeeID: employeeID,
		ReadAt:     time.Now(),
	}
	return r.db.FirstOrCreate(&read).Error
}

func (r *documentRepositoryImpl) GetReaders(docID uuid.UUID) ([]uuid.UUID, error) {
	var readers []uuid.UUID
	err := r.db.Model(&models.DocumentRead{}).
		Where("document_id = ?", docID).
		Pluck("employee_id", &readers).Error
	return readers, err
}

func (r *documentRepositoryImpl) Exists(docID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Document{}).
		Where("id = ?", docID).
		Count(&count).Error
	return count > 0, err
}

func (r *documentRepositoryImpl) FindByID(docID uuid.UUID) (*models.Document, error) {
	var document models.Document
	err := r.db.First(&document, "id = ?", docID).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}
