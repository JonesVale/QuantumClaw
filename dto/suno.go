package dto

type SunoRequest struct {
	Model            string `json:"model"`
	Action           string `json:"action"`
	Prompt           string `json:"prompt"`
	Tags             string `json:"tags,omitempty"`
	Style            string `json:"style,omitempty"`
	Title            string `json:"title,omitempty"`
	ContinueAt       int    `json:"continue_at,omitempty"`
	ContinueClipId   string `json:"continue_clip_id,omitempty"`
	MakeInstrumental bool   `json:"make_instrumental,omitempty"`
}

type SunoResponse struct {
	Id       string `json:"id"`
	Status   string `json:"status"`
	AudioUrl string `json:"audio_url,omitempty"`
	VideoUrl string `json:"video_url,omitempty"`
	ImageUrl string `json:"image_url,omitempty"`
	Title    string `json:"title,omitempty"`
	Lyric    string `json:"lyric,omitempty"`
}
