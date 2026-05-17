package dto

type OpenAIResponsesCompactionRequest struct {
	Model              string              `json:"model"`
	Input              OpenAIResponsesInput `json:"input"`
	PreviousResponseId string              `json:"previous_response_id,omitempty"`
	Store              bool                `json:"store,omitempty"`
	Metadata           map[string]string   `json:"metadata,omitempty"`
}

type OpenAIResponsesInput struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Role    string `json:"role,omitempty"`
}
