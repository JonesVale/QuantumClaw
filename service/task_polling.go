package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

type TaskPollingAdaptor interface {
	Init(info interface{})
	FetchTask(baseURL string, key string, body map[string]any, proxy string) (*TaskPollResponse, error)
	ParseTaskResult(body []byte) (*TaskInfo, error)
	AdjustBillingOnComplete(task *Task, taskResult *TaskInfo) int
}

type TaskPollResponse struct {
	StatusCode int
	Body       []byte
}

type TaskInfo struct {
	TaskID      string
	Status      string
	Url         string
	Progress    string
	Reason      string
	TotalTokens int
}

type Task struct {
	ID          int64
	TaskID      string
	Status      string
	Progress    string
	FinishTime  int64
	FailReason  string
	Quota       int
	Platform    string
	ChannelId   int
}

var GetTaskAdaptorFunc func(platform string) TaskPollingAdaptor

func sweepTimedOutTasks(ctx context.Context) {
	taskService := NewTaskService()
	count := taskService.PollPendingTasks()
	if count > 0 {
		logger.SysLog(fmt.Sprintf("轮询了 %d 个异步任务", count))
	}
}

func TaskPollingLoop() {
	for {
		time.Sleep(time.Duration(15) * time.Second)
		logger.SysLog("任务进度轮询开始")
		ctx := context.TODO()
		sweepTimedOutTasks(ctx)
		logger.SysLog("任务进度轮询完成")
	}
}

func DispatchPlatformUpdate(platform string, taskChannelM map[int][]string, taskM map[string]*Task) {
	switch platform {
	case "midjourney":
	case "suno":
		_ = UpdateSunoTasks(context.Background(), taskChannelM, taskM)
	default:
		if err := UpdateVideoTasks(context.Background(), platform, taskChannelM, taskM); err != nil {
			logger.SysLog(fmt.Sprintf("UpdateVideoTasks fail: %s", err))
		}
	}
}

func UpdateSunoTasks(ctx context.Context, taskChannelM map[int][]string, taskM map[string]*Task) error {
	for channelId, taskIds := range taskChannelM {
		err := updateSunoTasks(ctx, channelId, taskIds, taskM)
		if err != nil {
			logger.SysLog(fmt.Sprintf("渠道 #%d 更新异步任务失败: %s", channelId, err.Error()))
		}
	}
	return nil
}

func updateSunoTasks(ctx context.Context, channelId int, taskIds []string, taskM map[string]*Task) error {
	if len(taskIds) == 0 {
		return nil
	}
	adaptor := GetTaskAdaptorFunc("suno")
	if adaptor == nil {
		return errors.New("adaptor not found")
	}
	return nil
}

func UpdateVideoTasks(ctx context.Context, platform string, taskChannelM map[int][]string, taskM map[string]*Task) error {
	for channelId, taskIds := range taskChannelM {
		if err := updateVideoTasks(ctx, platform, channelId, taskIds, taskM); err != nil {
			logger.SysLog(fmt.Sprintf("Channel #%d failed to update video async tasks: %s", channelId, err.Error()))
		}
	}
	return nil
}

func updateVideoTasks(ctx context.Context, platform string, channelId int, taskIds []string, taskM map[string]*Task) error {
	if len(taskIds) == 0 {
		return nil
	}
	adaptor := GetTaskAdaptorFunc(platform)
	if adaptor == nil {
		return fmt.Errorf("video adaptor not found")
	}
	return nil
}

func RefundTaskQuota(ctx context.Context, task *Task, reason string) {
	if task.Quota == 0 {
		return
	}
	logger.SysLog(fmt.Sprintf("Refunding task %s quota: %d, reason: %s", task.TaskID, task.Quota, reason))
}

func RecalculateTaskQuota(ctx context.Context, task *Task, actualQuota int, reason string) {
	diff := actualQuota - task.Quota
	if diff > 0 {
		logger.SysLog(fmt.Sprintf("补扣任务 %s 额度: %d", task.TaskID, diff))
	} else if diff < 0 {
		logger.SysLog(fmt.Sprintf("退还任务 %s 额度: %d", task.TaskID, -diff))
	}
}

func RecalculateTaskQuotaByTokens(ctx context.Context, task *Task, tokens int) {
	logger.SysLog(fmt.Sprintf("按token重算任务 %s 额度: %d tokens", task.TaskID, tokens))
}