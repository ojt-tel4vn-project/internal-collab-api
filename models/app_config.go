package models

import (
	"time"

	"github.com/google/uuid"
)

// AppConfig là bảng key-value lưu cấu hình hệ thống (thay thế in-memory config)
type AppConfig struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Key       string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
