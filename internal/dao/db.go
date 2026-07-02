package dao

import (
	"context"
	"time"

	"github.com/CeruleanFlow/cerulean/internal/entity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const DefaultUserID = "user_default"

type Database struct {
	DB *gorm.DB

	Users  *UserDAO
	Papers *PaperDAO
	Chunks *ChunkDAO
}

func NewMySQLDatabase(dsn string) (*Database, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := autoMigrate(db); err != nil {
		return nil, err
	}

	database := &Database{
		DB:     db,
		Users:  NewUserDAO(db),
		Papers: NewPaperDAO(db),
		Chunks: NewChunkDAO(db),
	}

	if err := database.Users.EnsureDefaultUser(context.Background()); err != nil {
		return nil, err
	}

	return database, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.User{},
		&entity.Paper{},
		&entity.Chunk{},
		&entity.Task{},
	)
}
