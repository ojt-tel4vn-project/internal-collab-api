package middleware

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"go.uber.org/zap"
)

// Context key for roles
const RolesKey contextKey = "roles"

// Common role errors
var (
	ErrEmployeeNotFound        = errors.New("employee not found")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrProfileNotActive        = errors.New("profile not active, please complete setup first")
)

// CheckUserRole checks if employee has any of the required roles
func CheckUserRole(employeeID uuid.UUID, employeeRepo repository.EmployeeRepository, requiredRoles ...string) error {
	employee, err := employeeRepo.FindByID(employeeID)
	if err != nil {
		logger.Warn("Employee not found for role check", zap.String("employee_id", employeeID.String()))
		return ErrEmployeeNotFound
	}

	// Get employee single role name
	var roleName string
	if employee.Role != nil {
		roleName = employee.Role.Name
	}

	// Check if employee's role matches any of the required roles
	for _, requiredRole := range requiredRoles {
		if roleName == requiredRole {
			logger.Debug("Role check passed",
				zap.String("employee_id", employeeID.String()),
				zap.String("role", requiredRole),
			)
			return nil
		}
	}

	logger.Warn("Insufficient permissions",
		zap.String("employee_id", employeeID.String()),
		zap.Strings("required_roles", requiredRoles),
	)
	return ErrInsufficientPermissions
}

// CheckProfileStatus checks if employee profile is active
func CheckProfileStatus(employeeID uuid.UUID, employeeRepo repository.EmployeeRepository) error {
	employee, err := employeeRepo.FindByID(employeeID)
	if err != nil {
		logger.Warn("Employee not found for profile check", zap.String("employee_id", employeeID.String()))
		return ErrEmployeeNotFound
	}

	if employee.Status != models.StatusActive {
		logger.Warn("Employee profile not active",
			zap.String("employee_id", employeeID.String()),
			zap.String("status", string(employee.Status)),
		)
		return ErrProfileNotActive
	}

	logger.Debug("Profile status check passed",
		zap.String("employee_id", employeeID.String()),
	)
	return nil
}

// Helper function to get role from context
func GetRolesFromContext(ctx context.Context) (*models.Role, bool) {
	role, ok := ctx.Value(RolesKey).(*models.Role)
	return role, ok
}

// Helper function to check if user has specific role
func HasRole(ctx context.Context, roleName string) bool {
	role, ok := GetRolesFromContext(ctx)
	if !ok || role == nil {
		return false
	}
	return role.Name == roleName
}

// SetRolesInContext sets role in context
func SetRolesInContext(ctx context.Context, role *models.Role) context.Context {
	return context.WithValue(ctx, RolesKey, role)
}
