# Cerulean Server

Go backend for Cerulean, a paper-oriented RAG system for research reading, paper retrieval, source tracing, and citation-aware writing assistance.

This checkpoint focuses on the first real development loop:

1. upload a PDF;
2. calculate SHA256 and de-duplicate;
3. store the original paper artifact;
4. persist paper/chunk metadata;
5. create an ingest task;
6. generate placeholder chunks;
7. search real stored chunks through a local lexical backend.

The placeholder ingest stage is intentionally simple. Its job is to keep the API, database, and UI workflow stable before PaddleOCR, AstraFlow embeddings/rerank, Elasticsearch, Amaranth, and DeepSeek are wired.

## Run

```bash
cp .env.example .env
# optional: source .env manually or use your shell/IDE env loader
go run ./cmd/server
```

The dependency-free defaults are:

```text
CERULEAN_DB_DRIVER=json
CERULEAN_DB_PATH=.var/cerulean.json
CERULEAN_STORAGE_DRIVER=local
CERULEAN_LOCAL_STORAGE_DIR=.var/objects
CERULEAN_SEARCH_DRIVER=local
```

## API

```bash
# health
curl http://localhost:8080/api/v1/health

# upload a PDF
curl -F "file=@paper.pdf" http://localhost:8080/api/v1/papers

# list papers
curl http://localhost:8080/api/v1/papers

# get one paper
curl http://localhost:8080/api/v1/papers/{paper_id}

# download original PDF
curl -L http://localhost:8080/api/v1/papers/{paper_id}/download -o paper.pdf

# start placeholder ingestion
curl -X POST http://localhost:8080/api/v1/papers/{paper_id}/ingest

# inspect generated chunks
curl http://localhost:8080/api/v1/papers/{paper_id}/chunks

# search chunks
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"query":"PaddleOCR embedding Amaranth", "top_k": 5}'

# chat placeholder answer with sources
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"question":"what is the current pipeline?", "top_k": 5}'
```

## Project direction

Cerulean will evolve into:

```text
PDF upload
  -> MinIO artifact storage
  -> PaddleOCR parsing/layout JSON/page images
  -> page-aware chunks
  -> AstraFlow embedding
  -> Amaranth vector indexing
  -> Elasticsearch BM25 indexing
  -> AstraFlow reranking
  -> DeepSeek answer generation
  -> source/citation aware paper reading UI
```
