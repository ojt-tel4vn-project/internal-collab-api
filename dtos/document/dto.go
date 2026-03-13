package document

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

// CreateDocumentRequest: Payload from HR
type CreateDocumentRequest struct {
	Title       string                `json:"title" form:"title" required:"true"`
	Description string                `json:"description" form:"description"`
	CategoryID  uuid.UUID             `json:"category_id" form:"category_id" required:"true"`
	File        *multipart.FileHeader `json:"file" form:"file" required:"true"`
	Roles       string                `json:"roles" form:"roles" required:"true" doc:"Ex: employee,manager,hr"`
}

// DocumentResponse is the returned object when listing Documents
type DocumentResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CategoryID  uuid.UUID `json:"category_id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	Roles       string    `json:"roles"`
	UploadedBy  uuid.UUID `json:"uploaded_by"`
	IsRead      bool      `json:"is_read"` // For mapping in lists
	CreatedAt   time.Time `json:"created_at"`
}

type ListDocumentResponse struct {
	Body struct {
		Data []DocumentResponse `json:"data"`
	}
}

type SingleDocumentResponse struct {
	Body struct {
		Data DocumentResponse `json:"data"`
	}
}
