package local

import (
	"context"
	"sort"
	"strings"
	"unicode"

	"github.com/CeruleanFlow/cerulean/internal/domain"
	"github.com/CeruleanFlow/cerulean/internal/repository"
	"github.com/CeruleanFlow/cerulean/internal/search"
)

// Backend is a dependency-free lexical backend used before Elasticsearch and
// Amaranth are wired. It gives the UI a real searchable path after ingestion.
type Backend struct {
	chunks repository.ChunkRepository
}

func NewBackend(chunks repository.ChunkRepository) *Backend {
	return &Backend{chunks: chunks}
}

func (b *Backend) Name() string { return "local_keyword" }

func (b *Backend) Index(ctx context.Context, chunks []domain.Chunk) error {
	return b.chunks.UpsertMany(ctx, chunks)
}

func (b *Backend) Search(ctx context.Context, query search.Query) ([]search.Result, error) {
	if strings.TrimSpace(query.Text) == "" {
		return nil, nil
	}
	chunks, err := b.chunks.List(ctx, query.Filters)
	if err != nil {
		return nil, err
	}
	terms := tokenize(query.Text)
	results := make([]search.Result, 0, len(chunks))
	for _, chunk := range chunks {
		score := scoreText(chunk.Text, terms)
		if score <= 0 {
			continue
		}
		results = append(results, search.Result{
			ChunkID: chunk.ID,
			PaperID: chunk.PaperID,
			PageNo:  chunk.PageNo,
			Text:    chunk.Text,
			Score:   score,
			Backend: b.Name(),
		})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	if query.TopK > 0 && len(results) > query.TopK {
		results = results[:query.TopK]
	}
	return results, nil
}

func (b *Backend) DeleteByPaperID(ctx context.Context, paperID string) error {
	return b.chunks.DeleteByPaperID(ctx, paperID)
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-'
	})
	terms := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || seen[part] {
			continue
		}
		seen[part] = true
		terms = append(terms, part)
	}
	return terms
}

func scoreText(text string, terms []string) float64 {
	if len(terms) == 0 {
		return 0
	}
	lower := strings.ToLower(text)
	var score float64
	for _, term := range terms {
		count := strings.Count(lower, term)
		if count > 0 {
			score += 1 + float64(count)*0.25
		}
	}
	if score > 0 {
		score /= float64(len(terms))
	}
	return score
}
