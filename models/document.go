package models

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	FileName    string    `gorm:"type:varchar(255);not null" json:"file_name"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	Title       string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	CategoryID  uuid.UUID `gorm:"type:uuid;not null"`
	Roles       string    `gorm:"type:varchar(100);not null;default:'employee'"`
	FilePath    string    `gorm:"not null"`
	UploadedBy  uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}
