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
	GetLeaderboard(filter repository.LeaderboardFilter) ([]sticker.LeaderboardResult, error)
	UpdateGlobalConfig(point, month, day int) error
	CreateSticker(req sticker.CreateStickerRequest) (*models.StickerType, error)
	ListStickerTypes() ([]sticker.StickerTypeResponse, error)
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

func (s *stickerServiceImpl) ListStickerTypes() ([]sticker.StickerTypeResponse, error) {
	stickers, err := s.repo.ListStickerTypes()
	if err != nil {
		return nil, err
	}

	// Convert models to DTOs
	result := make([]sticker.StickerTypeResponse, len(stickers))
	for i, st := range stickers {
		result[i] = sticker.StickerTypeResponse{
			ID:           st.ID,
			Name:         st.Name,
			Description:  st.Description,
			PointCost:    st.PointCost,
			Category:     st.Category,
			IconURL:      st.IconURL,
			IsActive:     st.IsActive,
			DisplayOrder: st.DisplayOrder,
			CreatedAt:    st.CreatedAt,
			UpdatedAt:    st.UpdatedAt,
		}
	}
	return result, nil
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

func (s *stickerServiceImpl) GetLeaderboard(filter repository.LeaderboardFilter) ([]sticker.LeaderboardResult, error) {
	results, err := s.repo.GetLeaderboard(filter)
	if err != nil {
		return nil, err
	}

	// Convert repository results to DTOs
	dtoResults := make([]sticker.LeaderboardResult, len(results))
	for i, r := range results {
		dtoResults[i] = sticker.LeaderboardResult{
			EmployeeID: r.EmployeeID,
			FullName:   r.FullName,
			Total:      r.Total,
		}
	}
	return dtoResults, nil
}

func (s *stickerServiceImpl) UpdateGlobalConfig(points, month, day int) error {
	newConfig := &models.PointConfig{
		YearlyPoints: points,
		ResetMonth:   month,
		ResetDay:     day,
	}
	return s.configRepo.UpdatePointConfig(newConfig)
}
