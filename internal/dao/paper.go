package dao

import (
	"context"
	"errors"
	"time"

	"github.com/CeruleanFlow/cerulean/internal/domain"
	"github.com/CeruleanFlow/cerulean/internal/entity"
	"github.com/CeruleanFlow/cerulean/internal/repository"
	"gorm.io/gorm"
)

type PaperDAO struct {
	db *gorm.DB
}

func NewPaperDAO(db *gorm.DB) *PaperDAO {
	return &PaperDAO{db: db}
}

func (d *PaperDAO) Create(ctx context.Context, paper domain.Paper) error {
	model := toPaperEntity(paper)
	if model.UserID == "" {
		model.UserID = DefaultUserID
	}

	return d.db.WithContext(ctx).Create(&model).Error
}

func (d *PaperDAO) Update(ctx context.Context, paper domain.Paper) error {
	var count int64

	if err := d.db.WithContext(ctx).
		Model(&entity.Paper{}).
		Where("id = ?", paper.ID).
		Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		return repository.ErrNotFound
	}

	model := toPaperEntity(paper)

	return d.db.WithContext(ctx).Save(&model).Error
}

func (d *PaperDAO) Get(ctx context.Context, id string) (domain.Paper, error) {
	var model entity.Paper

	err := d.db.WithContext(ctx).
		First(&model, "id = ?", id).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Paper{}, repository.ErrNotFound
		}
		return domain.Paper{}, err
	}

	return paperEntityToDomain(model), nil
}

func (d *PaperDAO) FindBySHA256(ctx context.Context, sha256 string) (domain.Paper, error) {
	var model entity.Paper

	err := d.db.WithContext(ctx).
		First(&model, "sha256 = ?", sha256).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Paper{}, repository.ErrNotFound
		}
		return domain.Paper{}, err
	}

	return paperEntityToDomain(model), nil
}

func (d *PaperDAO) List(ctx context.Context) ([]domain.Paper, error) {
	var models []entity.Paper

	if err := d.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	papers := make([]domain.Paper, 0, len(models))
	for _, model := range models {
		papers = append(papers, paperEntityToDomain(model))
	}

	return papers, nil
}

func (d *PaperDAO) Delete(ctx context.Context, id string) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&entity.Paper{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return repository.ErrNotFound
		}

		if err := tx.Delete(&entity.Chunk{}, "paper_id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}

func toPaperEntity(paper domain.Paper) entity.Paper {
	return entity.Paper{
		ID:          paper.ID,
		UserID:      DefaultUserID,
		Title:       paper.Title,
		Filename:    paper.Filename,
		ContentType: paper.ContentType,
		Size:        paper.Size,
		SHA256:      paper.SHA256,
		ObjectKey:   paper.ObjectKey,
		Status:      string(paper.Status),
		PageCount:   paper.PageCount,
		Error:       paper.Error,
		CreatedAt:   normalizeTime(paper.CreatedAt),
		UpdatedAt:   normalizeTime(paper.UpdatedAt),
	}
}

func paperEntityToDomain(model entity.Paper) domain.Paper {
	return domain.Paper{
		ID:          model.ID,
		Title:       model.Title,
		Filename:    model.Filename,
		ContentType: model.ContentType,
		Size:        model.Size,
		SHA256:      model.SHA256,
		ObjectKey:   model.ObjectKey,
		Status:      domain.PaperStatus(model.Status),
		PageCount:   model.PageCount,
		Error:       model.Error,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func normalizeTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}
