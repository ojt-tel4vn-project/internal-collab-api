package models

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID         uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Title      string `gorm:"type:varchar(255);not null"`
	CategoryID uuid.UUID   `gorm:"not null"`
	FilePath   string `gorm:"not null"`
	UploadedBy uuid.UUID   `gorm:"not null"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}
