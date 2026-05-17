package dto

// Realtime event type constants.
const (
	RealtimeEventTypeError                   = "error"
	RealtimeEventTypeSessionUpdate           = "session.update"
	RealtimeEventTypeConversationCreate      = "conversation.item.create"
	RealtimeEventTypeResponseCreate          = "response.create"
	RealtimeEventInputAudioBufferAppend      = "input_audio_buffer.append"

	RealtimeEventTypeResponseDone                   = "response.done"
	RealtimeEventTypeSessionUpdated                 = "session.updated"
	RealtimeEventTypeSessionCreated                 = "session.created"
	RealtimeEventResponseAudioDelta                 = "response.audio.delta"
	RealtimeEventResponseAudioTranscriptionDelta    = "response.audio_transcript.delta"
	RealtimeEventResponseFunctionCallArgumentsDelta = "response.function_call_arguments.delta"
	RealtimeEventResponseFunctionCallArgumentsDone  = "response.function_call_arguments.done"
	RealtimeEventConversationItemCreated            = "conversation.item.created"
)

// RealtimeEvent represents a WebSocket realtime event.
type RealtimeEvent struct {
	EventID  string           `json:"event_id"`
	Type     string           `json:"type"`
	Session  *RealtimeSession `json:"session,omitempty"`
	Item     *RealtimeItem    `json:"item,omitempty"`
	Error    *ErrorInfo       `json:"error,omitempty"`
	Response *RealtimeResponse `json:"response,omitempty"`
	Delta    string           `json:"delta,omitempty"`
	Audio    string           `json:"audio,omitempty"`
}

// ErrorInfo holds error information in realtime events.
type ErrorInfo struct {
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// RealtimeResponse holds usage from a realtime response.
type RealtimeResponse struct {
	Usage *RealtimeUsage `json:"usage"`
}

// RealtimeUsage holds token usage for realtime sessions.
type RealtimeUsage struct {
	TotalTokens        int                `json:"total_tokens"`
	InputTokens        int                `json:"input_tokens"`
	OutputTokens       int                `json:"output_tokens"`
	InputTokenDetails  InputTokenDetails  `json:"input_token_details"`
	OutputTokenDetails OutputTokenDetails `json:"output_token_details"`
}

// InputTokenDetails breaks down input token usage.
type InputTokenDetails struct {
	CachedTokens int `json:"cached_tokens"`
	TextTokens   int `json:"text_tokens"`
	AudioTokens  int `json:"audio_tokens"`
}

// OutputTokenDetails breaks down output token usage.
type OutputTokenDetails struct {
	TextTokens  int `json:"text_tokens"`
	AudioTokens int `json:"audio_tokens"`
}

// RealtimeSession represents a realtime session configuration.
type RealtimeSession struct {
	Modalities              []string          `json:"modalities"`
	Instructions            string            `json:"instructions"`
	Voice                   string            `json:"voice"`
	InputAudioFormat        string            `json:"input_audio_format"`
	OutputAudioFormat       string            `json:"output_audio_format"`
	InputAudioTranscription InputAudioTranscription `json:"input_audio_transcription"`
	TurnDetection           interface{}       `json:"turn_detection"`
	Tools                   []RealTimeTool    `json:"tools"`
	ToolChoice              string            `json:"tool_choice"`
	Temperature             float64           `json:"temperature"`
}

// InputAudioTranscription configures audio transcription in realtime.
type InputAudioTranscription struct {
	Model string `json:"model"`
}

// RealTimeTool represents a tool definition for realtime sessions.
type RealTimeTool struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// RealtimeItem represents an item in a realtime conversation.
type RealtimeItem struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Status    string            `json:"status"`
	Role      string            `json:"role"`
	Content   []RealtimeContent `json:"content"`
	Name      *string           `json:"name,omitempty"`
	ToolCalls any               `json:"tool_calls,omitempty"`
	CallID    string            `json:"call_id,omitempty"`
}

// RealtimeContent represents a content block in a realtime item.
type RealtimeContent struct {
	Type       string `json:"type"`
	Text       string `json:"text,omitempty"`
	Audio      string `json:"audio,omitempty"`
	Transcript string `json:"transcript,omitempty"`
}
