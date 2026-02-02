package crypto

import "github.com/google/uuid"

type jwtServiceImpl struct{}

func NewJWTService() JWTService {
	return &jwtServiceImpl{}
}

func (s *jwtServiceImpl) GenerateToken(userID uuid.UUID, username, email string, expirationHours int) (string, error) {
	return GenerateToken(userID, username, email, expirationHours)
}

func (s *jwtServiceImpl) ValidateToken(tokenString string) (*Claims, error) {
	return ValidateToken(tokenString)
}

func (s *jwtServiceImpl) RefreshToken(tokenString string, expirationHours int) (string, error) {
	return RefreshToken(tokenString, expirationHours)
}

type passwordServiceImpl struct{}

func NewPasswordService() PasswordService {
	return &passwordServiceImpl{}
}

func (s *passwordServiceImpl) HashPassword(password string) (string, error) {
	return HashPassword(password)
}

func (s *passwordServiceImpl) VerifyPassword(hashedPassword, password string) error {
	return VerifyPassword(hashedPassword, password)
}
