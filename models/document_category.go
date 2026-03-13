package models

import "github.com/google/uuid"

type DocumentCategory struct {
	ID       uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name     string `gorm:"type:varchar(100);not null;unique"`
	ParentID *uuid.UUID  `gorm:"type:uuid;default:null"`
}
