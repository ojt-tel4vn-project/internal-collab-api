package middleware

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"go.uber.org/zap"
)

// Context keys for storing user information
type contextKey string

const (
	EmployeeIDKey contextKey = "employee_id"
	EmailKey      contextKey = "email"
	UsernameKey   contextKey = "username"
)

// ValidateJWTFromHeader validates JWT token from Authorization header
// This is a helper function to be used in handlers, not a traditional middleware
func ValidateJWTFromHeader(authHeader string, jwtService crypto.JWTService) (*crypto.Claims, error) {
	if authHeader == "" {
		logger.Warn("Missing Authorization header")
		return nil, crypto.ErrMissingAuthHeader
	}

	// Check Bearer prefix
	if !strings.HasPrefix(authHeader, "Bearer ") {
		logger.Warn("Invalid Authorization header format", zap.String("header", authHeader))
		return nil, crypto.ErrInvalidAuthFormat
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate token
	claims, err := jwtService.ValidateToken(tokenString)
	if err != nil {
		logger.Warn("Invalid or expired JWT token", zap.Error(err))
		return nil, err
	}

	logger.Debug("JWT authentication successful",
		zap.String("employee_id", claims.UserID.String()),
		zap.String("email", claims.Email),
	)

	return claims, nil
}

// Helper functions to extract values from context

// GetEmployeeIDFromContext extracts employee ID from context
func GetEmployeeIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	employeeID, ok := ctx.Value(EmployeeIDKey).(uuid.UUID)
	return employeeID, ok
}

// GetEmailFromContext extracts email from context
func GetEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(EmailKey).(string)
	return email, ok
}

// GetUsernameFromContext extracts username from context
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(UsernameKey).(string)
	return username, ok
}

// SetUserInfoInContext sets user information in context
func SetUserInfoInContext(ctx context.Context, claims *crypto.Claims) context.Context {
	ctx = context.WithValue(ctx, EmployeeIDKey, claims.UserID)
	ctx = context.WithValue(ctx, EmailKey, claims.Email)
	ctx = context.WithValue(ctx, UsernameKey, claims.Username)
	return ctx
}
