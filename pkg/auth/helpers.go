package auth

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"go.uber.org/zap"
)

// ValidateJWT validates JWT token from Authorization header
func ValidateJWT(authHeader string, jwtService crypto.JWTService) (*crypto.Claims, error) {
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	// Check Bearer prefix
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("invalid authorization format, expected: Bearer <token>")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		logger.Warn("JWT validation failed", zap.Error(err))
		return nil, err
	}

	return claims, nil
}

// CheckRole checks if employee has any of the required roles
func CheckRole(employeeID uuid.UUID, employeeRepo repository.EmployeeRepository, requiredRoles ...string) error {
	employee, err := employeeRepo.FindByID(employeeID)
	if err != nil {
		logger.Warn("Employee not found for role check", zap.String("employee_id", employeeID.String()))
		return errors.New("employee not found")
	}

	var roleName string
	if employee.Role != nil {
		roleName = employee.Role.Name
	}

	for _, required := range requiredRoles {
		if roleName == required {
			logger.Debug("Role check passed",
				zap.String("employee_id", employeeID.String()),
				zap.String("role", roleName),
			)
			return nil
		}
	}

	logger.Warn("Insufficient permissions",
		zap.String("employee_id", employeeID.String()),
		zap.Strings("required_roles", requiredRoles),
	)
	return errors.New("insufficient permissions")
}

// CheckProfileActive checks if employee profile is active
func CheckProfileActive(employeeID uuid.UUID, employeeRepo repository.EmployeeRepository) error {
	employee, err := employeeRepo.FindByID(employeeID)
	if err != nil {
		return errors.New("employee not found")
	}

	if employee.Status != "active" {
		return errors.New("profile not active, please complete setup first")
	}

	return nil
}
