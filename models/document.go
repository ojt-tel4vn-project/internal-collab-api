package models

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	CategoryID  uuid.UUID `gorm:"type:uuid;not null" json:"category_id"`
	FileName    string    `gorm:"type:varchar(255);not null" json:"file_name"`
	FilePath    string    `gorm:"type:varchar(500);not null" json:"file_path"`
	FileSize    int64     `gorm:"type:bigint" json:"file_size"`
	MimeType    string    `gorm:"type:varchar(100)" json:"mime_type"`
	Roles       string    `gorm:"type:varchar(100);not null;default:'employee'" json:"roles"`
	UploadedBy  uuid.UUID `gorm:"type:uuid;not null" json:"uploaded_by"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
