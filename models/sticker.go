package models

import (
	"time"

	"github.com/google/uuid"
)

// PointBalance quản lý điểm thưởng theo năm
type PointBalance struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"point_balance_id"`
	EmployeeID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_employee_year" json:"employee_id"`
	Year          int       `gorm:"not null;uniqueIndex:idx_employee_year" json:"year"`
	InitialPoints int       `gorm:"not null" json:"initial_points"`
	CurrentPoints int       `gorm:"not null" json:"current_points"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	Employee Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

type StickerType struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"sticker_type_id"`
	Name         string    `gorm:"type:varchar(100);not null" json:"name"`
	Description  string    `gorm:"type:varchar(255)" json:"description"`
	PointCost    int       `gorm:"not null" json:"point_cost"`
	Category     string    `gorm:"type:varchar(50);not null;index" json:"category"`
	IconURL      string    `gorm:"type:varchar(255);not null" json:"icon_url"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	DisplayOrder int       `gorm:"default:0" json:"display_order"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type StickerTransaction struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"sticker_transaction_id"`
	SenderID      uuid.UUID `gorm:"type:uuid;not null;index" json:"sender_id"`
	ReceiverID    uuid.UUID `gorm:"type:uuid;not null;index" json:"receiver_id"`
	StickerTypeID uuid.UUID `gorm:"type:uuid;not null;index" json:"sticker_type_id"`
	PointCost     int       `gorm:"not null" json:"point_cost"`
	Message       string    `gorm:"type:varchar(255)" json:"message"`
	CreatedAt     time.Time `gorm:"autoCreateTime;index"`

	Sender      Employee    `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Receiver    Employee    `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
	StickerType StickerType `gorm:"foreignKey:StickerTypeID" json:"sticker_type,omitempty"`
}
