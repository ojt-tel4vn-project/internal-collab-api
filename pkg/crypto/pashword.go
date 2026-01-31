package crypto

import (
	"errors"

	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	if password == "" {
		logger.Error("Hash password failed: empty password provided")
		return "", errors.New("Password should not be empty")
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password",
			zap.Error(err),
			zap.Int("cost", bcrypt.DefaultCost))
		return "", err
	}

	logger.Debug("Password hashed successfully",
		zap.Int("cost", bcrypt.DefaultCost),
		zap.Int("hash_length", len(hashPassword)),
	)

	return string(hashPassword), nil
}

func VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			logger.Warn("Password comparison failed: password mismatch")
		} else {
			logger.Error("Password comparison faild",
				zap.Error(err))
		}
		return err
	}

	return nil
}
