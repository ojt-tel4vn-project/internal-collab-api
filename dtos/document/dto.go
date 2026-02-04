package document

import (
	"mime/multipart"

	"github.com/google/uuid"
)

type CreateDocumentRequest struct {
	Title      string                `json:"title" form:"title" required:"true"`
	CategoryID uuid.UUID             `json:"category_id" form:"category_id" required:"true"`
	File       *multipart.FileHeader `json:"file" form:"file" required:"true"`
}
