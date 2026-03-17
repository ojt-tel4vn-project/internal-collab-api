package models

import (
	"time"

	"github.com/google/uuid"
)

type PointConfig struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"point_config_id"`
	YearlyPoints int       `gorm:"not null" json:"yearly_points"`
	ResetMonth   int       `gorm:"not null" json:"reset_month"`
	ResetDay     int       `gorm:"not null" json:"reset_day"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
