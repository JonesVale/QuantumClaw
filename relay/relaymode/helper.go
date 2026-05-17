package relaymode

import "strings"

func GetByPath(path string) int {
	relayMode := Unknown
	if strings.HasPrefix(path, "/v1/responses") {
		relayMode = Responses
	} else if strings.HasPrefix(path, "/v1/chat/completions") {
		relayMode = ChatCompletions
	} else if strings.HasPrefix(path, "/v1/completions") {
		relayMode = Completions
	} else if strings.HasPrefix(path, "/v1/embeddings") {
		relayMode = Embeddings
	} else if strings.HasSuffix(path, "embeddings") {
		relayMode = Embeddings
	} else if strings.HasPrefix(path, "/v1/moderations") {
		relayMode = Moderations
	} else if strings.HasPrefix(path, "/v1/images/generations") {
		relayMode = ImagesGenerations
	} else if strings.HasPrefix(path, "/v1/edits") {
		relayMode = Edits
	} else if strings.HasPrefix(path, "/v1/audio/speech") {
		relayMode = AudioSpeech
	} else if strings.HasPrefix(path, "/v1/audio/transcriptions") {
		relayMode = AudioTranscription
	} else if strings.HasPrefix(path, "/v1/audio/translations") {
		relayMode = AudioTranslation
	} else if strings.HasPrefix(path, "/v1/oneapi/proxy") {
		relayMode = Proxy
	} else if strings.HasPrefix(path, "/v1/assistants") {
		relayMode = Assistants
	} else if strings.HasPrefix(path, "/v1/threads") {
		relayMode = AssistantsThreads
	} else if strings.HasPrefix(path, "/v1/files") {
		relayMode = Files
	} else if strings.HasPrefix(path, "/v1/fine_tuning") {
		relayMode = FineTuning
	} else if strings.HasPrefix(path, "/mj/") {
		// Midjourney 异步任务：/mj/submit/*  /mj/task/*
		relayMode = Midjourney
	} else if strings.HasPrefix(path, "/video/") {
		// 视频生成异步任务：/video/submit/*  /video/task/*
		relayMode = VideoGeneration
	} else if strings.HasPrefix(path, "/suno/") {
		// Suno 音乐生成异步任务：/suno/submit/*  /suno/task/*
		relayMode = Suno
	} else if strings.HasPrefix(path, "/v1/messages") {
		// Claude Messages native API
		relayMode = ClaudeMessages
	} else if strings.HasPrefix(path, "/v1/batches") {
		// OpenAI Batch API
		relayMode = Batches
	} else if strings.HasPrefix(path, "/v1/vector_stores") {
		// OpenAI Vector Stores API
		relayMode = VectorStores
	}
	return relayMode
}
