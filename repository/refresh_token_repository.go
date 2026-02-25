package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	FindByToken(token string) (*models.RefreshToken, error)
	Revoke(token string) error
	RevokeAllUserTokens(userID uuid.UUID) error
}

type refreshTokenRepositoryImpl struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepositoryImpl{db: db}
}

func (r *refreshTokenRepositoryImpl) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepositoryImpl) FindByToken(token string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.Where("token = ? AND revoked = ?", token, false).First(&rt).Error
	if err != nil {
		return nil, err
	}
	if rt.ExpiresAt.Before(time.Now()) {
		return nil, gorm.ErrRecordNotFound
	}
	return &rt, nil
}

func (r *refreshTokenRepositoryImpl) Revoke(token string) error {
	return r.db.Model(&models.RefreshToken{}).Where("token = ?", token).Update("revoked", true).Error
}

func (r *refreshTokenRepositoryImpl) RevokeAllUserTokens(userID uuid.UUID) error {
	return r.db.Model(&models.RefreshToken{}).Where("user_id = ?", userID).Update("revoked", true).Error
}
