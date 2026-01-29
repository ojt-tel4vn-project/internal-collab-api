package services

import (
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
)

type TodoService interface {
	CreateTodo(req dtos.CreateTodoRequest) (*models.Todo, error)
	GetAllTodos() ([]models.Todo, error)
	GetTodoByID(id uint) (*models.Todo, error)
	UpdateTodo(id uint, req dtos.UpdateTodoRequest) (*models.Todo, error)
	DeleteTodo(id uint) error
}

type todoService struct {
	repo repository.TodoRepository
}

func NewTodoService(repo repository.TodoRepository) TodoService {
	return &todoService{repo: repo}
}

func (s *todoService) CreateTodo(req dtos.CreateTodoRequest) (*models.Todo, error) {
	todo := &models.Todo{
		Title:       req.Title,
		Description: req.Description,
	}
	err := s.repo.Create(todo)
	return todo, err
}

func (s *todoService) GetAllTodos() ([]models.Todo, error) {
	return s.repo.FindAll()
}

func (s *todoService) GetTodoByID(id uint) (*models.Todo, error) {
	return s.repo.FindByID(id)
}

func (s *todoService) UpdateTodo(id uint, req dtos.UpdateTodoRequest) (*models.Todo, error) {
	todo, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		todo.Title = req.Title
	}
	if req.Description != "" {
		todo.Description = req.Description
	}
	if req.Completed != nil {
		todo.Completed = *req.Completed
	}

	err = s.repo.Update(todo)
	return todo, err
}

func (s *todoService) DeleteTodo(id uint) error {
	return s.repo.Delete(id)
}
