package repository

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/CeruleanFlow/cerulean/internal/domain"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("already exists")
)

type PaperRepository interface {
	Create(ctx context.Context, paper domain.Paper) error
	Update(ctx context.Context, paper domain.Paper) error
	Get(ctx context.Context, id string) (domain.Paper, error)
	FindBySHA256(ctx context.Context, sha256 string) (domain.Paper, error)
	List(ctx context.Context) ([]domain.Paper, error)
	Delete(ctx context.Context, id string) error
}

type MemoryPaperRepository struct {
	mu     sync.RWMutex
	papers map[string]domain.Paper
}

func NewMemoryPaperRepository() *MemoryPaperRepository {
	return &MemoryPaperRepository{papers: make(map[string]domain.Paper)}
}

func (r *MemoryPaperRepository) Create(ctx context.Context, paper domain.Paper) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.papers[paper.ID]; ok {
		return ErrConflict
	}
	r.papers[paper.ID] = paper
	return nil
}

func (r *MemoryPaperRepository) Update(ctx context.Context, paper domain.Paper) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.papers[paper.ID]; !ok {
		return ErrNotFound
	}
	r.papers[paper.ID] = paper
	return nil
}

func (r *MemoryPaperRepository) Get(ctx context.Context, id string) (domain.Paper, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	paper, ok := r.papers[id]
	if !ok {
		return domain.Paper{}, ErrNotFound
	}
	return paper, nil
}

func (r *MemoryPaperRepository) FindBySHA256(ctx context.Context, sha256 string) (domain.Paper, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, paper := range r.papers {
		if paper.SHA256 == sha256 {
			return paper, nil
		}
	}
	return domain.Paper{}, ErrNotFound
}

func (r *MemoryPaperRepository) List(ctx context.Context) ([]domain.Paper, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]domain.Paper, 0, len(r.papers))
	for _, paper := range r.papers {
		items = append(items, paper)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items, nil
}

func (r *MemoryPaperRepository) Delete(ctx context.Context, id string) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.papers[id]; !ok {
		return ErrNotFound
	}
	delete(r.papers, id)
	return nil
}
