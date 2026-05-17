package types

// FileType represents the type of file in a request.
type FileType string

const (
	FileTypeImage FileType = "image"
	FileTypeAudio FileType = "audio"
	FileTypeVideo FileType = "video"
	FileTypeFile  FileType = "file"
)

// TokenType categorises how tokens are counted.
type TokenType string

const (
	TokenTypeTextNumber TokenType = "text_number"
	TokenTypeTokenizer  TokenType = "tokenizer"
	TokenTypeImage      TokenType = "image"
)

// TokenCountMeta holds metadata for token counting.
type TokenCountMeta struct {
	TokenType     TokenType   `json:"token_type,omitempty"`
	CombineText   string      `json:"combine_text,omitempty"`
	ToolsCount    int         `json:"tools_count,omitempty"`
	NameCount     int         `json:"name_count,omitempty"`
	MessagesCount int         `json:"messages_count,omitempty"`
	Files         []*FileMeta `json:"files,omitempty"`
	MaxTokens     int         `json:"max_tokens,omitempty"`
	ImagePriceRatio float64   `json:"image_ratio,omitempty"`
}

// FileMeta describes a file included in a request.
type FileMeta struct {
	FileType FileType   `json:"file_type"`
	Source   FileSource `json:"source"`
	Detail   string     `json:"detail,omitempty"`
}

func NewFileMeta(fileType FileType, source FileSource) *FileMeta {
	return &FileMeta{
		FileType: fileType,
		Source:   source,
	}
}

func NewImageFileMeta(source FileSource, detail string) *FileMeta {
	return &FileMeta{
		FileType: FileTypeImage,
		Source:   source,
		Detail:   detail,
	}
}

func (f *FileMeta) GetIdentifier() string {
	if f.Source != nil {
		return f.Source.GetIdentifier()
	}
	return "unknown"
}

func (f *FileMeta) IsURL() bool {
	return f.Source != nil && f.Source.IsURL()
}

// RequestMeta holds metadata about the original request.
type RequestMeta struct {
	OriginalModelName string `json:"original_model_name"`
	UserUsingGroup    string `json:"user_using_group"`
	PromptTokens      int    `json:"prompt_tokens"`
	PreConsumedQuota  int    `json:"pre_consumed_quota"`
}
