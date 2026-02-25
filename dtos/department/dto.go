package department

import (
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
)

type CreateDepartmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (r *CreateDepartmentRequest) ToModel() *models.Department {
	return &models.Department{
		Name:        r.Name,
		Description: r.Description,
	}
}

type UpdateDepartmentRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *UpdateDepartmentRequest) ToModel(id uuid.UUID) *models.Department {
	return &models.Department{
		ID:          id,
		Name:        r.Name,
		Description: r.Description,
	}
}

type DepartmentResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func FromModel(dept *models.Department) *DepartmentResponse {
	return &DepartmentResponse{
		ID:          dept.ID,
		Name:        dept.Name,
		Description: dept.Description,
		CreatedAt:   dept.CreatedAt,
		UpdatedAt:   dept.UpdatedAt,
	}
}
