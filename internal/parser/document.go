package parser

import "context"

type Page struct {
	PageNo int    `json:"page_no"`
	Text   string `json:"text"`
}

type Document struct {
	PageCount int    `json:"page_count"`
	Pages     []Page `json:"pages"`
	Text      string `json:"text"`
}

type Parser interface {
	ParseFile(ctx context.Context, path string) (Document, error)
}
