package role

import (
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
)

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (r *CreateRoleRequest) ToModel() *models.Role {
	return &models.Role{
		Name:        r.Name,
		Description: r.Description,
	}
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *UpdateRoleRequest) ToModel(id uuid.UUID) *models.Role {
	return &models.Role{
		ID:          id,
		Name:        r.Name,
		Description: r.Description,
	}
}

type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func FromModel(role *models.Role) *RoleResponse {
	return &RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
	}
}
