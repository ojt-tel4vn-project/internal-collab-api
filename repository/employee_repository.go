package repository

import (
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type EmployeeRepository interface {
	BaseRepository[models.Employee]
	FindByEmail(email string) (*models.Employee, error)
	FindByEmployeeCode(code string) (*models.Employee, error)
	FindByPasswordResetToken(token string) (*models.Employee, error)
	FindEmployeesByBirthday(month, day int) ([]models.Employee, error)
	FindSubordinates(managerID uuid.UUID) ([]models.Employee, error)
	FindAllBirthdays() ([]models.Employee, error)
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

// FindAll overrides BaseRepository to preload Role and Department
func (r *employeeRepository) FindAll() ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.Preload("Role").Preload("Department").Find(&employees).Error
	return employees, err
}

func (r *employeeRepository) FindByID(id uuid.UUID) (*models.Employee, error) {
	var employee models.Employee
	err := r.db.Preload("Role").Preload("Department").First(&employee, id).Error
	return &employee, err
}

func (r *employeeRepository) FindByEmail(email string) (*models.Employee, error) {
	var employee models.Employee
	err := r.db.Preload("Role").Preload("Department").Where("email = ?", email).First(&employee).Error
	return &employee, err
}

func (r *employeeRepository) FindByEmployeeCode(code string) (*models.Employee, error) {
	var employee models.Employee
	err := r.db.Preload("Role").Preload("Department").Where("employee_code = ?", code).First(&employee).Error
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

func (r *employeeRepository) FindByPasswordResetToken(token string) (*models.Employee, error) {
	var employee models.Employee
	err := r.db.Preload("Role").Preload("Department").Where("password_reset_token = ?", token).First(&employee).Error
	return &employee, err
}

func (r *employeeRepository) FindEmployeesByBirthday(month, day int) ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.Preload("Department").
		Where("EXTRACT(MONTH FROM date_of_birth) = ? AND EXTRACT(DAY FROM date_of_birth) = ?", month, day).
		Find(&employees).Error
	return employees, err
}

// FindAllBirthdays returns all active employees sorted by birth month/day (for calendar use)
func (r *employeeRepository) FindAllBirthdays() ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.Preload("Department").
		Where("status = 'active'").
		Order("EXTRACT(MONTH FROM date_of_birth), EXTRACT(DAY FROM date_of_birth)").
		Find(&employees).Error
	return employees, err
}

func (r *employeeRepository) FindSubordinates(managerID uuid.UUID) ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.Preload("Department").Where("manager_id = ?", managerID).Find(&employees).Error
	return employees, err
}
