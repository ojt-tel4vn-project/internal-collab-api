package models

import (
	"time"

	"github.com/google/uuid"
)

type DocumentRead struct {
	DocumentID uuid.UUID `gorm:"primaryKey"`
	EmployeeID uuid.UUID `gorm:"primaryKey"`
	ReadAt     time.Time
}
