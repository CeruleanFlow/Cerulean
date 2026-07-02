# Cerulean API Draft

Base path: `/api/v1`

## Health

```http
GET /health
```

## Papers

```http
POST /papers
Content-Type: multipart/form-data
field: file=<paper.pdf>
```

Uploads one PDF, calculates SHA256, stores the original artifact, persists metadata, and returns a `Paper`. Duplicate SHA256 uploads return the existing `Paper`.

```http
GET /papers
GET /papers/{paper_id}
GET /papers/{paper_id}/download
GET /papers/{paper_id}/chunks
POST /papers/{paper_id}/ingest
```

`POST /ingest` currently starts an asynchronous placeholder pipeline. Later it will enqueue PaddleOCR and indexing work.

## Tasks

```http
GET /tasks/{task_id}
```

## Search

```http
POST /search
Content-Type: application/json

{
  "query": "what is the method?",
  "top_k": 5,
  "filters": {
    "paper_id": "paper_xxx"
  }
}
```

## Chat

```http
POST /chat
Content-Type: application/json

{
  "question": "summarize the contributions",
  "top_k": 5,
  "filters": {
    "paper_id": "paper_xxx"
  }
}
```

Current chat is a placeholder answer with retrieved sources. DeepSeek will replace it after retrieval is stable.
