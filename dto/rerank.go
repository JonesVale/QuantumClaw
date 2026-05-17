package dto

type RerankRequest struct {
	Model          string  `json:"model"`
	Query          string  `json:"query"`
	Documents      []any   `json:"documents"`
	TopN          int     `json:"top_n,omitempty"`
	ReturnDocuments bool   `json:"return_documents,omitempty"`
	MaxChunksPerDoc int    `json:"max_chunks_per_doc,omitempty"`
}

type RerankResponse struct {
	Id      string        `json:"id"`
	Results []RerankResult `json:"results"`
	Meta    any           `json:"meta,omitempty"`
}

type RerankResult struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
	Document       any     `json:"document,omitempty"`
}
