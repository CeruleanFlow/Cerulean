package main

import (
	"fmt"
	"github.com/CeruleanFlow/cerulean/internal/dao"
	"log"
	"net/http"
	"strings"

	"github.com/CeruleanFlow/cerulean/internal/api"
	"github.com/CeruleanFlow/cerulean/internal/config"
	"github.com/CeruleanFlow/cerulean/internal/ingest"
	"github.com/CeruleanFlow/cerulean/internal/rag"
	"github.com/CeruleanFlow/cerulean/internal/repository"
	"github.com/CeruleanFlow/cerulean/internal/search"
	"github.com/CeruleanFlow/cerulean/internal/search/amaranth"
	"github.com/CeruleanFlow/cerulean/internal/search/elastic"
	"github.com/CeruleanFlow/cerulean/internal/search/local"
	"github.com/CeruleanFlow/cerulean/internal/storage"
	"github.com/CeruleanFlow/cerulean/internal/task"
)

func main() {
	cfg := config.Load()

	paperRepo, chunkRepo, userDAO, err := buildRepositories(cfg)
	if err != nil {
		log.Fatal(err)
	}
	objectStore, err := buildObjectStorage(cfg)
	if err != nil {
		log.Fatal(err)
	}
	taskManager := task.NewMemoryManager()
	searchBackend := buildSearchBackend(cfg, chunkRepo)

	ingestService := ingest.NewService(paperRepo, chunkRepo, objectStore, searchBackend, taskManager)
	ragService := rag.NewService(paperRepo, searchBackend)

	router := api.NewRouter(api.RouterOptions{
		Config:        cfg,
		PaperRepo:     paperRepo,
		ChunkRepo:     chunkRepo,
		UserDAO:       userDAO,
		TaskManager:   taskManager,
		ObjectStore:   objectStore,
		IngestService: ingestService,
		RAGService:    ragService,
	})

	log.Printf("Cerulean server listening on %s", cfg.HTTPAddr)
	log.Printf("database=%s storage=%s search=%s", cfg.DBDriver, cfg.StorageDriver, cfg.SearchDriver)
	if err := http.ListenAndServe(cfg.HTTPAddr, router); err != nil {
		log.Fatal(err)
	}
}

func buildRepositories(cfg config.Config) (repository.PaperRepository, repository.ChunkRepository, *dao.UserDAO, error) {
	switch strings.ToLower(cfg.DBDriver) {
	case "", "mysql":
		database, err := dao.NewMySQLDatabase(cfg.MySQLDSN)
		if err != nil {
			return nil, nil, nil, err
		}
		return database.Papers, database.Chunks, database.Users, nil

	case "json":
		repo, err := repository.NewJSONRepository(cfg.DBPath)
		if err != nil {
			return nil, nil, nil, err
		}
		return repo, repository.NewJSONChunkRepository(repo), nil, nil

	case "memory":
		return repository.NewMemoryPaperRepository(), repository.NewMemoryChunkRepository(), nil, nil

	default:
		return nil, nil, nil, fmt.Errorf("unsupported CERULEAN_DB_DRIVER=%q; supported: mysql, json, memory", cfg.DBDriver)
	}
}

func buildObjectStorage(cfg config.Config) (storage.ObjectStorage, error) {
	switch strings.ToLower(cfg.StorageDriver) {
	case "", "local":
		return storage.NewLocalObjectStorage(cfg.LocalStorageDir), nil
	case "minio":
		return nil, fmt.Errorf("minio storage is the next adapter to wire; use CERULEAN_STORAGE_DRIVER=local for this dependency-free checkpoint")
	default:
		return nil, fmt.Errorf("unsupported CERULEAN_STORAGE_DRIVER=%q; supported: local", cfg.StorageDriver)
	}
}

func buildSearchBackend(cfg config.Config, chunks repository.ChunkRepository) search.Backend {
	localBackend := local.NewBackend(chunks)
	switch strings.ToLower(cfg.SearchDriver) {
	case "", "local":
		return localBackend
	case "hybrid":
		elasticBackend := elastic.NewBackend(cfg.ElasticURL, cfg.ElasticIndex)
		amaranthBackend := amaranth.NewBackend(cfg.AmaranthURL, cfg.AmaranthCollection)
		return search.NewHybridBackend(elasticBackend, amaranthBackend, search.NewRRFusion(60))
	default:
		log.Printf("unknown CERULEAN_SEARCH_DRIVER=%q; fallback to local", cfg.SearchDriver)
		return localBackend
	}
}
