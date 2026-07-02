package entity

import "time"

type Paper struct {
	ID string `gorm:"primaryKey;size:64"`

	UserID string `gorm:"size:64;index;not null"`

	Title       string `gorm:"size:512;not null"`
	Filename    string `gorm:"size:512;not null"`
	ContentType string `gorm:"size:128"`

	Size      int64  `gorm:"not null"`
	SHA256    string `gorm:"size:64;uniqueIndex;not null"`
	ObjectKey string `gorm:"size:1024;not null"`

	Status    string `gorm:"size:32;index;not null"`
	PageCount int
	Error     string `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Paper) TableName() string {
	return "papers"
}
