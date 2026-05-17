package constant

type TaskPlatform string

const (
	TaskPlatformMidjourney TaskPlatform = "midjourney"
	TaskPlatformVideo      TaskPlatform = "video"
	TaskPlatformSuno       TaskPlatform = "suno"
	TaskPlatformMJProxy    TaskPlatform = "mj_proxy"
)

var UpdateTask = true
