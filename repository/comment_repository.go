package repository

import (
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(comment *models.Comment) error
	FindByAttendanceID(attendanceID uuid.UUID) ([]models.Comment, error)
	CountByAttendanceID(attendanceID uuid.UUID) (int64, error)
	FindByID(id uuid.UUID) (*models.Comment, error)
	MarkRead(id uuid.UUID) error
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

func (r *commentRepository) FindByAttendanceID(attendanceID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.
		Preload("Author").
		Where("attendance_id = ?", attendanceID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

func (r *commentRepository) CountByAttendanceID(attendanceID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Comment{}).Where("attendance_id = ?", attendanceID).Count(&count).Error
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

func (r *commentRepository) MarkRead(id uuid.UUID) error {
	return r.db.Model(&models.Comment{}).Where("id = ?", id).Update("is_read", true).Error
}

func (r *commentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Comment{}, "id = ?", id).Error
}
