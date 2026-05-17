package service

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"

	"github.com/quantumclaw/quantumclaw/dto"
)

// StreamScanner scans an SSE stream for "data:" lines and extracts usage information.
type StreamScanner struct {
	scanner *bufio.Scanner
	data    []byte
	done    bool
}

// NewStreamScanner creates a new StreamScanner from the given reader.
func NewStreamScanner(reader io.Reader) *StreamScanner {
	return &StreamScanner{
		scanner: bufio.NewScanner(reader),
	}
}

// Scan advances the scanner to the next SSE data event.
// Returns false when the stream ends or reaches the [DONE] marker.
func (s *StreamScanner) Scan() bool {
	if s.done {
		return false
	}
	for s.scanner.Scan() {
		line := s.scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				s.done = true
				return false
			}
			s.data = []byte(data)
			return true
		}
	}
	return false
}

// Data returns the raw bytes of the current SSE data event.
func (s *StreamScanner) Data() []byte {
	return s.data
}

// ExtractUsage attempts to extract usage information from the current data event.
func (s *StreamScanner) ExtractUsage() *dto.Usage {
	var streamResp struct {
		Usage *dto.Usage `json:"usage"`
	}
	if err := json.Unmarshal(s.data, &streamResp); err != nil {
		return nil
	}
	return streamResp.Usage
}
