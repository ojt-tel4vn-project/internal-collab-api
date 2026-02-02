package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseRepository defines the standard operations to be performed on data models.
type BaseRepository[T any] interface {
	Create(entity *T) error
	FindAll() ([]T, error)
	FindByID(id uuid.UUID) (*T, error)
	Update(entity *T) error
	Delete(id uuid.UUID) error
}

// baseRepository implements BaseRepository for a generic type T.
type baseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository creates a new instance of the generic base repository.
func NewBaseRepository[T any](db *gorm.DB) BaseRepository[T] {
	return &baseRepository[T]{db: db}
}

func (r *baseRepository[T]) Create(entity *T) error {
	return r.db.Create(entity).Error
}

func (r *baseRepository[T]) FindAll() ([]T, error) {
	var entities []T
	err := r.db.Find(&entities).Error
	return entities, err
}

func (r *baseRepository[T]) FindByID(id uuid.UUID) (*T, error) {
	var entity T
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *baseRepository[T]) Update(entity *T) error {
	return r.db.Save(entity).Error
}

func (r *baseRepository[T]) Delete(id uuid.UUID) error {
	var entity T
	return r.db.Delete(&entity, id).Error
}
