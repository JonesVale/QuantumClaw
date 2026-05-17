package dto

type SensitiveCheckRequest struct {
	Content string   `json:"content"`
	Types   []string `json:"types,omitempty"`
}

type SensitiveCheckResponse struct {
	Passed  bool     `json:"passed"`
	Matched []string `json:"matched,omitempty"`
	Message string   `json:"message,omitempty"`
}
