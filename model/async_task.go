package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// AsyncTaskPlatform 异步任务平台类型
type AsyncTaskPlatform string

const (
	PlatformMidjourney AsyncTaskPlatform = "midjourney"
	PlatformVideo      AsyncTaskPlatform = "video"
	PlatformSuno       AsyncTaskPlatform = "suno"
	PlatformKling      AsyncTaskPlatform = "kling"
	PlatformJimeng     AsyncTaskPlatform = "jimeng"
)

// AsyncTaskStatus 异步任务状态
type AsyncTaskStatus string

const (
	TaskStatusPending    AsyncTaskStatus = "pending"     // 等待处理
	TaskStatusQueued     AsyncTaskStatus = "queued"      // 已排队
	TaskStatusProcessing AsyncTaskStatus = "processing"  // 处理中
	TaskStatusSuccess    AsyncTaskStatus = "success"      // 成功
	TaskStatusFailed     AsyncTaskStatus = "failed"       // 失败
	TaskStatusCancelled  AsyncTaskStatus = "cancelled"   // 已取消
)

// AsyncTask 通用异步任务模型
type AsyncTask struct {
	ID           int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	TaskID       string           `json:"task_id" gorm:"type:varchar(191);uniqueIndex"` // 外部任务ID（平台返回）
	Platform     AsyncTaskPlatform `json:"platform" gorm:"type:varchar(30);index"`        // 平台：midjourney/video/suno
	UserID       int              `json:"user_id" gorm:"index"`
	ChannelID    int              `json:"channel_id" gorm:"index"`
	Group        string           `json:"group" gorm:"type:varchar(50);default:'default'"`
	Action       string           `json:"action" gorm:"type:varchar(40);index"` // 操作类型：imagine/upscale/vary/generate
	Status       AsyncTaskStatus  `json:"status" gorm:"type:varchar(20);index;default:'pending'"`
	Progress     int              `json:"progress" gorm:"default:0"` // 进度 0-100
	Quota        int              `json:"quota" gorm:"default:0"`    // 消耗配额
	FailReason   string           `json:"fail_reason" gorm:"type:text"`
	SubmitTime   int64            `json:"submit_time" gorm:"index"` // 提交时间（Unix秒）
	StartTime    int64            `json:"start_time" gorm:"index"`  // 开始处理时间
	FinishTime   int64            `json:"finish_time" gorm:"index"` // 完成时间
	RequestData  string           `json:"request_data" gorm:"type:text"`  // 请求数据（JSON字符串）
	ResponseData string           `json:"response_data" gorm:"type:text"` // 响应数据（JSON字符串）
	ImageURL     string           `json:"image_url" gorm:"type:text"`     // 结果图片URL（MJ用）
	VideoURL     string           `json:"video_url" gorm:"type:text"`     // 结果视频URL（视频用）
	AudioURL     string           `json:"audio_url" gorm:"type:text"`     // 结果音频URL（Suno用）
}

// TableName 指定表名
func (AsyncTask) TableName() string {
	return "async_tasks"
}

// IsFinished 判断任务是否已完成（成功或失败）
func (t *AsyncTask) IsFinished() bool {
	return t.Status == TaskStatusSuccess || t.Status == TaskStatusFailed || t.Status == TaskStatusCancelled
}

// SetRequestData 设置请求数据（自动序列化为JSON字符串）
func (t *AsyncTask) SetRequestData(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	t.RequestData = string(bytes)
	return nil
}

// GetRequestData 获取请求数据（从JSON字符串反序列化）
func (t *AsyncTask) GetRequestData(v interface{}) error {
	if t.RequestData == "" {
		return nil
	}
	return json.Unmarshal([]byte(t.RequestData), v)
}

// SetResponseData 设置响应数据
func (t *AsyncTask) SetResponseData(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	t.ResponseData = string(bytes)
	return nil
}

// GetResponseData 获取响应数据
func (t *AsyncTask) GetResponseData(v interface{}) error {
	if t.ResponseData == "" {
		return nil
	}
	return json.Unmarshal([]byte(t.ResponseData), v)
}

// MidjourneyTask Midjourney 专用任务扩展（存在则关联查询）
type MidjourneyTask struct {
	ID          int64    `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	TaskID      string    `json:"task_id" gorm:"type:varchar(191);uniqueIndex"` // 对应 AsyncTask.ID 或外部ID
	MJID        string    `json:"mj_id" gorm:"type:varchar(191);index"`        // Midjourney 返回的任务ID
	Prompt      string    `json:"prompt" gorm:"type:text"`                      // 原始提示词
	PromptEn    string    `json:"prompt_en" gorm:"type:text"`                  // 英文提示词（翻译后）
	Description string    `json:"description"`
	State       string    `json:"state" gorm:"type:varchar(500)"` // 自定义状态（回调用）
	Buttons     string    `json:"buttons" gorm:"type:text"`        // 可用按钮JSON
	Properties  string    `json:"properties" gorm:"type:text"`    // 扩展属性JSON
}

// TableName 指定表名
func (MidjourneyTask) TableName() string {
	return "midjourney_tasks"
}

// VideoTask 视频生成任务扩展
type VideoTask struct {
	ID            int64    `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	TaskID        string    `json:"task_id" gorm:"type:varchar(191);uniqueIndex"`
	PlatformTaskID string   `json:"platform_task_id" gorm:"type:varchar(191);index"` // 平台任务ID
	Model         string    `json:"model" gorm:"type:varchar(100)"`                   // 模型：kling/jimeng等
	Prompt        string    `json:"prompt" gorm:"type:text"`
	ImageURL      string    `json:"image_url" gorm:"type:text"`  // 参考图片URL
	VideoURLs     string    `json:"video_urls" gorm:"type:text"` // 结果视频URL列表（JSON数组）
}

// TableName 指定表名
func (VideoTask) TableName() string {
	return "video_tasks"
}

// SunoTask Suno 音乐生成任务扩展
type SunoTask struct {
	ID         int64    `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	TaskID     string    `json:"task_id" gorm:"type:varchar(191);uniqueIndex"`
	SunoTaskID string    `json:"suno_task_id" gorm:"type:varchar(191);index"` // Suno返回的任务ID
	Action     string    `json:"action" gorm:"type:varchar(40)"`                // song/lyrics/description-mode
	Title      string    `json:"title" gorm:"type:varchar(500)"`
	Lyrics     string    `json:"lyrics" gorm:"type:text"`
	AudioURL   string    `json:"audio_url" gorm:"type:text"`
	ImageURL   string    `json:"image_url" gorm:"type:text"` // 封面URL
	Model      string    `json:"model" gorm:"type:varchar(100)"`
}

// TableName 指定表名
func (SunoTask) TableName() string {
	return "suno_tasks"
}

// ==================== 数据库操作 ====================

// CreateAsyncTask 创建异步任务
func CreateAsyncTask(task *AsyncTask) error {
	return DB.Create(task).Error
}

// GetAsyncTaskByID 根据ID获取任务
func GetAsyncTaskByID(id int64) (*AsyncTask, error) {
	var task AsyncTask
	err := DB.First(&task, "id = ?", id).Error
	return &task, err
}

// GetAsyncTaskByTaskID 根据外部TaskID获取任务
func GetAsyncTaskByTaskID(taskID string) (*AsyncTask, error) {
	var task AsyncTask
	err := DB.First(&task, "task_id = ?", taskID).Error
	return &task, err
}

// GetUserTasks 获取用户的任务列表
func GetUserTasks(userID int, platform string, status string, startIdx int, num int) ([]*AsyncTask, error) {
	var tasks []*AsyncTask
	query := DB.Where("user_id = ?", userID)

	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("id desc").Limit(num).Offset(startIdx).Find(&tasks).Error
	return tasks, err
}

// GetAllUnfinishedTasks 获取所有未完成的任务（用于轮询）
func GetAllUnfinishedTasks() ([]*AsyncTask, error) {
	var tasks []*AsyncTask
	err := DB.Where("status IN ?", []string{
		string(TaskStatusPending),
		string(TaskStatusQueued),
		string(TaskStatusProcessing),
	}).Order("id asc").Find(&tasks).Error
	return tasks, err
}

// UpdateAsyncTaskStatus 更新任务状态
func UpdateAsyncTaskStatus(taskID string, status AsyncTaskStatus, progress int, failReason string) error {
	updates := map[string]interface{}{
		"status":   status,
		"progress": progress,
	}
	if failReason != "" {
		updates["fail_reason"] = failReason
	}
	if status == TaskStatusProcessing {
		updates["start_time"] = time.Now().Unix()
	}
	if status == TaskStatusSuccess || status == TaskStatusFailed {
		updates["finish_time"] = time.Now().Unix()
	}
	return DB.Model(&AsyncTask{}).Where("task_id = ?", taskID).Updates(updates).Error
}

// DeleteAsyncTask 删除任务
func DeleteAsyncTask(taskID string) error {
	return DB.Where("task_id = ?", taskID).Delete(&AsyncTask{}).Error
}

// ==================== Midjourney 任务操作 ====================

// CreateMidjourneyTask 创建Midjourney任务
func CreateMidjourneyTask(task *MidjourneyTask) error {
	return DB.Create(task).Error
}

// GetMidjourneyTaskByMJID 根据MJ ID获取任务
func GetMidjourneyTaskByMJID(mjID string) (*MidjourneyTask, error) {
	var task MidjourneyTask
	err := DB.First(&task, "mj_id = ?", mjID).Error
	return &task, err
}

// GetMidjourneyTaskByTaskID 根据TaskID获取Midjourney任务
func GetMidjourneyTaskByTaskID(taskID string) (*MidjourneyTask, error) {
	var task MidjourneyTask
	err := DB.First(&task, "task_id = ?", taskID).Error
	return &task, err
}

// GetAllMidjourneyTasks 获取所有MJ任务（带过滤）
func GetAllMidjourneyTasks(startIdx int, num int, mjID string, startTimestamp string, endTimestamp string) ([]*MidjourneyTask, error) {
	var tasks []*MidjourneyTask
	query := DB

	if mjID != "" {
		query = query.Where("mj_id = ?", mjID)
	}
	if startTimestamp != "" {
		query = query.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != "" {
		query = query.Where("created_at <= ?", endTimestamp)
	}

	err := query.Order("id desc").Limit(num).Offset(startIdx).Find(&tasks).Error
	return tasks, err
}

// GetUserMidjourneyTasks 获取用户的MJ任务
func GetUserMidjourneyTasks(userID int, startIdx int, num int, mjID string) ([]*MidjourneyTask, error) {
	var tasks []*MidjourneyTask
	query := DB.Where("task_id IN (SELECT task_id FROM async_tasks WHERE user_id = ?)", userID)

	if mjID != "" {
		query = query.Where("mj_id = ?", mjID)
	}

	err := query.Order("id desc").Limit(num).Offset(startIdx).Find(&tasks).Error
	return tasks, err
}

// UpdateMidjourneyTask 更新MJ任务
func UpdateMidjourneyTask(mjID string, updates map[string]interface{}) error {
	return DB.Model(&MidjourneyTask{}).Where("mj_id = ?", mjID).Updates(updates).Error
}

// ==================== JSON 辅助类型 ====================

// JSONMap 通用JSON字段类型（用于gorm的json类型）
type JSONMap map[string]interface{}

// Scan 实现sql.Scanner接口
func (m *JSONMap) Scan(val interface{}) error {
	if val == nil {
		*m = make(JSONMap)
		return nil
	}
	bytes, ok := val.([]byte)
	if !ok {
		return nil
	}
	if len(bytes) == 0 {
		*m = make(JSONMap)
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// Value 实现driver.Valuer接口
func (m JSONMap) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}
	return json.Marshal(m)
}

// ==================== Video 任务操作 ====================

// CreateVideoTask 创建视频生成任务
func CreateVideoTask(task *VideoTask) error {
	return DB.Create(task).Error
}

// GetVideoTaskByTaskID 根据TaskID获取视频任务
func GetVideoTaskByTaskID(taskID string) (*VideoTask, error) {
	var task VideoTask
	err := DB.First(&task, "task_id = ?", taskID).Error
	return &task, err
}

// UpdateVideoTask 更新视频任务
func UpdateVideoTask(taskID string, updates map[string]interface{}) error {
	return DB.Model(&VideoTask{}).Where("task_id = ?", taskID).Updates(updates).Error
}

// ==================== Suno 任务操作 ====================

// CreateSunoTask 创建Suno音乐生成任务
func CreateSunoTask(task *SunoTask) error {
	return DB.Create(task).Error
}

// GetSunoTaskByTaskID 根据TaskID获取Suno任务
func GetSunoTaskByTaskID(taskID string) (*SunoTask, error) {
	var task SunoTask
	err := DB.First(&task, "task_id = ?", taskID).Error
	return &task, err
}

// UpdateSunoTask 更新Suno任务
func UpdateSunoTask(taskID string, updates map[string]interface{}) error {
	return DB.Model(&SunoTask{}).Where("task_id = ?", taskID).Updates(updates).Error
}
