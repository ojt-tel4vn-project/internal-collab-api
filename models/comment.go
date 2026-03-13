package models

import (
	"time"

	"github.com/google/uuid"
)

// Comment represents an employee's opinion/dispute about their own attendance record
type Comment struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	AttendanceID uuid.UUID   `gorm:"type:uuid;not null;index" json:"attendance_id"`
	Attendance   *Attendance `gorm:"foreignKey:AttendanceID;references:ID" json:"attendance,omitempty"`
	AuthorID     uuid.UUID   `gorm:"type:uuid;not null" json:"author_id"`
	Author       *Employee   `gorm:"foreignKey:AuthorID;references:ID" json:"author,omitempty"`
	Content      string      `gorm:"type:text;not null" json:"content"`
	IsRead       bool        `gorm:"type:boolean;default:false" json:"is_read"`
	ParentID     *uuid.UUID  `gorm:"type:uuid" json:"parent_id,omitempty"` // for threaded replies
	Parent       *Comment    `gorm:"foreignKey:ParentID;references:ID" json:"parent,omitempty"`
	CreatedAt    time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
}
