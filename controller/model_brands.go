package controller

import (
	"sort"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/encrypt"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
	"time"

	"encoding/json"
	"io"
	"os"
)

// 鈹€鈹€ 鍝佺墝鍒版笭閬撶被鍨嬬殑鏄犲皠 鈹€鈹€
// 灏嗗湪閰嶇疆鏃惰嚜鍔ㄥ～鍏?BaseURL銆佹ā鍨嬪垪琛ㄧ瓑

// BrandConfig 鎻忚堪涓€涓搧鐗屾彁渚涘晢鐨勬妧鏈厤缃?
type BrandConfig struct {
	ProviderName  string   `json:"provider_name"`  // model_metadata.Provider 鍊?
	ChannelType   int      `json:"channel_type"`   // channeltype.XXX 甯搁噺
	DefaultURL    string   `json:"default_url"`    // 鑷姩濉厖鐨?BaseURL
	ChannelName   string   `json:"channel_name"`   // 娓犻亾鏄剧ず鍚嶇О
	AutoFields    []string `json:"auto_fields"`    // 鍙嚜鍔ㄥ～鍏呯殑瀛楁鍚?
	RequiredUser  []string `json:"required_user"`  // 蹇呴』鐢ㄦ埛濉殑锛堣嚦灏?API Key锛?
	OptionalUser  []string `json:"optional_user"`  // 鍙€夌敤鎴峰～
	Notes         string   `json:"notes"`          // 鎻愮ず璇?
}

// 鍝佺墝閰嶇疆娉ㄥ唽琛?
// 更新 brandConfigs 后需同步更新 GetModelBrands 中的 ProviderName 常量
var brandConfigs = []BrandConfig{
	{
		ProviderName: "OpenAI", ChannelType: channeltype.OpenAI,
		DefaultURL: "https://api.openai.com", ChannelName: "OpenAI",
		AutoFields:   []string{"base_url", "models"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill your OpenAI API key. Models: GPT-4o, GPT-4o-mini, GPT-4, GPT-3.5-turbo",
	},
	{
		ProviderName: "Anthropic", ChannelType: channeltype.Anthropic,
		DefaultURL: "https://api.anthropic.com", ChannelName: "Anthropic",
		AutoFields:   []string{"base_url", "models"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill your Anthropic API key. Models: Claude 3.5 Sonnet, Claude 3 Opus, etc.",
	},
	{
		ProviderName: "Google", ChannelType: channeltype.Gemini,
		DefaultURL: "https://generativelanguage.googleapis.com", ChannelName: "Google Gemini",
		AutoFields:   []string{"base_url", "models"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill your Google AI Studio API key. Models: Gemini 2.0 Flash, Gemini 2.5 Pro, etc.",
	},
	{
		ProviderName: "DeepSeek", ChannelType: channeltype.DeepSeek,
		DefaultURL: "https://api.deepseek.com", ChannelName: "DeepSeek",
		AutoFields:   []string{"base_url", "models"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill your DeepSeek API key. Models: DeepSeek-V3, DeepSeek-R1",
	},
	{
		ProviderName: "Alibaba", ChannelType: channeltype.Ali,
		DefaultURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", ChannelName: "Ali (Qwen)",
		AutoFields:   []string{"base_url", "models"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill your Alibaba DashScope API key. Models: Qwen Turbo, Qwen Plus, Qwen Max",
	},
	{
		ProviderName: "Mistral", ChannelType: channeltype.Mistral,
		DefaultURL: "https://api.mistral.ai", ChannelName: "Mistral AI",
		AutoFields:   []string{"base_url", "models"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill your Mistral API key. Models: Mistral Large, Mistral Nemo",
	},
	// Meta (Llama models) - open source, can use TogetherAI as main provider
	{
		ProviderName: "Meta", ChannelType: channeltype.TogetherAI,
		DefaultURL: "https://api.together.xyz", ChannelName: "Together AI (Llama)",
		AutoFields:   []string{"base_url"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill Together AI API key for Llama models. Or use Ollama/OpenAI Compatible for local deployment.",
	},
	// Quantum providers
	{
		ProviderName: "IonQ", ChannelType: channeltype.IonQ,
		DefaultURL: "https://api.ionq.co", ChannelName: "IonQ",
		AutoFields:   []string{"base_url"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill your IonQ API key for quantum computing access.",
	},
	{
		ProviderName: "IBM", ChannelType: channeltype.IBMQ,
		DefaultURL: "https://api.quantum.ibm.com", ChannelName: "IBM Q",
		AutoFields:   []string{"base_url"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill your IBM Quantum API key.",
	},
	{
		ProviderName: "Rigetti", ChannelType: channeltype.Rigetti,
		DefaultURL: "https://api.qcs.rigetti.com", ChannelName: "Rigetti",
		AutoFields:   []string{"base_url"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill your Rigetti Quantum Cloud Services API key.",
	},
	{
		ProviderName: "Zhipu", ChannelType: channeltype.Zhipu,
		DefaultURL: "https://open.bigmodel.cn/api/paas/v4", ChannelName: "Zhipu AI",
		AutoFields:   []string{"base_url", "models"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill your Zhipu AI (GLM) API key.",
	},
	{
		ProviderName: "Baidu", ChannelType: channeltype.Baidu,
		DefaultURL: "https://aip.baidubce.com", ChannelName: "Baidu Qianfan",
		AutoFields:   []string{"base_url", "models"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill your Baidu Qianfan API key (access_token).",
	},
	{
		ProviderName: "Tencent", ChannelType: channeltype.Tencent,
		DefaultURL: "https://api.hunyuan.cloud.tencent.com", ChannelName: "Tencent Hunyuan",
		AutoFields:   []string{"base_url", "models"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill your Tencent Hunyuan API key.",
	},
	{
		ProviderName: "Moonshot", ChannelType: channeltype.Moonshot,
		DefaultURL: "https://api.moonshot.cn", ChannelName: "Moonshot/Kimi",
		AutoFields:   []string{"base_url", "models"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill your Moonshot (Kimi) API key.",
	},
	{
		ProviderName: "Groq", ChannelType: channeltype.Groq,
		DefaultURL: "https://api.groq.com/openai", ChannelName: "Groq",
		AutoFields:   []string{"base_url", "models"},
		RequiredUser: []string{"api_key"},
		Notes:        "Fill your Groq API key. Models: Mixtral, Llama, Gemma",
	},
}

// 鈹€鈹€ Response types 鈹€鈹€

// BrandInfo 杩斿洖缁欏墠绔殑鍝佺墝瀹屾暣淇℃伅
type BrandInfo struct {
	ProviderName  string           `json:"provider_name"`
	ChannelType   int              `json:"channel_type"`
	ChannelName   string           `json:"channel_name"`
	DefaultURL    string           `json:"default_url"`
	AutoFields    []string         `json:"auto_fields"`
	RequiredUser  []string         `json:"required_user"`
	Notes         string           `json:"notes"`
	IsConfigured  bool             `json:"is_configured"`
	ChannelID     int              `json:"channel_id"`
	ExistingKey   string           `json:"existing_key"`
	Models        []BrandModelInfo `json:"models"`        // 璇ュ搧鐗屼笅鐨勬ā鍨嬪垪琛?
	MissingFields []string         `json:"missing_fields"` // 鑷姩濉厖涓嶄簡锛岄渶瑕佺敤鎴疯ˉ鐨?
}

type BrandModelInfo struct {
	Name         string  `json:"name"`
	DisplayName  string  `json:"display_name"`
	Description  string  `json:"description"`
	ContextWindow int    `json:"context_window"`
	InputPrice   float64 `json:"input_price"`
	OutputPrice  float64 `json:"output_price"`
}

// 鈹€鈹€ Controller 鈹€鈹€

// GetModelBrands 杩斿洖鎵€鏈夋ā鍨嬪搧鐗屽強閰嶇疆鐘舵€?
// GET /api/admin/model-brands
func GetModelBrands(c *gin.Context) {
	// 1. 浠?model_metadata 璇诲彇鎵€鏈夋ā鍨嬫彁渚涘晢锛圖ISTINCT锛?
	var providers []struct{ Provider string }
	model.DB.Model(&model.ModelMetadata{}).
		Where("provider != '' AND provider != 'Other' AND languages_type = ?", "English").
		Select("DISTINCT provider").
		Find(&providers)

	// 2. 璇诲彇鎵€鏈夋ā鍨嬫寜 provider 鍒嗙粍
	var metadata []model.ModelMetadata
	model.DB.Where("languages_type = ?", "English").
		Find(&metadata)
	modelsByProvider := make(map[string][]model.ModelMetadata)
	for _, m := range metadata {
		if m.Provider != "" {
			modelsByProvider[m.Provider] = append(modelsByProvider[m.Provider], m)
		}
	}

	// 3. 璇诲彇鐜版湁娓犻亾锛堟寜 name 绱㈠紩锛屽洜涓哄姩鎬佸搧鐗屾病鏈夊浐瀹?type锛?
	var channels []model.Channel
	model.DB.Find(&channels)
	channelByName := make(map[string]*model.Channel)
	for i := range channels {
		channelByName[strings.ToLower(channels[i].Name)] = &channels[i]
	}

	// 4. 鏋勫缓宸茬煡鍝佺墝閰嶇疆鏌ユ壘琛?
		// Load provider overrides from DB (channel_providers table)
	var channelProviders []model.ChannelProvider
	model.DB.Order("type_id asc").Find(&channelProviders)
	providerURLOverride := make(map[int]string)
	providerModelsOverride := make(map[string][]string)
	for _, cp := range channelProviders {
		if cp.BaseURL != "" {
			providerURLOverride[cp.TypeID] = cp.BaseURL
		}
		if cp.ProviderSlug != "" && len(cp.GetModels()) > 0 {
			providerModelsOverride[cp.ProviderSlug] = cp.GetModels()
		}
	}

	knownConfigs := make(map[string]*BrandConfig)
	for i := range brandConfigs {
		knownConfigs[strings.ToLower(brandConfigs[i].ProviderName)] = &brandConfigs[i]
	}

	var brands []BrandInfo

	for _, cfg := range brandConfigs {
		providerName := cfg.ProviderName
		if _, exists := modelsByProvider[providerName]; !exists {
		}

		// Override URL from DB if available
		if overrideURL, ok := providerURLOverride[cfg.ChannelType]; ok {
			cfg.DefaultURL = overrideURL
		}

		// Override model list from DB if available
		providerModels := modelsByProvider[providerName]
		if overrideModels, ok := providerModelsOverride[cfg.ProviderName]; ok {
			filteredModels := make([]model.ModelMetadata, 0)
			for _, pm := range providerModels {
				for _, om := range overrideModels {
					if strings.EqualFold(pm.ModelName, om) {
						filteredModels = append(filteredModels, pm)
						break
					}
				}
			}
			if len(filteredModels) > 0 {
				providerModels = filteredModels
			}
		}

		bi := buildBrandInfoFromConfig(&cfg, providerModels, channelByName)
		brands = append(brands, bi)
	}

	// 鍐嶅鐞嗗姩鎬佸彂鐜扮殑鏈煡鎻愪緵鍟?
	for _, p := range providers {
		pn := p.Provider
		if _, isKnown := knownConfigs[strings.ToLower(pn)]; isKnown {
			continue // 宸插鐞?
		}
		if _, exists := modelsByProvider[pn]; !exists {
			continue
		}
		// 鑷姩鐢熸垚鍝佺墝閰嶇疆
		genCfg := generateBrandConfig(pn)
		bi := buildBrandInfoFromConfig(&genCfg, modelsByProvider[pn], channelByName)
		brands = append(brands, bi)
	}

	// 鎺掑簭锛氬凡閰嶇疆鐨勬帓鍓嶉潰
	sort.SliceStable(brands, func(i, j int) bool {
		if brands[i].IsConfigured != brands[j].IsConfigured {
			return brands[i].IsConfigured
		}
		return brands[i].ProviderName < brands[j].ProviderName
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    brands,
	})
}

// buildBrandInfoFromConfig 浠?BrandConfig + 妯″瀷鍒楄〃 + 娓犻亾鏋勫缓 BrandInfo
func buildBrandInfoFromConfig(cfg *BrandConfig, providerModels []model.ModelMetadata, channelByName map[string]*model.Channel) BrandInfo {
	bi := BrandInfo{
		ProviderName: cfg.ProviderName,
		ChannelType:  cfg.ChannelType,
		ChannelName:  cfg.ChannelName,
		DefaultURL:   cfg.DefaultURL,
		AutoFields:   cfg.AutoFields,
		RequiredUser: cfg.RequiredUser,
		Notes:        cfg.Notes,
	}

	// 濉厖妯″瀷鍒楄〃 + 鍙傝€冧环鏍?
	for _, pm := range providerModels {
		bmi := BrandModelInfo{
			Name:          pm.ModelName,
			DisplayName:   pm.DisplayName,
			Description:   pm.Description,
			ContextWindow: pm.ContextWindow,
		}
		if rp := model.GetReferencePrice(pm.ModelName); rp != nil {
			bmi.InputPrice = rp.InputPrice
			bmi.OutputPrice = rp.OutputPrice
		}
		bi.Models = append(bi.Models, bmi)
	}

	// 妫€鏌ユ笭閬撻厤缃紙鎸夋笭閬撳悕鏌ユ壘锛?
	channelKey := strings.ToLower(cfg.ChannelName)
	if ch, ok := channelByName[channelKey]; ok {
		bi.IsConfigured = ch.Key != "" && !strings.HasPrefix(ch.Key, "PUT_YOUR")
		bi.ChannelID = ch.Id
		if bi.IsConfigured {
			bi.ExistingKey = maskKey(ch.Key)
		}
	}

	// 鐗规畩锛氭鏌ユ槸鍚︽寜 type 绱㈠紩涔熸湁娓犻亾锛堝吋瀹规棫娓犻亾锛?
	if !bi.IsConfigured {
		var chByType model.Channel
		if err := model.DB.Where("type = ?", cfg.ChannelType).First(&chByType).Error; err == nil {
			bi.IsConfigured = chByType.Key != "" && !strings.HasPrefix(chByType.Key, "PUT_YOUR")
			bi.ChannelID = chByType.Id
			if bi.IsConfigured {
				bi.ExistingKey = maskKey(chByType.Key)
			}
		}
	}

	// 鐗规畩鎻愮ず锛歁eta/Llama 璧?TogetherAI
	if cfg.ChannelType == channeltype.TogetherAI && cfg.ProviderName == "Meta" {
		bi.MissingFields = []string{
			"model_mapping",
			"Llama models are open-source and available through multiple providers. You may need to update model IDs manually in the channel settings.",
		}
	}

	return bi
}

// generateBrandConfig 涓烘湭鐭ユ彁渚涘晢鑷姩鐢熸垚鍝佺墝閰嶇疆
func generateBrandConfig(providerName string) BrandConfig {
	lowerName := strings.ToLower(providerName)
	chname := providerName
	ctype := channeltype.OpenAI   // 榛樿 OpenAI 鍏煎
	defURL := "https://api." + strings.ToLower(providerName) + ".com"
	notes := "Fill your " + providerName + " API key."

	// 灏濊瘯鏅鸿兘鐚滄祴娓犻亾绫诲瀷鍜岄粯璁?URL
	if strings.Contains(lowerName, "azure") {
		ctype = channeltype.Azure
		defURL = "https://YOUR_RESOURCE.openai.azure.com"
		notes = "Fill your Azure OpenAI endpoint and API key."
	} else if strings.Contains(lowerName, "baidu") || strings.Contains(lowerName, "ernie") {
		ctype = channeltype.Baidu
		defURL = "https://aip.baidubce.com"
		notes = "Fill your Baidu Qianfan API key (access_token)."
	} else if strings.Contains(lowerName, "glm") || strings.Contains(lowerName, "zhipu") || strings.Contains(lowerName, "chatglm") {
		ctype = channeltype.Zhipu
		defURL = "https://open.bigmodel.cn/api/paas/v4"
		notes = "Fill your Zhipu AI API key."
	} else if strings.Contains(lowerName, "tencent") || strings.Contains(lowerName, "hunyuan") {
		ctype = channeltype.Tencent
		defURL = "https://api.hunyuan.cloud.tencent.com"
		notes = "Fill your Tencent Hunyuan API key."
	} else if strings.Contains(lowerName, "moonshot") || strings.Contains(lowerName, "kimi") {
		ctype = channeltype.Moonshot
		defURL = "https://api.moonshot.cn"
		notes = "Fill your Moonshot/Kimi API key."
	} else if strings.Contains(lowerName, "cohere") {
		ctype = channeltype.Cohere
		defURL = "https://api.cohere.ai"
	} else if strings.Contains(lowerName, "groq") {
		ctype = channeltype.Groq
		defURL = "https://api.groq.com/openai/v1"
	} else if strings.Contains(lowerName, "together") {
		ctype = channeltype.TogetherAI
		defURL = "https://api.together.xyz"
	} else if strings.Contains(lowerName, "ollama") {
		ctype = channeltype.Ollama
		defURL = "http://localhost:11434"
	} else if strings.Contains(lowerName, "cloudflare") || strings.Contains(lowerName, "workers") {
		ctype = channeltype.Cloudflare
		defURL = "https://api.cloudflare.com/client/v4/accounts/YOUR_ACCOUNT/ai/v1"
	} else if strings.Contains(lowerName, "perplexity") {
		ctype = channeltype.OpenAICompatible
		defURL = "https://api.perplexity.ai"
	}

	return BrandConfig{
		ProviderName:  providerName,
		ChannelType:   ctype,
		DefaultURL:    defURL,
		ChannelName:   chname,
		AutoFields:    []string{"base_url", "models"},
		RequiredUser:  []string{"api_key"},
		Notes:         notes,
	}
}

// ConfigureModelBrand 閰嶇疆涓€涓搧鐗岀殑 API Key锛堜粎闇€ API Key锛屽叾浣欒嚜鍔ㄥ～鍏咃級
// POST /api/admin/model-brands/configure
func ConfigureModelBrand(c *gin.Context) {
	var req struct {
		ProviderName string `json:"provider_name" binding:"required"`
		APIKey       string `json:"api_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "provider_name and api_key are required"})
		return
	}

	// 鏌ユ壘鍝佺墝閰嶇疆
	var matchedCfg *BrandConfig
	for i := range brandConfigs {
		if strings.EqualFold(brandConfigs[i].ProviderName, req.ProviderName) {
			matchedCfg = &brandConfigs[i]
			break
		}
	}
	if matchedCfg == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "unknown provider: " + req.ProviderName})
		return
	}

	// 鏌ユ壘宸叉湁娓犻亾鎴栧垱寤烘柊娓犻亾
	var channel model.Channel
	result := model.DB.Where("type = ?", matchedCfg.ChannelType).First(&channel)

	if result.Error != nil {
		// 鍒涘缓鏂版笭閬?
		now := time.Now().Unix()
		baseURL := matchedCfg.DefaultURL
		if baseURL == "" {
			if matchedCfg.ChannelType >= 0 && matchedCfg.ChannelType < len(channeltype.ChannelBaseURLs) {
				baseURL = channeltype.ChannelBaseURLs[matchedCfg.ChannelType]
			}
		}
		models := getChannelModelsFromBrand(matchedCfg.ProviderName, matchedCfg.ChannelType)

		encKey := req.APIKey
		if config.CryptoSecret != "" {
			if ek, e := encrypt.EncryptChannelKey(req.APIKey, config.CryptoSecret); e == nil {
				encKey = ek
			}
		}
		newChannel := &model.Channel{
			Type:        matchedCfg.ChannelType,
			Name:        matchedCfg.ChannelName,
			Key:         encKey,
			BaseURL:     model.StrPtr(baseURL),
			Models:      models,
			Group:       "default",
			Status:      model.ChannelStatusEnabled,
			Weight:      model.UintPtr(1),
			CreatedTime: now,
		}
		if err := model.DB.Create(newChannel).Error; err != nil {
			logger.SysError("failed to create channel for brand: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "failed to create channel"})
			return
		}
		channel = *newChannel
		// 同步更新 Ability 表，否则 Distribute 找不到渠道
		if err := newChannel.AddAbilities(); err != nil {
			logger.SysError("ConfigureModelBrand: AddAbilities failed: " + err.Error())
		}
		// 自动发现模型：调用该品牌的 API 获取最新模型列表
		go func(chID int, chType int, chBaseURL string, chKey string, chName string) {
			if err := autoDiscoverModels(chID, chType, chBaseURL, chKey, chName); err != nil {
				logger.SysError("autoDiscoverModels failed for " + chName + ": " + err.Error())
			}
		}(channel.Id, channel.Type, *channel.BaseURL, channel.Key, channel.Name)
	} else {
		// 鏇存柊宸叉湁娓犻亾锛堝彧鏇存柊 Key锛?
		encryptedKey := req.APIKey
		if config.CryptoSecret != "" {
			if ek, e := encrypt.EncryptChannelKey(req.APIKey, config.CryptoSecret); e == nil {
				encryptedKey = ek
			}
		}
		updates := map[string]interface{}{
			"Key": encryptedKey,
		}
		// 濡傛灉 BaseURL 涓虹┖鍒欏～鍏呴粯璁?URL
		if channel.BaseURL == nil || *channel.BaseURL == "" {
			baseURL := matchedCfg.DefaultURL
			if baseURL == "" && matchedCfg.ChannelType >= 0 && matchedCfg.ChannelType < len(channeltype.ChannelBaseURLs) {
				baseURL = channeltype.ChannelBaseURLs[matchedCfg.ChannelType]
			}
			updates["base_url"] = baseURL
		}
		// 重新配置时始终更新模型列表（从 model_metadata 表动态获取最新模型）
		updates["models"] = getChannelModelsFromBrand(matchedCfg.ProviderName, matchedCfg.ChannelType)
		// 濡傛灉娓犻亾琚鐢紝閲嶆柊鍚敤
		updates["status"] = model.ChannelStatusEnabled

		if err := model.DB.Model(&channel).Updates(updates).Error; err != nil {
			logger.SysError("failed to update channel for brand: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "failed to update channel"})
			return
		}
		// 同步更新 Ability 表，否则 Distribute 找不到渠道
		if err := channel.UpdateAbilities(); err != nil {
			logger.SysError("ConfigureModelBrand: UpdateAbilities failed: " + err.Error())
		}
		// 自动发现模型：重新配置后也刷新模型列表
		baseURL := ""
		if channel.BaseURL != nil {
			baseURL = *channel.BaseURL
		}
		go func(chID int, chType int, chBaseURL string, chKey string, chName string) {
			if err := autoDiscoverModels(chID, chType, chBaseURL, chKey, chName); err != nil {
				logger.SysError("autoDiscoverModels failed for " + chName + ": " + err.Error())
			}
		}(channel.Id, channel.Type, baseURL, channel.Key, channel.Name)
	}

	// 杩斿洖閰嶇疆缁撴灉
	missingInfo := checkMissingInfo(matchedCfg)
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       fmt.Sprintf("%s configured successfully", req.ProviderName),
		"channel_id":    channel.Id,
		"auto_filled":   matchedCfg.AutoFields,
		"missing_info":  missingInfo,
		"channel_name":  matchedCfg.ChannelName,
	})
}

// getChannelModelsFromBrand 从 model_metadata 表获取该品牌的所有模型名
// 比 DefaultChannelModels 更及时，因为 model_metadata 会在启动时自动从 ratio 表填充
// autoDiscoverModels 调用品牌 API 的 /v1/models 端点自动发现可用模型
// 支持 OpenAI 兼容 API、DeepSeek、Anthropic 等主流品牌
// 新发现的模型自动加入渠道 models 字段和 model_metadata 表
func autoDiscoverModels(channelID int, channelType int, baseURL string, apiKey string, channelName string) error {
	if apiKey == "" || strings.HasPrefix(apiKey, "PUT_YOUR") {
		return nil // 没有真实 API Key，跳过
	}

	// 构建 /v1/models 请求 URL
	modelURL := baseURL
	if modelURL == "" {
		if channelType >= 0 && channelType < len(channeltype.ChannelBaseURLs) {
			modelURL = channeltype.ChannelBaseURLs[channelType]
		}
	}
	if modelURL == "" {
		return nil
	}
	// 确保 URL 格式正确
	modelURL = strings.TrimSuffix(modelURL, "/")
	if !strings.HasSuffix(modelURL, "/v1") {
		modelURL += "/v1"
	}
	modelURL += "/models"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", modelURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body[:min(200, len(body))]))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// 解析 OpenAI 兼容的 /v1/models 响应格式
	// {"object":"list","data":[{"id":"gpt-4o","object":"model",...},...]}
	var modelList struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &modelList); err != nil {
		// 尝试 Anthropic 格式
		var anthropicList struct {
			Data []struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			} `json:"data"`
		}
		if err2 := json.Unmarshal(body, &anthropicList); err2 != nil {
			return fmt.Errorf("parse response: %w", err)
		}
		modelList.Data = make([]struct { ID string `json:"id"` }, 0)
		for _, m := range anthropicList.Data {
			if m.ID != "" {
				modelList.Data = append(modelList.Data, struct { ID string `json:"id"` }{ID: m.ID})
			}
		}
	}

	if len(modelList.Data) == 0 {
		return nil // 没有模型，跳过
	}

	// 提取所有模型 ID
	discovered := make([]string, 0, len(modelList.Data))
	seen := make(map[string]bool)
	for _, m := range modelList.Data {
		id := strings.TrimSpace(m.ID)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		discovered = append(discovered, id)
	}

	// 更新渠道 models 字段：新增的加进去，已有的保留
	var channel model.Channel
	if err := model.DB.First(&channel, channelID).Error; err != nil {
		return fmt.Errorf("find channel: %w", err)
	}

	// 合并现有 models + 新发现 models
	existingModels := make(map[string]bool)
	if channel.Models != "" {
		for _, m := range strings.Split(channel.Models, ",") {
			m = strings.TrimSpace(m)
			if m != "" {
				existingModels[m] = true
			}
		}
	}

	added := 0
	for _, id := range discovered {
		if !existingModels[id] {
			existingModels[id] = true
			added++
		}
	}

	// 写回渠道 models
	newModels := make([]string, 0, len(existingModels))
	for m := range existingModels {
		newModels = append(newModels, m)
	}
	model.DB.Model(&channel).Update("models", strings.Join(newModels, ","))

	// 同步更新 Ability 表，否则渠道模型变更后 Distribute 找不到
	if err := channel.UpdateAbilities(); err != nil {
		logger.SysError("autoDiscoverModels: UpdateAbilities failed for " + channelName + ": " + err.Error())
	}

	// 新发现的模型加入 model_metadata（只增不删）
	providerName := channelName
	var metadataCount int64
	model.DB.Model(&model.ModelMetadata{}).Where("provider = ?", providerName).Count(&metadataCount)
	if metadataCount > 0 {
		providerName = channelName // 用渠道名作为 provider
	}

	now := time.Now().Unix()
	languages := []string{"中文简体", "中文繁体", "English", "Français", "日本語", "Русский", "Tiếng Việt"}
	metaAdded := 0

	for _, id := range discovered {
		// 检查是否已存在（按 model_name 去重）
		var existing int64
		model.DB.Model(&model.ModelMetadata{}).Where("model_name = ?", id).Count(&existing)
		if existing > 0 {
			continue
		}

		// 生成显示名
		displayName := id
		parts := strings.Split(id, "/")
		if len(parts) > 1 {
			displayName = parts[len(parts)-1]
		}
		displayParts := strings.Split(strings.ReplaceAll(displayName, "-", " "), " ")
		for i, p := range displayParts {
			if len(p) > 0 {
				displayParts[i] = strings.ToUpper(p[:1]) + p[1:]
			}
		}
		displayName = strings.Join(displayParts, " ")

		for _, lang := range languages {
			m := &model.ModelMetadata{
				ModelName:       id,
				LanguagesType:   lang,
				DisplayName:     displayName,
				Provider:        providerName,
				Description:     displayName + " - " + providerName + " (auto-discovered)",
				UseCase:         "chat",
				InputModalities: `["Text"]`,
				CreatedTime:     now,
				UpdatedTime:     now,
			}
			if err := model.DB.Create(m).Error; err != nil {
				continue
			}
			metaAdded++
		}
	}

	// 新发现的模型添加默认参考价格（如果还没有）
	priceAdded := 0
	for _, id := range discovered {
		if rp := model.GetReferencePrice(id); rp != nil {
			continue
		}
		rp := &model.ReferencePrice{
			ModelName:   id,
			Provider:    providerName,
			InputPrice:  0.0001, // 默认低价
			OutputPrice: 0.0002,
			Currency:    "USD",
			Source:      "auto_discovered",
			FetchedAt:   now,
		}
		if err := model.DB.Create(rp).Error; err != nil {
			continue
		}
		priceAdded++
	}

	if added > 0 || metaAdded > 0 || priceAdded > 0 {
		logger.SysLog(fmt.Sprintf("autoDiscoverModels[%s]: %d new models in channel, %d metadata entries, %d prices added",
			channelName, added, metaAdded, priceAdded))
	}

	return nil
}

func getChannelModelsFromBrand(providerName string, channelType int) string {
	var models []model.ModelMetadata
	model.DB.Where("provider = ? AND languages_type = ?", providerName, "English").
		Select("model_name, display_name").Find(&models)
	if len(models) > 0 {
		seen := make(map[string]bool)
		names := make([]string, 0, len(models))
		for _, m := range models {
			name := strings.TrimSpace(m.ModelName)
			if name == "" {
				continue
			}
			// 如果 model_name != display_name，说明 model_name 就是原始 API 调用名（deepseek-chat）
			// 如果相等，说明是旧种子数据，需要转换（DeepSeek Chat → deepseek-chat）
			finalName := name
			if name == m.DisplayName {
				// 旧种子：DeepSeek Chat → deepseek-chat
				finalName = strings.ToLower(name)
				finalName = strings.ReplaceAll(finalName, " ", "-")
			}
			// 去重（忽略大小写）
			key := strings.ToLower(finalName)
			if !seen[key] {
				seen[key] = true
				names = append(names, finalName)
			}
		}
		if len(names) > 0 {
			return strings.Join(names, ",")
		}
	}
	return model.DefaultChannelModels[channelType]
}

func maskKey(key string) string {
	return helper.MaskAPIKey(key)
}

func checkMissingInfo(cfg *BrandConfig) []string {
	var missing []string
	if cfg.ChannelType == channeltype.TogetherAI && cfg.ProviderName == "Meta" {
		missing = append(missing, "model_mapping",
			"Llama models are open-source and available through multiple providers. You may need to update model IDs manually in the channel settings.")
	}
	return missing
}

// SyncAllBrandModels 遍历所有已配置品牌的渠道，触发自动模型发现
// POST /api/admin/model-brands/sync-all
func SyncAllBrandModels(c *gin.Context) {
	var channels []model.Channel
	model.DB.Find(&channels)

	if len(channels) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "no channels to sync", "total": 0})
		return
	}

	// 构建 brand config 查找表
	knownConfigs := make(map[int]*BrandConfig)
	for i := range brandConfigs {
		knownConfigs[brandConfigs[i].ChannelType] = &brandConfigs[i]
	}

	type syncResult struct {
		Name    string
		Success bool
		Error   string
	}

	results := []syncResult{}
	for i := range channels {
		ch := &channels[i]
		// 只对有真实 API Key 的渠道触发自动发现
		if ch.Key == "" || strings.HasPrefix(ch.Key, "PUT_YOUR") {
			continue
		}
		// 检查是否需要跳过（量子等非 OpenAI 兼容品牌）
		if cfg, ok := knownConfigs[ch.Type]; ok {
			// 检查该品牌是否支持自动发现
			supportsDiscover := false
			for _, f := range cfg.AutoFields {
				if f == "models" {
					supportsDiscover = true
					break
				}
			}
			if !supportsDiscover {
				continue
			}
		}

		// 执行自动发现
		baseURL := ""
		if ch.BaseURL != nil {
			baseURL = *ch.BaseURL
		}

		// 在 goroutine 中异步执行
		result := syncResult{Name: ch.Name, Success: true}
		if err := autoDiscoverModels(ch.Id, ch.Type, baseURL, ch.Key, ch.Name); err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			logger.SysLog("SyncAllBrandModels[" + ch.Name + "] completed")
		}
		results = append(results, result)
	}

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      fmt.Sprintf("synced %d/%d channels", successCount, len(results)),
		"total":        len(results),
		"synced":       successCount,
		"results":      results,
	})
}


// AutoConfigureAllFromEnv reads all *_API_KEY environment variables
// and auto-configures corresponding channel keys at startup.
// This is the single source of truth for mapping .env API keys to platform channels.
// It either updates placeholder keys on existing seed channels, or CREATES
// channels for completely new providers not in the seed data.
func AutoConfigureAllFromEnv() []string {
	var results []string

	// Map environment variable names to providers.
	// When a new provider is added to .env, just add its env var here.
	type envEntry struct {
		EnvVar       string
		ProviderName string
	}
	envMap := []envEntry{
		{"DEEPSEEK_API_KEY", "DeepSeek"},
		{"GROQ_API_KEY", "Groq"},
		{"OPENAI_API_KEY", "OpenAI"},
		{"ANTHROPIC_API_KEY", "Anthropic"},
		{"GEMINI_API_KEY", "Google"},
		{"SILICONFLOW_API_KEY", "SiliconFlow"},
		{"MISTRAL_API_KEY", "Mistral"},
		{"MOONSHOT_API_KEY", "Moonshot"},
		{"BAIDU_API_KEY", "Baidu"},
		{"ZHIPU_API_KEY", "Zhipu"},
		{"TENCENT_API_KEY", "Tencent"},
		{"ALIBABA_API_KEY", "Alibaba"},
	}

	now := time.Now().Unix()
	// Read .env file directly as fallback if os.Getenv returns empty
	// This handles cases where godotenv/autoload may not load correctly
	envVars := make(map[string]string)
	if envData, err := os.ReadFile(".env"); err == nil {
		for _, line := range strings.Split(string(envData), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
				envVars[parts[0]] = parts[1]
			}
		}
	}

	for _, entry := range envMap {
		apiKey := os.Getenv(entry.EnvVar)
		if apiKey == "" {
			if key, ok := envVars[entry.EnvVar]; ok && key != "" {
				apiKey = key
			}
		}
		if apiKey == "" {
			continue
		}

		// Decrypt if .env key uses QC! encrypted format (AES-GCM + Qsc suffix)
		if config.CryptoSecret != "" && strings.HasPrefix(apiKey, encrypt.EnvKeyPrefix) {
			decrypted, err := encrypt.DecryptEnvKey(apiKey, config.CryptoSecret)
			if err != nil {
				results = append(results, entry.ProviderName+": env key decrypt failed: "+err.Error())
				continue
			}
			apiKey = decrypted
			logger.SysLog("AutoConfigure[" + entry.ProviderName + "]: decrypted env key (QC! format)")
		}

		matchedCfg := findBrandConfig(entry.ProviderName)
		if matchedCfg == nil {
			results = append(results, entry.ProviderName+": unknown provider (add to brandConfigs)")
			continue
		}

		var channel model.Channel
		result := model.DB.Where("type = ?", matchedCfg.ChannelType).First(&channel)

		if result.Error != nil {
			// Channel does not exist -> CREATE it automatically
			// This handles providers not in the seed data or after DB reset
			baseURL := matchedCfg.DefaultURL
			if baseURL == "" && matchedCfg.ChannelType >= 0 && matchedCfg.ChannelType < len(channeltype.ChannelBaseURLs) {
				baseURL = channeltype.ChannelBaseURLs[matchedCfg.ChannelType]
			}
			models := getChannelModelsFromBrand(matchedCfg.ProviderName, matchedCfg.ChannelType)

			newChannel := &model.Channel{
				Type:        matchedCfg.ChannelType,
				Name:        matchedCfg.ChannelName,
				Key:         apiKey,
				BaseURL:     model.StrPtr(baseURL),
				Models:      models,
				Group:       "default",
				Status:      model.ChannelStatusEnabled,
				Weight:      model.UintPtr(1),
				CreatedTime: now,
			}
			if err := model.DB.Create(newChannel).Error; err != nil {
				results = append(results, entry.ProviderName+": create failed: "+err.Error())
				continue
			}
			channel = *newChannel
			// Populate abilities so Distribute can route
			if err := newChannel.AddAbilities(); err != nil {
				logger.SysError("AutoConfigure[" + entry.ProviderName + "] AddAbilities: " + err.Error())
			}

			// Run auto-discover
			go func(chID int, chType int, u string, k string, n string) {
				if err := autoDiscoverModels(chID, chType, u, k, n); err != nil {
					logger.SysError("autoDiscover[" + n + "]: " + err.Error())
				}
			}(channel.Id, channel.Type, baseURL, apiKey, channel.Name)

			results = append(results, entry.ProviderName+": CREATED, auto-discover started")
		} else if strings.Contains(channel.Key, "PUT_YOUR_API_KEY_HERE") || channel.Key == "" {
			// Channel exists but key is placeholder -> update with real key
			if err := model.DB.Model(&channel).Update("key", apiKey).Error; err != nil {
				results = append(results, entry.ProviderName+": update key failed: "+err.Error())
				continue
			}
			// Ensure key is encrypted in DB
			if config.CryptoSecret != "" {
				var fresh model.Channel
				if err := model.DB.First(&fresh, channel.Id).Error; err == nil && fresh.Key != "" {
					_, decErr := encrypt.DecryptChannelKey(fresh.Key, config.CryptoSecret)
					if decErr != nil {
						// Plaintext key: encrypt it
						encrypted, encErr := encrypt.EncryptChannelKey(fresh.Key, config.CryptoSecret)
						if encErr == nil {
							model.DB.Model(&fresh).Update("key", encrypted)
						}
					}
				}
			}
			// Also ensure abilities exist (in case they were lost)
			if err := channel.AddAbilities(); err != nil {
				logger.SysError("AutoConfigure[" + entry.ProviderName + "] AddAbilities: " + err.Error())
			}
			// Run auto-discover
			baseURL := ""
			if channel.BaseURL != nil {
				baseURL = *channel.BaseURL
			}
			go func(chID int, chType int, u string, k string, n string) {
				if err := autoDiscoverModels(chID, chType, u, k, n); err != nil {
					logger.SysError("autoDiscover[" + n + "]: " + err.Error())
				}
			}(channel.Id, channel.Type, baseURL, apiKey, channel.Name)
			results = append(results, entry.ProviderName+": KEY UPDATED, auto-discover started")
		} else {
			// Channel has a real key -> always update from .env (source of truth)
			// This ensures key changes in .env propagate to the channel on every restart
			logger.SysLog("AutoConfigure[" + entry.ProviderName + "] else: key updated")
			if err := model.DB.Model(&channel).Update("key", apiKey).Error; err != nil {
				results = append(results, entry.ProviderName+": update key failed: "+err.Error())
				continue
			}
			// Encrypt the new key in DB
			if config.CryptoSecret != "" {
				var fresh model.Channel
				if err := model.DB.First(&fresh, channel.Id).Error; err == nil && fresh.Key != "" {
					logger.SysLog("AutoConfigure[" + entry.ProviderName + "] encrypt check: fresh.Key[:20]=" + fresh.Key[:20] + ", len=" + fmt.Sprint(len(fresh.Key)))
					_, decErr := encrypt.DecryptChannelKey(fresh.Key, config.CryptoSecret)
					if decErr != nil {
						logger.SysLog("AutoConfigure[" + entry.ProviderName + "] encrypt: key is plaintext, encrypting...")
						encrypted, encErr := encrypt.EncryptChannelKey(fresh.Key, config.CryptoSecret)
						if encErr == nil {
							logger.SysLog("AutoConfigure[" + entry.ProviderName + "] encrypt: encrypted key[:20]=" + encrypted[:20] + ", last4=" + encrypted[len(encrypted)-4:])
							model.DB.Model(&fresh).Update("key", encrypted)
						}
					} else {
						logger.SysLog("AutoConfigure[" + entry.ProviderName + "] encrypt: key already encrypted, skipping")
					}
				}
			}
			// Re-populate abilities to reflect any model changes
			if err := channel.AddAbilities(); err != nil {
				logger.SysError("AutoConfigure[" + entry.ProviderName + "] AddAbilities: " + err.Error())
			}
			// Run auto-discover with new key
			baseURL := ""
			if channel.BaseURL != nil {
				baseURL = *channel.BaseURL
			}
			go func(chID int, chType int, u string, k string, n string) {
				if err := autoDiscoverModels(chID, chType, u, k, n); err != nil {
					logger.SysError("autoDiscover[" + n + "]: " + err.Error())
				}
			}(channel.Id, channel.Type, baseURL, apiKey, channel.Name)
			results = append(results, entry.ProviderName+": KEY UPDATED from .env, auto-discover started")
		}
	}
	return results
}
func findBrandConfig(providerName string) *BrandConfig {
	for i := range brandConfigs {
		if strings.EqualFold(brandConfigs[i].ProviderName, providerName) {
			return &brandConfigs[i]
		}
	}
	cfg := generateBrandConfig(providerName)
	if cfg.ProviderName != "" {
		return &cfg
	}
	return nil
}

// ConfigureAllBrands is the HTTP handler for POST /api/admin/model-brands/configure-all
func ConfigureAllBrands(c *gin.Context) {
	results := AutoConfigureAllFromEnv()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "auto-configure completed",
		"results": results,
	})
}
