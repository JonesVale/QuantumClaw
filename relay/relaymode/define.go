package relaymode

const (
	Unknown = iota
	ChatCompletions
	Completions
	Embeddings
	Moderations
	ImagesGenerations
	Edits
	AudioSpeech
	AudioTranscription
	AudioTranslation
	// Proxy is a special relay mode for proxying requests to custom upstream
	Proxy
	// Assistants API
	Assistants
	AssistantsFiles
	AssistantsThreads
	// Files API
	Files
	// Fine-tuning API
	FineTuning
	Responses
	// Async task modes (Midjourney, Video, Suno, etc.)
	Midjourney
	VideoGeneration
	Suno
	// Claude Messages API
	ClaudeMessages
	// OpenAI Batch API
	Batches
	// OpenAI Vector Stores API
	VectorStores
)
