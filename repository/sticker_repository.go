package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StickerRepository interface {
	// transaction support
	WithTransaction(tx *gorm.DB) StickerRepository

	// points
	GetPointBalance(employeeID uuid.UUID, year int) (*models.PointBalance, error)
	UpdatePointBalance(balance *models.PointBalance) error

	// sticker
	GetStickerType(stickerTypeID uuid.UUID) (*models.StickerType, error)

	// transaction
	CreateStickerTransaction(tx *models.StickerTransaction) error

	// leaderboard
	GetLeaderboard(filter LeaderboardFilter) ([]LeaderboardResult, error)
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

func (r *stickerRepositoryImpl) GetPointBalance(employeeID uuid.UUID, year int) (*models.PointBalance, error) {
	var balance models.PointBalance
	err := r.db.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("employee_id = ? AND year = ?", employeeID, year).
		First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

func (r *stickerRepositoryImpl) UpdatePointBalance(balance *models.PointBalance) error {
	return r.db.Save(balance).Error
}

func (r *stickerRepositoryImpl) GetStickerType(stickerTypeID uuid.UUID) (*models.StickerType, error) {
	var stickerType models.StickerType
	err := r.db.First(&stickerType, "id = ?", stickerTypeID).Error
	if err != nil {
		return nil, err
	}
	return &stickerType, nil
}

func (r *stickerRepositoryImpl) CreateStickerTransaction(tx *models.StickerTransaction) error {
	return r.db.Create(tx).Error
}

func (r *stickerRepositoryImpl) GetLeaderboard(filter LeaderboardFilter) ([]LeaderboardResult, error) {
	var results []LeaderboardResult
	query := r.db.
		Table("sticker_transactions st").
		Select(`
			st.receiver_id as employee_id,
			e.full_name,
			count(*) as total
		`).
		Joins("JOIN employees e ON e.id = st.receiver_id")

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
		Group("st.receiver_id, e.full_name").
		Order("total DESC").
		Limit(filter.Limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}
	return results, nil
}
