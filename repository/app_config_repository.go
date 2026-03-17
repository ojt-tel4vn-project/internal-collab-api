package repository

import (
	"errors"

	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AppConfigRepository interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

type appConfigRepository struct {
	db *gorm.DB
}

func NewAppConfigRepository(db *gorm.DB) AppConfigRepository {
	return &appConfigRepository{db: db}
}

// Get trả về giá trị theo key. Trả về error nếu key không tồn tại.
func (r *appConfigRepository) Get(key string) (string, error) {
	var cfg models.AppConfig
	err := r.db.Where("key = ?", key).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", gorm.ErrRecordNotFound
		}
		return "", err
	}
	return cfg.Value, nil
}

// Set tạo mới hoặc cập nhật giá trị theo key (upsert).
func (r *appConfigRepository) Set(key, value string) error {
	cfg := models.AppConfig{Key: key, Value: value}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&cfg).Error
}
