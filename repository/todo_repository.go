package repository

import (
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type TodoRepository interface {
	Create(todo *models.Todo) error
	FindAll() ([]models.Todo, error)
	FindByID(id uint) (*models.Todo, error)
	Update(todo *models.Todo) error
	Delete(id uint) error
}

type todoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) Create(todo *models.Todo) error {
	return r.db.Create(todo).Error
}

func (r *todoRepository) FindAll() ([]models.Todo, error) {
	var todos []models.Todo
	err := r.db.Order("created_at desc").Find(&todos).Error
	return todos, err
}

func (r *todoRepository) FindByID(id uint) (*models.Todo, error) {
	var todo models.Todo
	err := r.db.First(&todo, id).Error
	return &todo, err
}

func (r *todoRepository) Update(todo *models.Todo) error {
	return r.db.Save(todo).Error
}

func (r *todoRepository) Delete(id uint) error {
	return r.db.Delete(&models.Todo{}, id).Error
}
