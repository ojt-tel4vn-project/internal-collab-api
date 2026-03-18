package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type StickerRepository interface {
	WithTransaction(tx *gorm.DB) StickerRepository
	GetPointBalance(employeeID uuid.UUID, year int) (*models.PointBalance, error)
	UpdatePointBalance(balance *models.PointBalance) error
	GetStickerTypeByID(stickerTypeID uuid.UUID) (*models.StickerType, error)
	CreateStickerTransaction(tx *models.StickerTransaction) error
	GetLeaderboard(filter LeaderboardFilter) ([]LeaderboardResult, error)
	CreatePointBalance(pointBalance *models.PointBalance) error
	CreateSticker(sticker *models.StickerType) error
	ListStickerTypes() ([]models.StickerType, error)
}

type LeaderboardFilter struct {
	Limit        int
	StartDate    *time.Time
	EndDate      *time.Time
	DepartmentID *uuid.UUID
}

type LeaderboardResult struct {
	EmployeeID uuid.UUID `gorm:"column:employee_id" json:"employee_id"`
	FullName   string    `gorm:"->;column:full_name" json:"full_name"`
	Total      int       `gorm:"column:total" json:"total"`
	Department string    `gorm:"column:department_name" json:"department"`
}

type stickerRepositoryImpl struct {
	db *gorm.DB
}

func NewStickerRepository(db *gorm.DB) StickerRepository {
	return &stickerRepositoryImpl{
		db: db,
	}
}

func (r *stickerRepositoryImpl) WithTransaction(tx *gorm.DB) StickerRepository {
	return &stickerRepositoryImpl{db: tx}
}

func (r *stickerRepositoryImpl) CreatePointBalance(pointBalance *models.PointBalance) error {
	return r.db.Create(pointBalance).Error
}

func (r *stickerRepositoryImpl) CreateSticker(sticker *models.StickerType) error {
	return r.db.Create(sticker).Error
}

func (r *stickerRepositoryImpl) GetPointBalance(employeeID uuid.UUID, year int) (*models.PointBalance, error) {
	var balance models.PointBalance
	err := r.db.
		Where("employee_id = ? AND year = ?", employeeID, year).
		First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

func (r *stickerRepositoryImpl) UpdatePointBalance(balance *models.PointBalance) error {
	return r.db.Model(&models.PointBalance{}).
		Where("id = ?", balance.ID).
		Update("current_points", balance.CurrentPoints).Error
}

func (r *stickerRepositoryImpl) GetStickerTypeByID(stickerTypeID uuid.UUID) (*models.StickerType, error) {
	var stickerType models.StickerType
	err := r.db.
		Where("id = ? AND is_active = ?", stickerTypeID, true).
		First(&stickerType).Error
	if err != nil {
		return nil, err
	}
	return &stickerType, nil
}

func (r *stickerRepositoryImpl) ListStickerTypes() ([]models.StickerType, error) {
	var stickers []models.StickerType
	err := r.db.
		Where("is_active = ?", true).
		Order("display_order ASC").
		Find(&stickers).Error

	return stickers, err
}

func (r *stickerRepositoryImpl) CreateStickerTransaction(tx *models.StickerTransaction) error {
	return r.db.Create(tx).Error
}

func (r *stickerRepositoryImpl) GetLeaderboard(filter LeaderboardFilter) ([]LeaderboardResult, error) {
	var results []LeaderboardResult
	query := r.db.
		Table("sticker_transactions as st").
		Select(`
			st.receiver_id as employee_id,
			e.full_name,
			d.name as department_name,
			count(*) as total
		`).
		Joins("JOIN employees as e ON e.id = st.receiver_id").
		Joins("LEFT JOIN departments as d ON d.id = e.department_id")

	// Filter by time range
	if filter.StartDate != nil {
		query = query.Where("st.created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("st.created_at <= ?", filter.EndDate.Add(24*time.Hour)) // include the end date
	}

	// Filter by department
	if filter.DepartmentID != nil {
		query = query.Where("e.department_id = ?", *filter.DepartmentID)
	}

	if filter.Limit <= 0 {
		filter.Limit = 10 // default limit
	}

	err := query.
		Group("st.receiver_id, e.full_name, d.name").
		Order("total DESC").
		Limit(filter.Limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}
	return results, nil
}
