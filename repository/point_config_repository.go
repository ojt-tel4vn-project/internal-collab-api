package repository

import (
	"time"

	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type PointConfigRepository interface {
	GetPointConfig() (*models.PointConfig, error)
	WithTransaction(tx *gorm.DB) PointConfigRepository
	UpdatePointConfig(config *models.PointConfig) error
}

type pointConfigRepositoryImpl struct {
	db *gorm.DB
}

func NewPointConfigRepository(db *gorm.DB) PointConfigRepository {
	return &pointConfigRepositoryImpl{
		db: db,
	}
}

func (r *pointConfigRepositoryImpl) GetPointConfig() (*models.PointConfig, error) {
	var config models.PointConfig
	// Lấy bản ghi cập nhật sau cùng => lần nhập gần nhất
	err := r.db.Order("updated_at desc").First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *pointConfigRepositoryImpl) WithTransaction(tx *gorm.DB) PointConfigRepository {
	return &pointConfigRepositoryImpl{db: tx}
}

func (r *pointConfigRepositoryImpl) UpdatePointConfig(config *models.PointConfig) error {
	var existing models.PointConfig
	if err := r.db.First(&existing).Error; err != nil {
		return err
	}

	existing.YearlyPoints = config.YearlyPoints
	existing.ResetMonth = config.ResetMonth
	existing.ResetDay = config.ResetDay
	existing.UpdatedAt = time.Now()

	return r.db.Save(&existing).Error
}
