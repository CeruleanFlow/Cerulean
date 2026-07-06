package rag

import (
	"context"

	"github.com/CeruleanFlow/cerulean/internal/domain"
	"github.com/CeruleanFlow/cerulean/internal/repository"
	"github.com/CeruleanFlow/cerulean/internal/search"
)

type Service struct {
	papers repository.PaperRepository
	search search.Backend
}

func NewService(papers repository.PaperRepository, searchBackend search.Backend) *Service {
	return &Service{papers: papers, search: searchBackend}
}

func (s *Service) Search(ctx context.Context, req domain.SearchRequest) (domain.SearchResponse, error) {
	if s.search == nil {
		return domain.SearchResponse{
			Query:   req.Query,
			Results: []domain.SearchResult{},
		}, nil
	}

	if req.TopK <= 0 {
		req.TopK = 10
	}

	return s.search.Search(ctx, req)
}

//func (s *Service) Chat(ctx context.Context, req domain.ChatRequest) (domain.ChatResponse, error) {
//	if req.TopK <= 0 {
//		req.TopK = 5
//	}
//	results, err := s.search.Search(ctx, req)
//	if err != nil {
//		return domain.ChatResponse{}, err
//	}
//	sources := toSources(results)
//	answer := fmt.Sprintf("MVP mock answer for: %q. Wire LLM completion after retrieval is ready.", req.Question)
//	return domain.ChatResponse{Answer: answer, Sources: sources}, nil
//}

func toSources(results []search.Result) []domain.Source {
	sources := make([]domain.Source, 0, len(results))
	for _, result := range results {
		sources = append(sources, domain.Source{
			ChunkID: result.ChunkID,
			PaperID: result.PaperID,
			PageNo:  result.PageNo,
			Text:    result.Text,
			Score:   result.Score,
			Backend: result.Backend,
		})
	}
	return sources
}
