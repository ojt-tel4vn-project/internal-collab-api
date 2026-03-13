package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"gorm.io/gorm"
)

type StickerService interface {
	SendSticker(senderID, receiverID, stickerTypeID uuid.UUID, message string) error
	GetPointBalance(employeeID uuid.UUID) (*models.PointBalance, error)
	GetLeaderboard(filter repository.LeaderboardFilter) ([]repository.LeaderboardResult, error)
	UpdateGlobalConfig(points, month, day int) error
}

type stickerServiceImpl struct {
	repo       repository.StickerRepository
	configRepo repository.PointConfigRepository
	db         *gorm.DB
}

func NewStickerService(repo repository.StickerRepository, configRepo repository.PointConfigRepository, db *gorm.DB) StickerService {
	return &stickerServiceImpl{
		repo:       repo,
		configRepo: configRepo,
		db:         db,
	}
}

func (s *stickerServiceImpl) SendSticker(senderID, receiverID, stickerTypeID uuid.UUID, message string) error {
	if senderID == receiverID {
		return errors.New("cannot send sticker to yourself")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Use transaction-scoped repository
		txRepo := s.repo.WithTransaction(tx)

		// Get sticker type
		stickerType, err := txRepo.GetStickerType(stickerTypeID)
		if err != nil {
			return errors.New("sticker type not found")
		}

		// Get sender's point balance for current year
		year := time.Now().Year()
		senderBalance, err := txRepo.GetPointBalance(senderID, year)
		if err != nil {
			return errors.New("sender point balance not found")
		}

		// Check if sender has enough points
		if senderBalance.CurrentPoints < stickerType.PointCost {
			return errors.New("not enough points to send sticker")
		}

		// Deduct points from sender
		senderBalance.CurrentPoints -= stickerType.PointCost
		if err := txRepo.UpdatePointBalance(senderBalance); err != nil {
			return err
		}

		// Create sticker transaction (receiver gets sticker but NOT points)
		transaction := &models.StickerTransaction{
			SenderID:      senderID,
			ReceiverID:    receiverID,
			StickerTypeID: stickerTypeID,
			PointCost:     stickerType.PointCost,
			Message:       message,
		}
		if err := txRepo.CreateStickerTransaction(transaction); err != nil {
			return err
		}
		return nil
	})
}

func (s *stickerServiceImpl) GetPointBalance(employeeID uuid.UUID) (*models.PointBalance, error) {
	year := time.Now().Year()
	balance, err := s.repo.GetPointBalance(employeeID, year)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Lấy cấu hình điểm
			config, configErr := s.configRepo.GetPointConfig()
			initialPoints := 500 // Mặc định nếu không có cấu hình
			if configErr == nil && config != nil {
				initialPoints = config.YearlyPoints
			}
			// Tạo bản ghi mới
			newBalance := &models.PointBalance{
				EmployeeID:    employeeID,
				Year:          year,
				InitialPoints: initialPoints,
				CurrentPoints: initialPoints,
			}
			if err := s.repo.UpdatePointBalance(newBalance); err != nil {
				return nil, err
			}
			return newBalance, nil
		}
		return nil, err
	}
	return balance, nil
}

func (s *stickerServiceImpl) GetLeaderboard(filter repository.LeaderboardFilter) ([]repository.LeaderboardResult, error) {
	return s.repo.GetLeaderboard(filter)
}

func (s *stickerServiceImpl) UpdateGlobalConfig(points, month, day int) error {
	newConfig := &models.PointConfig{
		YearlyPoints: points,
		ResetMonth:   month,
		ResetDay:     day,
	}
	return s.configRepo.UpdatePointConfig(newConfig)
}
