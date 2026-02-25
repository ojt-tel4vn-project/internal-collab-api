package document

import "github.com/google/uuid"

type CreateDocumentCategoryRequest struct {
	Name     string `json:"name" required:"true"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}