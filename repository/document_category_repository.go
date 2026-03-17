package repository

import (
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type DocumentCategoryRepository interface {
	Create(c *models.DocumentCategory) error
	FindByID(id uuid.UUID) (*models.DocumentCategory, error)
	FindByName(name string) (*models.DocumentCategory, error)
	FindAll() ([]models.DocumentCategory, error)
}

type documentCategoryRepositoryImpl struct {
	db *gorm.DB
}

func NewDocumentCategoryRepository(db *gorm.DB) DocumentCategoryRepository {
	return &documentCategoryRepositoryImpl{db: db}
}

func (r *documentCategoryRepositoryImpl) Create(c *models.DocumentCategory) error {
	return r.db.Create(c).Error
}

func (r *documentCategoryRepositoryImpl) FindByID(id uuid.UUID) (*models.DocumentCategory, error) {
	var category models.DocumentCategory
	err := r.db.First(&category, "id = ?", id).Error
	return &category, err
}

func (r *documentCategoryRepositoryImpl) FindByName(name string) (*models.DocumentCategory, error) {
	var category models.DocumentCategory
	err := r.db.First(&category, "name = ?", name).Error
	return &category, err
}

func (r *documentCategoryRepositoryImpl) FindAll() ([]models.DocumentCategory, error) {
	var categories []models.DocumentCategory
	return categories, r.db.Order("created_at desc").Find(&categories).Error
}