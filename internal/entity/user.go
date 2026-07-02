package entity

import "time"

type User struct {
	ID    string `gorm:"primary_key;size:64" json:"id"`
	Name  string `gorm:"size:128;not null" json:"name"`
	Email string `gorm:"size:255" json:"email"`

	DeepSeekAPIKey  string `gorm:"type:text"`
	DeepSeekBaseURL string `gorm:"size:512;not null;default:'https://api.deepseek.com'"`
	DeepSeekModel   string `gorm:"size:128;not null;default:'deepseek-chat'"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (User) TableName() string {
	return "users"
}
