package models

import (
	"time"

	"github.com/google/uuid"
)

// Comment represents a comment on a document
type Comment struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	DocumentID uuid.UUID  `gorm:"type:uuid;not null;index" json:"document_id"`
	AuthorID   uuid.UUID  `gorm:"type:uuid;not null" json:"author_id"`
	Author     *Employee  `gorm:"foreignKey:AuthorID;references:ID" json:"author,omitempty"`
	Content    string     `gorm:"type:text;not null" json:"content"`
	ParentID   *uuid.UUID `gorm:"type:uuid" json:"parent_id,omitempty"` // for threaded replies
	Parent     *Comment   `gorm:"foreignKey:ParentID;references:ID" json:"parent,omitempty"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}
