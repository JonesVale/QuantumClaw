package types

// ChannelError holds metadata about a channel error for retry/ban decisions.
type ChannelError struct {
	ChannelID   int    `json:"channel_id"`
	ChannelType int    `json:"channel_type"`
	ChannelName string `json:"channel_name"`
	IsMultiKey  bool   `json:"is_multi_key"`
	AutoBan     bool   `json:"auto_ban"`
	UsingKey    string `json:"using_key"`
}

func NewChannelError(channelID int, channelType int, channelName string, isMultiKey bool, usingKey string, autoBan bool) *ChannelError {
	return &ChannelError{
		ChannelID:   channelID,
		ChannelType: channelType,
		ChannelName: channelName,
		IsMultiKey:  isMultiKey,
		AutoBan:     autoBan,
		UsingKey:    usingKey,
	}
}
