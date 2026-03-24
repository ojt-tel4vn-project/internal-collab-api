package repository

import (
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type DepartmentRepository interface {
	Create(department *models.Department) error
	FindAll() ([]models.Department, error)
	FindByID(id uuid.UUID) (*models.Department, error)
	FindByName(name string) (*models.Department, error)
}

type departmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) Create(department *models.Department) error {
	return r.db.Create(department).Error
}

func (r *departmentRepository) FindAll() ([]models.Department, error) {
	var depts []models.Department
	err := r.db.Order("name ASC").Find(&depts).Error
	return depts, err
}

func (r *departmentRepository) FindByID(id uuid.UUID) (*models.Department, error) {
	var dept models.Department
	err := r.db.Where("id = ?", id).First(&dept).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found, return nil without error
		}
		return nil, err // Other errors
	}
	return &dept, nil
}

func (r *departmentRepository) FindByName(name string) (*models.Department, error) {
	var dept models.Department
	err := r.db.Where("name = ?", name).First(&dept).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found, return nil without error
		}
		return nil, err // Other errors
	}
	return &dept, nil
}
