package dto

type MidjourneyRequest struct {
	Action   string `json:"action"`
	Prompt   string `json:"prompt,omitempty"`
	ImageUrl string `json:"image_url,omitempty"`
	MaskUrl  string `json:"mask_url,omitempty"`
	CustomId string `json:"custom_id,omitempty"`
	TaskId   string `json:"task_id,omitempty"`
	Mode     string `json:"mode,omitempty"`
	Style    string `json:"style,omitempty"`
	Ratio    string `json:"ratio,omitempty"`
}

type MidjourneyResponse struct {
	Id         string `json:"id"`
	Action     string `json:"action"`
	Status     string `json:"status"`
	ImageUrl   string `json:"image_url,omitempty"`
	Progress   int    `json:"progress,omitempty"`
	FailReason string `json:"fail_reason,omitempty"`
}
