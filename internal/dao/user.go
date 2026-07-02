package dao

import (
	"context"
	"errors"
	"time"

	"github.com/CeruleanFlow/cerulean/internal/entity"
	"gorm.io/gorm"
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (d *UserDAO) EnsureDefaultUser(ctx context.Context) error {
	var user entity.User

	err := d.db.WithContext(ctx).
		First(&user, "id = ?", DefaultUserID).
		Error

	if err == nil {
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	now := time.Now()

	user = entity.User{
		ID:              DefaultUserID,
		Name:            "default",
		Email:           "",
		DeepSeekBaseURL: "https://api.deepseek.com",
		DeepSeekModel:   "deepseek-chat",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	return d.db.WithContext(ctx).Create(&user).Error
}

func (d *UserDAO) GetDefault(ctx context.Context) (entity.User, error) {
	var user entity.User

	err := d.db.WithContext(ctx).
		First(&user, "id = ?", DefaultUserID).
		Error

	return user, err
}

func (d *UserDAO) UpdateDeepSeekConfig(ctx context.Context, apiKey string, baseURL string, model string) (entity.User, error) {
	var user entity.User

	err := d.db.WithContext(ctx).
		First(&user, "id = ?", DefaultUserID).
		Error

	if err != nil {
		return entity.User{}, err
	}

	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}

	if model == "" {
		model = "deepseek-chat"
	}

	user.DeepSeekAPIKey = apiKey
	user.DeepSeekBaseURL = baseURL
	user.DeepSeekModel = model
	user.UpdatedAt = time.Now()

	err = d.db.WithContext(ctx).Save(&user).Error
	return user, err
}
