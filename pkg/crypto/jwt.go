package crypto

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"go.uber.org/zap"
)

var jwtSecret []byte

// Claims struct chứa thông tin trong JWT token
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	jwt.RegisteredClaims
}

// InitJWT khởi tạo JWT secret từ config (gọi khi start app)
func InitJWT(secret string) {
	if secret == "" {
		logger.Warn("JWT secret is empty, using default (NOT RECOMMENDED FOR PRODUCTION)")
		secret = "your-secret-key-change-this-in-production"
	}
	jwtSecret = []byte(secret)
	logger.Info("JWT initialized",
		zap.Int("secret_length", len(secret)),
	)
}

// GenerateToken tạo JWT token mới
func GenerateToken(userID uuid.UUID, username, email string, expirationHours int) (string, error) {
	if len(jwtSecret) == 0 {
		return "", errors.New("JWT secret not initialized")
	}

	expirationTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		logger.Error("Failed to generate JWT token",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("username", username),
		)
		return "", err
	}

	logger.Info("JWT token generated successfully",
		zap.String("user_id", userID.String()),
		zap.String("username", username),
		zap.Time("expires_at", expirationTime),
		zap.Int("expiration_hours", expirationHours),
	)

	return tokenString, nil
}

// ValidateToken xác thực và parse JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	if len(jwtSecret) == 0 {
		return nil, errors.New("JWT secret not initialized")
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// Kiểm tra signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Warn("Invalid signing method in JWT token",
				zap.String("method", token.Method.Alg()),
			)
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		logger.Warn("JWT token validation failed",
			zap.Error(err),
		)
		return nil, err
	}

	if !token.Valid {
		logger.Warn("Invalid JWT token")
		return nil, errors.New("invalid token")
	}

	logger.Debug("JWT token validated successfully",
		zap.String("user_id", claims.UserID.String()),
		zap.String("username", claims.Username),
		zap.Time("expires_at", claims.ExpiresAt.Time),
	)

	return claims, nil
}

// RefreshToken tạo token mới từ token cũ (nếu còn hợp lệ)
func RefreshToken(tokenString string, expirationHours int) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		logger.Error("Failed to refresh token: invalid token",
			zap.Error(err),
		)
		return "", err
	}

	logger.Info("Refreshing JWT token",
		zap.String("user_id", claims.UserID.String()),
		zap.String("username", claims.Username),
		zap.Int("new_expiration_hours", expirationHours),
	)

	// Tạo token mới với thông tin từ token cũ
	return GenerateToken(claims.UserID, claims.Username, claims.Email, expirationHours)
}
