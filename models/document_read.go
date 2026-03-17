package models

import (
	"time"

	"github.com/google/uuid"
)

type DocumentRead struct {
	DocumentID uuid.UUID `gorm:"type:uuid;primaryKey"`
	EmployeeID uuid.UUID `gorm:"type:uuid;primaryKey"`
	ReadAt     time.Time `gorm:"autoCreateTime"`
}
