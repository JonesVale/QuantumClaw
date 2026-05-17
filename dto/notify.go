package dto

type NotifyRequest struct {
	Type       string `json:"type"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	TargetUser string `json:"target_user,omitempty"`
	Channel    string `json:"channel,omitempty"`
}
