package services

import (
	"errors"

	"github.com/google/uuid"
	models "github.com/ojt-tel4vn-project/internal-collab-api/models/document"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"gorm.io/gorm"
)

type DocumentCategoryService interface {
	Create(name string, parentID *uuid.UUID) (*models.DocumentCategory, error)
}

type documentCategoryServiceImpl struct {
	repo repository.DocumentCategoryRepository
}

func NewDocumentCategoryService(repo repository.DocumentCategoryRepository) DocumentCategoryService {
	return &documentCategoryServiceImpl{repo: repo}
}

func (s *documentCategoryServiceImpl) Create(name string, parentID *uuid.UUID) (*models.DocumentCategory, error) {
	if _, err := s.repo.FindByName(name); err == nil {
		return nil, errors.New("Document category already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if parentID != nil {
		if _, err := s.repo.FindByID(*parentID); err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.New("Parent category not found")
			}
			return nil, err
		}
	}

	category := &models.DocumentCategory{
		Name:     name,
		ParentID: parentID,
	}	
	if err := s.repo.Create(category); err != nil {
		return nil, err
	}
	return category, nil
}
