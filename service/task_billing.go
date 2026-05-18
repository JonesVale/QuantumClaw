package service

// RecordTaskConsumption records the quota consumption for an async task.
// Called after an async task (Midjourney, Suno, etc.) completes successfully.
func RecordTaskConsumption(taskID string, userID int, quota int64, modelName string) error {
	RecordConsumeLog(userID, 0, 0, 0, modelName, "", int(quota), "task:"+taskID)
	return nil
}
