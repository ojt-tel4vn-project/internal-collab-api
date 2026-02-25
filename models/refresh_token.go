package models

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"type:varchar(255);not null;unique;index" json:"token"`
	ExpiresAt time.Time `gorm:"type:timestamp;not null" json:"expires_at"`
	Revoked   bool      `gorm:"type:boolean;default:false;not null" json:"revoked"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	User      Employee  `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
