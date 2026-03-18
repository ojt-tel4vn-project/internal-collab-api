package services

import (
	"errors"

	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"go.uber.org/zap"
)

var (
	ErrDepartmentExists = errors.New("department name already exists")
)

type DepartmentService interface {
	GetAllDepartments() ([]models.Department, error)
	CreateDepartment(name, description string) (*models.Department, error)
}

type departmentService struct {
	repo repository.DepartmentRepository
}

func NewDepartmentService(repo repository.DepartmentRepository) DepartmentService {
	return &departmentService{repo: repo}
}

func (s *departmentService) GetAllDepartments() ([]models.Department, error) {
	return s.repo.FindAll()
}

func (s *departmentService) CreateDepartment(name, description string) (*models.Department, error) {
	// Check if already exists
	if existing, _ := s.repo.FindByName(name); existing != nil {
		return nil, ErrDepartmentExists
	}

	dept := &models.Department{
		Name:        name,
		Description: description,
	}

	if err := s.repo.Create(dept); err != nil {
		logger.Error("Failed to create department", zap.Error(err), zap.String("name", name))
		return nil, err
	}

	return dept, nil
}
