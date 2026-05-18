package service

// RecordToolUsage records the quota consumption for a tool call
// (function calling, web search, code interpreter, etc.).
func RecordToolUsage(toolName string, userID int, quota int64) error {
	RecordConsumeLog(userID, 0, 0, 0, "tool:"+toolName, "", int(quota), "tool:"+toolName)
	return nil
}
