package service

// TaskBilling handles billing for async tasks (Midjourney, Suno, Video, etc.).
// This is a stub for future implementation.
type TaskBilling struct {
	TaskID     string
	UserID     int
	Quota      int64
	ModelName  string
}

// RecordTaskConsumption records the quota consumption for an async task.
// TODO: implement actual task billing logic
func RecordTaskConsumption(taskID string, userID int, quota int64, modelName string) error {
	// RecordConsumeLog(userID, 0, 0, 0, modelName, "", int(quota), "task:"+taskID)
	return nil
}
