package entity

import (
	"time"

	"gorm.io/datatypes"
)

type Task struct {
	ID string `gorm:"primaryKey;size:96"`

	UserID  string `gorm:"size:64;index;not null"`
	PaperID string `gorm:"size:64;index"`

	Type   string `gorm:"size:64;index;not null"`
	Status string `gorm:"size:32;index;not null"`

	Error    string            `gorm:"type:text"`
	Metadata datatypes.JSONMap `gorm:"type:json"`

	StartedAt  *time.Time
	FinishedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Task) TableName() string {
	return "tasks"
}
