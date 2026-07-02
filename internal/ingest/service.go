package ingest

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/CeruleanFlow/cerulean/internal/domain"
	"github.com/CeruleanFlow/cerulean/internal/repository"
	"github.com/CeruleanFlow/cerulean/internal/search"
	"github.com/CeruleanFlow/cerulean/internal/storage"
	"github.com/CeruleanFlow/cerulean/internal/task"
)

type Service struct {
	papers repository.PaperRepository
	chunks repository.ChunkRepository
	store  storage.ObjectStorage
	search search.Backend
	tasks  task.Manager
}

func NewService(papers repository.PaperRepository, chunks repository.ChunkRepository, store storage.ObjectStorage, searchBackend search.Backend, tasks task.Manager) *Service {
	return &Service{papers: papers, chunks: chunks, store: store, search: searchBackend, tasks: tasks}
}

func (s *Service) StartPaperIngest(ctx context.Context, paperID string) (task.Task, error) {
	paper, err := s.papers.Get(ctx, paperID)
	if err != nil {
		return task.Task{}, err
	}
	now := time.Now()
	job := task.Task{
		ID:        fmt.Sprintf("task_%d", now.UnixNano()),
		PaperID:   paperID,
		Type:      "paper_ingest",
		Status:    task.Queued,
		Message:   "queued local placeholder ingestion; PaddleOCR worker will replace this stage",
		CreatedAt: now,
		UpdatedAt: now,
	}
	paper.Status = domain.PaperProcessing
	paper.Error = ""
	paper.UpdatedAt = now
	if err := s.papers.Update(ctx, paper); err != nil {
		return task.Task{}, err
	}
	if err := s.tasks.Create(ctx, job); err != nil {
		return task.Task{}, err
	}

	// Keep this asynchronous so the public API shape is already compatible with
	// the future PaddleOCR worker and queue based pipeline.
	go s.runPlaceholderIngest(context.Background(), job, paper)
	return job, nil
}

func (s *Service) runPlaceholderIngest(ctx context.Context, job task.Task, paper domain.Paper) {
	now := time.Now()
	job.Status = task.Running
	job.Message = "building placeholder chunks"
	job.UpdatedAt = now
	_ = s.tasks.Update(ctx, job)

	markdown := placeholderMarkdown(paper)
	artifactKey := fmt.Sprintf("papers/%s/parsed/document.md", paper.ID)
	if _, err := s.store.Put(ctx, artifactKey, bytes.NewReader([]byte(markdown)), int64(len(markdown)), storage.PutOptions{ContentType: "text/markdown; charset=utf-8"}); err != nil {
		s.fail(ctx, job, paper, err)
		return
	}

	chunks := makePlaceholderChunks(paper, artifactKey)
	if err := s.chunks.DeleteByPaperID(ctx, paper.ID); err != nil {
		s.fail(ctx, job, paper, err)
		return
	}
	if err := s.chunks.UpsertMany(ctx, chunks); err != nil {
		s.fail(ctx, job, paper, err)
		return
	}
	if s.search != nil {
		if err := s.search.Index(ctx, chunks); err != nil {
			s.fail(ctx, job, paper, err)
			return
		}
	}

	now = time.Now()
	paper.Status = domain.PaperParsed
	paper.PageCount = 1
	paper.Error = ""
	paper.UpdatedAt = now
	_ = s.papers.Update(ctx, paper)
	job.Status = task.Succeeded
	job.Message = "placeholder ingestion completed; next step is wiring PaddleOCR output into the same chunk path"
	job.UpdatedAt = now
	_ = s.tasks.Update(ctx, job)
}

func (s *Service) fail(ctx context.Context, job task.Task, paper domain.Paper, err error) {
	now := time.Now()
	paper.Status = domain.PaperFailed
	paper.Error = err.Error()
	paper.UpdatedAt = now
	_ = s.papers.Update(ctx, paper)
	job.Status = task.Failed
	job.Message = err.Error()
	job.UpdatedAt = now
	_ = s.tasks.Update(ctx, job)
}

func placeholderMarkdown(paper domain.Paper) string {
	return fmt.Sprintf("# %s\n\n"+
		"This is a placeholder parsed document for **%s**.\n\n"+
		"The current Cerulean pipeline has stored the original PDF object at `%s` and created searchable placeholder chunks. The next implementation step is to replace this placeholder with PaddleOCR output, page-level layout JSON, chunking, embedding, reranking, and hybrid retrieval.\n\n"+
		"Suggested future metadata:\n\n"+
		"- sha256: `%s`\n"+
		"- content_type: `%s`\n"+
		"- file_size: `%d`\n", paper.Title, paper.Filename, paper.ObjectKey, paper.SHA256, paper.ContentType, paper.Size)
}

func makePlaceholderChunks(paper domain.Paper, artifactKey string) []domain.Chunk {
	now := time.Now()
	base := []string{
		fmt.Sprintf("Paper title: %s. Original filename: %s. This paper has been uploaded into Cerulean and is ready for OCR parsing.", paper.Title, paper.Filename),
		"Cerulean is a paper-oriented RAG system. The planned pipeline is PDF upload, MinIO artifact storage, PaddleOCR document parsing, chunking, AstraFlow embedding, Amaranth vector retrieval, Elasticsearch lexical retrieval, AstraFlow reranking, and DeepSeek answer generation.",
		"This placeholder chunk exists so search and RAG screens can be tested before PaddleOCR is connected. After OCR is wired, this chunk will be replaced by page-aware content with page numbers, section labels, and citation information.",
	}
	chunks := make([]domain.Chunk, 0, len(base))
	for i, text := range base {
		chunks = append(chunks, domain.Chunk{
			ID:        fmt.Sprintf("%s_chunk_%03d", paper.ID, i+1),
			PaperID:   paper.ID,
			PageNo:    1,
			Index:     i,
			Text:      strings.TrimSpace(text),
			ObjectKey: artifactKey,
			Metadata: map[string]string{
				"source": "placeholder_ingest",
				"title":  paper.Title,
			},
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	return chunks
}
