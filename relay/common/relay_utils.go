package common

import (
	"strings"
)

func GetFullRequestURL(baseURL string, requestURL string, channelType int) string {
	fullURL := strings.TrimRight(baseURL, "/") + requestURL
	return fullURL
}

func GetRequestURLPath(relayMode int, modelName string) string {
	switch relayMode {
	default:
		return "/v1/chat/completions"
	}
}
