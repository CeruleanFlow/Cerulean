package entity

import (
	"time"

	"gorm.io/datatypes"
)

type Chunk struct {
	ID string `gorm:"primaryKey;size:96"`

	UserID  string `gorm:"size:64;index;not null"`
	PaperID string `gorm:"size:64;index;not null"`

	PageNo     int `gorm:"index"`
	ChunkIndex int `gorm:"column:chunk_index;index"`

	Text      string `gorm:"type:longtext;not null"`
	ObjectKey string `gorm:"size:1024"`
	VectorID  string `gorm:"size:128;index"`

	Metadata datatypes.JSONMap `gorm:"type:json"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Chunk) TableName() string {
	return "chunks"
}
