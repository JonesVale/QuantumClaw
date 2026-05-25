package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// StartDailyModelSync 每天零点自动从上游 API 同步模型列表到数据库
func StartDailyModelSync() {
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Add(24 * time.Hour)
	duration := nextMidnight.Sub(now)

	logger.SysLog(fmt.Sprintf("daily model sync scheduled: first run at %s (in %v)", nextMidnight.Format("15:04:05"), duration))
	time.Sleep(duration)

	for {
		syncAllChannels()
		time.Sleep(24 * time.Hour)
	}
}

func syncAllChannels() {
	logger.SysLog("daily model sync: starting...")

	var channels []model.Channel
	model.DB.Find(&channels)

	successCount := 0
	failCount := 0
	newModelCount := 0

	for _, ch := range channels {
		if ch.Key == "" || len(ch.Key) < 8 || ch.Status != model.ChannelStatusEnabled {
			continue
		}

		// Step 1: 从上游拉取模型名列表，更新 channel.Models
		oldModels := ch.Models
		err := ch.UpdateModelsFromProvider()
		if err != nil {
			logger.SysWarn(fmt.Sprintf("daily model sync: channel #%d %s fetch failed: %v", ch.Id, ch.Name, err))
			failCount++
			continue
		}
		successCount++

		// Step 2: 解析模型名列表
		modelNames := parseModelNames(ch.Models)
		if len(modelNames) == 0 {
			continue
		}

		// Step 3: 检查 model_metadata，不存在则新增
		for _, mn := range modelNames {
			var count int64
			model.DB.Model(&model.ModelMetadata{}).Where("model_name = ? AND languages_type = ?", mn, "English").Count(&count)
			if count > 0 {
				continue // 已存在，跳过
			}

			// 新增 model_metadata 记录（7种语言各一条）
			now := time.Now().Unix()
			languages := []string{"中文简体", "中文繁体", "English", "Fran\u00e7ais", "\u65e5\u672c\u8a9e", "\u0420\u0443\u0441\u0441\u043a\u0438\u0439", "Ti\u1ebfng Vi\u1ec7t"}
			provider := ch.Name
			if p, ok := channelTypeNameMap[ch.Type]; ok {
				provider = p
			}

			for _, lang := range languages {
				entry := &model.ModelMetadata{
					ModelName:     mn,
					LanguagesType: lang,
					DisplayName:   mn,
					Description:   "",
					UseCase:       "",
					ContextWindow: 0,
					Provider:      provider,
					CreatedTime:   now,
					UpdatedTime:   now,
				}
				if err := model.DB.Create(entry).Error; err != nil {
					logger.SysWarn(fmt.Sprintf("daily model sync: create metadata failed for %s (%s): %v", mn, lang, err))
				}
			}
			newModelCount++
			logger.SysLog(fmt.Sprintf("daily model sync: new model added: %s (provider: %s)", mn, provider))
		}

		// 记录模型数量变化
		oldCount := len(parseModelNames(oldModels))
		newCount := len(modelNames)
		if newCount != oldCount {
			logger.SysLog(fmt.Sprintf("daily model sync: channel #%d %s models changed: %d -> %d", ch.Id, ch.Name, oldCount, newCount))
		}
	}

	logger.SysLog(fmt.Sprintf("daily model sync: completed (%d success, %d failed, %d new models)",
		successCount, failCount, newModelCount))
}

func parseModelNames(modelsStr string) []string {
	var result []string
	for _, m := range strings.Split(modelsStr, ",") {
		m = strings.TrimSpace(m)
		if m != "" {
			result = append(result, m)
		}
	}
	return result
}

// channelTypeNameMap maps channel type IDs to provider names
var channelTypeNameMap = map[int]string{
	1:   "OpenAI",
	14:  "Anthropic",
	24:  "Google",
	36:  "DeepSeek",
	29:  "Groq",
	44:  "SiliconFlow",
	25:  "Moonshot",
	17:  "Alibaba",
	15:  "Baidu",
	16:  "Zhipu AI",
	23:  "Tencent",
	27:  "MiniMax",
	28:  "Mistral",
	35:  "Cohere",
	37:  "Cloudflare",
	38:  "DeepL",
	39:  "Together AI",
	45:  "xAI",
	30:  "Ollama",
	33:  "AWS",
	100: "IonQ",
	101: "IBM",
	102: "Rigetti",
	103: "AWS Braket",
	104: "Azure Quantum",
	105: "Google Quantum",
}
