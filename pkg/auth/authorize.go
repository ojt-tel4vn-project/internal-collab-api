package auth

import (
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
)

type AuthOptions struct {
	Roles         []string
	RequireActive bool
}

func Authorize(
	authHeader string,
	jwtService crypto.JWTService,
	employeeRepo repository.EmployeeRepository,
	opts AuthOptions,
) (*crypto.Claims, error) {

	claims, err := ValidateJWT(authHeader, jwtService)
	if err != nil {
		return nil, err
	}

	if len(opts.Roles) > 0 {
		if err := CheckRole(claims.UserID, employeeRepo, opts.Roles...); err != nil {
			return nil, err
		}
	}

	if opts.RequireActive {
		if err := CheckProfileActive(claims.UserID, employeeRepo); err != nil {
			return nil, err
		}
	}

	return claims, nil
}
