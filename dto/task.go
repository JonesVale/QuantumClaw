package dto

type TaskSubmitRequest struct {
	Model   string `json:"model"`
	Action  string `json:"action"`
	Prompt  string `json:"prompt,omitempty"`
	ImageUrl string `json:"image_url,omitempty"`
	MaskUrl  string `json:"mask_url,omitempty"`
}

type TaskError struct {
	Err        error
	StatusCode int
	Message    string
}

func NewTaskError(err error, statusCode int, message string) *TaskError {
	return &TaskError{Err: err, StatusCode: statusCode, Message: message}
}
