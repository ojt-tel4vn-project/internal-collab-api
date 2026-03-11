package repository

import (
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(comment *models.Comment) error
	FindByDocumentID(documentID uuid.UUID) ([]models.Comment, error)
	CountByDocumentID(documentID uuid.UUID) (int64, error)
	FindByID(id uuid.UUID) (*models.Comment, error)
	Delete(id uuid.UUID) error
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

func (r *commentRepository) FindByDocumentID(documentID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.
		Preload("Author").
		Where("document_id = ?", documentID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

func (r *commentRepository) CountByDocumentID(documentID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Comment{}).Where("document_id = ?", documentID).Count(&count).Error
	return count, err
}

func (r *commentRepository) FindByID(id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.Preload("Author").First(&comment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Comment{}, "id = ?", id).Error
}
