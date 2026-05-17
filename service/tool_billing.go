package service

// ToolBilling handles billing for tool calls (function calling, web search, etc.).
// This is a stub for future implementation.
type ToolBilling struct {
	ToolName string
}

// RecordToolUsage records the quota consumption for a tool call.
// TODO: implement tool billing logic
func RecordToolUsage(toolName string, userID int, quota int64) error {
	return nil
}
