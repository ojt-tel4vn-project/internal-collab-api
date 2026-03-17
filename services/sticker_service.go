package services

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos/sticker"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"gorm.io/gorm"
)

type StickerService interface {
	SendSticker(senderID, receiverID, stickerTypeID uuid.UUID, message string) error
	GetPointBalance(employeeID uuid.UUID) (*models.PointBalance, error)
	GetLeaderboard(filter repository.LeaderboardFilter) ([]repository.LeaderboardResult, error)
	UpdateGlobalConfig(point, month, day int) error
	CreateSticker(req sticker.CreateStickerRequest) (*models.StickerType, error)
	ListStickerTypes() ([]models.StickerType, error)
}

type stickerServiceImpl struct {
	repo       repository.StickerRepository
	configRepo repository.PointConfigRepository
	db         *gorm.DB
}

var (
	ErrSendToYourself  = errors.New("cannot send sticker to yourself")
	ErrStickerNotFound = errors.New("sticker type not found")
	ErrPointNotFound   = errors.New("sender point balance not found")
	ErrNotEnoughPoints = errors.New("not enough points")
	ErrStickerInactive = errors.New("sticker type is inactive")
)

func NewStickerService(repo repository.StickerRepository, configRepo repository.PointConfigRepository, db *gorm.DB) StickerService {
	return &stickerServiceImpl{
		repo:       repo,
		configRepo: configRepo,
		db:         db,
	}
}

func (s *stickerServiceImpl) CreateSticker(req sticker.CreateStickerRequest) (*models.StickerType, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("sticker name cannot be empty")
	}

	if req.PointCost <= 0 {
		return nil, errors.New("point cost must be greater than zero")
	}

	newSticker := &models.StickerType{
		Name:         req.Name,
		Description:  req.Description,
		PointCost:    req.PointCost,
		Category:     req.Category,
		IconURL:      req.IconURL,
		DisplayOrder: req.DisplayOrder,
		IsActive:     true,
	}
	if err := s.repo.CreateSticker(newSticker); err != nil {
		return nil, err
	}
	return newSticker, nil

}

func (s *stickerServiceImpl) SendSticker(senderID, receiverID, stickerTypeID uuid.UUID, message string) error {
	if len(message) > 500 {
		return errors.New("message too long")
	}
	if senderID == receiverID {
		return ErrSendToYourself
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Use transaction-scoped repository
		txRepo := s.repo.WithTransaction(tx)

		// Get sticker type
		stickerType, err := txRepo.GetStickerTypeByID(stickerTypeID)
		if err != nil {
			return ErrStickerNotFound
		}

		if !stickerType.IsActive {
			return ErrStickerInactive
		}

		// Get sender's point balance for current year
		year := time.Now().Year()
		senderBalance, err := txRepo.GetPointBalance(senderID, year)
		if err != nil {
			return ErrPointNotFound
		}

		// Check if sender has enough points
		if senderBalance.CurrentPoints < stickerType.PointCost {
			return ErrNotEnoughPoints
		}

		// Create sticker transaction - points will be deducted by database trigger automatically
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

func (s *stickerServiceImpl) ListStickerTypes() ([]models.StickerType, error) {
	return s.repo.ListStickerTypes()
}

func (s *stickerServiceImpl) GetPointBalance(employeeID uuid.UUID) (*models.PointBalance, error) {
	year := time.Now().Year()
	balance, err := s.repo.GetPointBalance(employeeID, year)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			config, configErr := s.configRepo.GetPointConfig()
			initialPoints := 500
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
			if err := s.repo.CreatePointBalance(newBalance); err != nil {
				existing, getErr := s.repo.GetPointBalance(employeeID, year)
				if getErr != nil {
					return nil, getErr
				}
				return existing, nil
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

func (s *stickerServiceImpl) UpdateGlobalConfig(point, month, day int) error {
	newConfig := &models.PointConfig{
		YearlyPoints: point,
		ResetMonth:   month,
		ResetDay:     day,
	}
	if point <= 0 {
		return errors.New("Yearly point must be greater than zero")
	}
	if month < 1 || month > 12 {
		return errors.New("month must be between 1 and 12")
	}
	if day < 1 || day > 31 {
		return errors.New("day must be between 1 and 31")
	}
	return s.configRepo.UpdatePointConfig(newConfig)
}
