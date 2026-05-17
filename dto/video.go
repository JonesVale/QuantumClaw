package dto

type VideoRequest struct {
	Model          string `json:"model"`
	Action         string `json:"action"`
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	ImageUrl       string `json:"image_url,omitempty"`
	Duration       int    `json:"duration,omitempty"`
	Resolution     string `json:"resolution,omitempty"`
}

type VideoResponse struct {
	Id       string `json:"id"`
	Status   string `json:"status"`
	VideoUrl string `json:"video_url,omitempty"`
	Progress int    `json:"progress,omitempty"`
}
