package repository

import (
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type EmployeeRepository interface {
	BaseRepository[models.Employee]
	FindByEmail(email string) (*models.Employee, error)
}

type employeeRepository struct {
	BaseRepository[models.Employee]
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &employeeRepository{
		BaseRepository: NewBaseRepository[models.Employee](db),
		db:             db,
	}
}

func (r *employeeRepository) FindByID(id uuid.UUID) (*models.Employee, error) {
	var employee models.Employee
	err := r.db.Preload("Roles").Preload("Department").First(&employee, id).Error
	return &employee, err
}

func (r *employeeRepository) FindByEmail(email string) (*models.Employee, error) {
	var employee models.Employee
	err := r.db.Preload("Roles").Preload("Department").Where("email = ?", email).First(&employee).Error
	return &employee, err
}
