package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/CeruleanFlow/cerulean/internal/domain"
)

// JSONRepository is a dependency-free development database. It keeps the MVP
// easy to run before we introduce PostgreSQL/SQLite migrations.
//
// It implements PaperRepository directly. Use NewJSONChunkRepository(repo) to
// get the ChunkRepository view over the same JSON file.
type JSONRepository struct {
	mu   sync.RWMutex
	path string
	data jsonRepositoryData
}

type jsonRepositoryData struct {
	Papers map[string]domain.Paper `json:"papers"`
	Chunks map[string]domain.Chunk `json:"chunks"`
}

func NewJSONRepository(path string) (*JSONRepository, error) {
	repo := &JSONRepository{path: path}
	if err := repo.load(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *JSONRepository) Create(ctx context.Context, paper domain.Paper) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data.Papers[paper.ID]; ok {
		return ErrConflict
	}
	r.data.Papers[paper.ID] = paper
	return r.saveLocked()
}

func (r *JSONRepository) Update(ctx context.Context, paper domain.Paper) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data.Papers[paper.ID]; !ok {
		return ErrNotFound
	}
	r.data.Papers[paper.ID] = paper
	return r.saveLocked()
}

func (r *JSONRepository) Get(ctx context.Context, id string) (domain.Paper, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	paper, ok := r.data.Papers[id]
	if !ok {
		return domain.Paper{}, ErrNotFound
	}
	return paper, nil
}

func (r *JSONRepository) FindBySHA256(ctx context.Context, sha256 string) (domain.Paper, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, paper := range r.data.Papers {
		if paper.SHA256 == sha256 {
			return paper, nil
		}
	}
	return domain.Paper{}, ErrNotFound
}

func (r *JSONRepository) List(ctx context.Context) ([]domain.Paper, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]domain.Paper, 0, len(r.data.Papers))
	for _, paper := range r.data.Papers {
		items = append(items, paper)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items, nil
}

func (r *JSONRepository) Delete(ctx context.Context, id string) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data.Papers[id]; !ok {
		return ErrNotFound
	}
	delete(r.data.Papers, id)
	for chunkID, chunk := range r.data.Chunks {
		if chunk.PaperID == id {
			delete(r.data.Chunks, chunkID)
		}
	}
	return r.saveLocked()
}

func (r *JSONRepository) upsertChunks(ctx context.Context, chunks []domain.Chunk) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, chunk := range chunks {
		r.data.Chunks[chunk.ID] = chunk
	}
	return r.saveLocked()
}

func (r *JSONRepository) listChunks(ctx context.Context, filters map[string]string) ([]domain.Chunk, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]domain.Chunk, 0, len(r.data.Chunks))
	for _, chunk := range r.data.Chunks {
		if matchChunkFilters(chunk, filters) {
			items = append(items, chunk)
		}
	}
	sortChunks(items)
	return items, nil
}

func (r *JSONRepository) deleteChunksByPaperID(ctx context.Context, paperID string) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	for chunkID, chunk := range r.data.Chunks {
		if chunk.PaperID == paperID {
			delete(r.data.Chunks, chunkID)
		}
	}
	return r.saveLocked()
}

func (r *JSONRepository) load() error {
	r.data = jsonRepositoryData{
		Papers: make(map[string]domain.Paper),
		Chunks: make(map[string]domain.Chunk),
	}
	file, err := os.Open(r.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(&r.data); err != nil {
		return err
	}
	if r.data.Papers == nil {
		r.data.Papers = make(map[string]domain.Paper)
	}
	if r.data.Chunks == nil {
		r.data.Chunks = make(map[string]domain.Chunk)
	}
	return nil
}

func (r *JSONRepository) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return err
	}
	tmp := r.path + ".tmp"
	file, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r.data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, r.path)
}

type JSONChunkRepository struct {
	repo *JSONRepository
}

func NewJSONChunkRepository(repo *JSONRepository) *JSONChunkRepository {
	return &JSONChunkRepository{repo: repo}
}

func (r *JSONChunkRepository) UpsertMany(ctx context.Context, chunks []domain.Chunk) error {
	return r.repo.upsertChunks(ctx, chunks)
}

func (r *JSONChunkRepository) List(ctx context.Context, filters map[string]string) ([]domain.Chunk, error) {
	return r.repo.listChunks(ctx, filters)
}

func (r *JSONChunkRepository) ListByPaperID(ctx context.Context, paperID string) ([]domain.Chunk, error) {
	return r.repo.listChunks(ctx, map[string]string{"paper_id": paperID})
}

func (r *JSONChunkRepository) DeleteByPaperID(ctx context.Context, paperID string) error {
	return r.repo.deleteChunksByPaperID(ctx, paperID)
}
