package jina

// ChannelName is the display name for the Jina channel.
const ChannelName = "Jina"

// ModelList contains the known Jina models (embeddings, rerank, reader).
var ModelList = []string{
	"jina-embeddings-v3",
	"jina-reranker-v2-base-multilingual",
	"jina-colbert-v2",
	"jina-reader",
}
