package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ModelMetadata stores model descriptions and capabilities per language.
// model_name + languages_type is unique: each model has one row per language.
type ModelMetadata struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	ModelName       string `gorm:"type:varchar(255);not null;uniqueIndex:idx_model_lang" json:"model_name"`
	LanguagesType   string `gorm:"type:varchar(100);not null;uniqueIndex:idx_model_lang" json:"languages_type"`
	DisplayName     string `gorm:"type:varchar(255)" json:"display_name"`
	Description     string `gorm:"type:text" json:"description"`
	UseCase         string `gorm:"type:varchar(50)" json:"use_case"`
	ContextWindow   int    `json:"context_window"`
	InputModalities string `gorm:"type:text" json:"input_modalities"`
	Series          string `gorm:"type:varchar(100)" json:"series"`
	Provider        string `gorm:"type:varchar(100)" json:"provider"`
	// ── 详情字段（可为空）──
	KnowledgeCutoff  string `gorm:"type:varchar(50)" json:"knowledge_cutoff"`
	BenchmarkScores  string `gorm:"type:text" json:"benchmark_scores"`
	Capabilities     string `gorm:"type:text" json:"capabilities"`
	RecommendedFor   string `gorm:"type:varchar(500)" json:"recommended_for"`
	OpenSource       bool   `gorm:"default:false" json:"open_source"`
	License          string `gorm:"type:varchar(100)" json:"license"`
	Strengths        string `gorm:"type:varchar(500)" json:"strengths"`

	CreatedTime     int64  `json:"created_time"`
	UpdatedTime     int64  `json:"updated_time"`
}

// Modalities returns the parsed list of input modalities.
func (m *ModelMetadata) Modalities() []string {
	var mods []string
	if m.InputModalities != "" {
		_ = json.Unmarshal([]byte(m.InputModalities), &mods)
	}
	return mods
}

// ParsedBenchmarks returns the parsed benchmark scores map.
func (m *ModelMetadata) ParsedBenchmarks() map[string]float64 {
	result := make(map[string]float64)
	if m.BenchmarkScores != "" {
		_ = json.Unmarshal([]byte(m.BenchmarkScores), &result)
	}
	return result
}

// ParsedCapabilities returns the parsed capabilities list.
func (m *ModelMetadata) ParsedCapabilities() []string {
	var caps []string
	if m.Capabilities != "" {
		_ = json.Unmarshal([]byte(m.Capabilities), &caps)
	}
	return caps
}

// ── Response DTO for GET /api/model-catalog ──

// CatalogModelResponse is what the frontend uses on the Models page.
type CatalogModelResponse struct {
	Name            string   `json:"name"`
	DisplayName     string   `json:"display_name"`
	Description     string   `json:"description"`
	UseCase         string   `json:"use_case"`
	ContextWindow   int      `json:"context_window"`
	InputModalities []string `json:"input_modalities"`
	Series          string   `json:"series"`
	Provider        string   `json:"provider"`

	// ── 新增详情字段 ──
	KnowledgeCutoff string             `json:"knowledge_cutoff"`
	BenchmarkScores map[string]float64 `json:"benchmark_scores"`
	Capabilities    []string           `json:"capabilities"`
	RecommendedFor  string             `json:"recommended_for"`
	OpenSource      bool               `json:"open_source"`
	License         string             `json:"license"`
	Strengths       string             `json:"strengths"`

	// Channel-backed fields (nullable when no channel is configured)
	ChannelID   int     `json:"channel_id"`
	ChannelName string  `json:"channel_name"`
	InputPrice  float64 `json:"input_price"`
	OutputPrice float64 `json:"output_price"`
	Status      int     `json:"status"` // 1 = configured, 0 = unconfigured
	Group       string  `json:"group"`
}

// backfillModelDetails fills in knowledge_cutoff, benchmark_scores, capabilities etc.
// for models that already exist in the DB but have empty new detail fields.
func backfillModelDetails() {
	now := time.Now().Unix()

	type modelDetail struct {
		name      string
		cutoff    string
		benchmarks map[string]float64
		caps      []string
		recommend string
		openSrc   bool
		license   string
		strengths string
	}

	details := []modelDetail{
		{"GPT-4o", "2025-04", map[string]float64{"mmlu": 0.887, "human_eval": 0.901, "gsm8k": 0.953},
			[]string{"Chat", "Vision", "Code Generation", "Reasoning", "Function Calling"},
			"gpt-4o_recommend", false, "",
			"gpt-4o_strengths"},
		{"GPT-4o-mini", "2025-04", map[string]float64{"mmlu": 0.821, "human_eval": 0.872, "gsm8k": 0.877},
			[]string{"Chat", "Vision", "Code Generation", "Function Calling"},
			"gpt-4o-mini_recommend", false, "",
			"gpt-4o-mini_strengths"},
		{"Claude 3.5 Sonnet", "2025-01", map[string]float64{"mmlu": 0.884, "human_eval": 0.896, "gsm8k": 0.942},
			[]string{"Chat", "Vision", "Code Generation", "Reasoning", "File Analysis"},
			"claude-3.5-sonnet_recommend", false, "",
			"claude-3.5-sonnet_strengths"},
		{"Claude 3.5 Haiku", "2025-01", map[string]float64{"mmlu": 0.829, "gsm8k": 0.886},
			[]string{"Chat", "Vision", "Code Generation", "File Analysis"},
			"claude-3.5-haiku_recommend", false, "",
			"claude-3.5-haiku_strengths"},
		{"Claude Opus 4", "2025-04", map[string]float64{"mmlu": 0.899, "human_eval": 0.927, "gsm8k": 0.958},
			[]string{"Chat", "Vision", "Code Generation", "Reasoning", "File Analysis", "Tool Use"},
			"claude-opus-4_recommend", false, "",
			"claude-opus-4_strengths"},
		{"Gemini 2.0 Flash", "2025-05", map[string]float64{"mmlu": 0.886, "human_eval": 0.848, "gsm8k": 0.941},
			[]string{"Chat", "Vision", "Code Generation", "Audio/Video Understanding"},
			"gemini-2.0-flash_recommend", false, "",
			"gemini-2.0-flash_strengths"},
		{"Gemini 2.0 Pro", "2025-05", map[string]float64{"mmlu": 0.907, "human_eval": 0.892, "gsm8k": 0.954},
			[]string{"Chat", "Vision", "Code Generation", "Reasoning", "Audio/Video Understanding"},
			"gemini-2.0-pro_recommend", false, "",
			"gemini-2.0-pro_strengths"},
		{"DeepSeek Chat", "2025-03", map[string]float64{"mmlu": 0.843, "human_eval": 0.819, "gsm8k": 0.887},
			[]string{"Chat", "Code Generation", "Reasoning"},
			"deepseek-chat_recommend", true, "MIT",
			"deepseek-chat_strengths"},
		{"DeepSeek R1", "2025-03", map[string]float64{"mmlu": 0.901, "human_eval": 0.967, "gsm8k": 0.959, "math": 0.976},
			[]string{"Reasoning", "Code Generation", "Math", "Problem Solving"},
			"deepseek-r1_recommend", true, "MIT",
			"deepseek-r1_strengths"},
		{"DeepSeek V3", "2025-03", map[string]float64{"mmlu": 0.878, "human_eval": 0.893, "gsm8k": 0.925},
			[]string{"Chat", "Code Generation", "Reasoning"},
			"deepseek-v3_recommend", true, "MIT",
			"deepseek-v3_strengths"},
		{"Qwen Max", "2025-03", map[string]float64{"mmlu": 0.861, "human_eval": 0.783, "gsm8k": 0.878},
			[]string{"Chat", "Code Generation", "Multilingual"},
			"qwen-max_recommend", false, "",
			"qwen-max_strengths"},
		{"Qwen Plus", "2025-03", map[string]float64{"mmlu": 0.845, "gsm8k": 0.872},
			[]string{"Chat", "Code Generation", "Multilingual"},
			"qwen-plus_recommend", false, "",
			"qwen-plus_strengths"},
		{"Qwen 2.5 VL", "2025-03", map[string]float64{"mmlu": 0.847, "gsm8k": 0.892},
			[]string{"Vision", "Chat", "OCRs"},
			"qwen-2.5-vl_recommend", true, "Apache 2.0",
			"qwen-2.5-vl_strengths"},
		{"Mistral Large", "2025-02", map[string]float64{"mmlu": 0.853, "human_eval": 0.835, "gsm8k": 0.872},
			[]string{"Chat", "Code Generation", "Multilingual", "Reasoning"},
			"mistral-large_recommend", false, "",
			"mistral-large_strengths"},
		{"Mixtral 8x7B", "2024-10", map[string]float64{"mmlu": 0.826, "human_eval": 0.751, "gsm8k": 0.851},
			[]string{"Chat", "Multilingual"},
			"mixtral-8x7b_recommend", true, "Apache 2.0",
			"mixtral-8x7b_strengths"},
		{"Llama 3.1 70B", "2024-12", map[string]float64{"mmlu": 0.858, "human_eval": 0.801, "gsm8k": 0.901},
			[]string{"Chat", "Code Generation", "Reasoning", "Tool Use"},
			"llama-3.1-70b_recommend", true, "Llama 3.1 License",
			"llama-3.1-70b_strengths"},
		{"Llama 3.1 405B", "2024-12", map[string]float64{"mmlu": 0.874, "human_eval": 0.841, "gsm8k": 0.921},
			[]string{"Chat", "Code Generation", "Reasoning", "Tool Use"},
			"llama-3.1-405b_recommend", true, "Llama 3.1 License",
			"llama-3.1-405b_strengths"},
		{"Llama 3.2 Vision", "2024-12", map[string]float64{"mmlu": 0.832, "gsm8k": 0.879},
			[]string{"Vision", "Chat", "Code Generation"},
			"llama-3.2-vision_recommend", true, "Llama 3.2 License",
			"llama-3.2-vision_strengths"},
		{"o1", "2025-04", map[string]float64{"mmlu": 0.921, "human_eval": 0.924, "gsm8k": 0.968, "math": 0.942},
			[]string{"Reasoning", "Code Generation", "Math", "Problem Solving", "Vision"},
			"o1_recommend", false, "",
			"o1_strengths"},
		{"o3", "2025-04", map[string]float64{"mmlu": 0.933, "human_eval": 0.971, "gsm8k": 0.989, "math": 0.967, "arc": 0.957},
			[]string{"Reasoning", "Code Generation", "Math", "Problem Solving", "Vision", "Scientific Research"},
			"o3_recommend", false, "",
			"o3_strengths"},
		{"Codestral", "2025-01", map[string]float64{"human_eval": 0.882, "mbpp": 0.856},
			[]string{"Code Generation", "Code Completion", "Code Review"},
			"codestral_recommend", false, "",
			"codestral_strengths"},
		{"GPT-4 Turbo", "2024-12", map[string]float64{"mmlu": 0.864, "human_eval": 0.832, "gsm8k": 0.915},
			[]string{"Chat", "Vision", "Code Generation"},
			"gpt-4-turbo_recommend", false, "",
			"gpt-4-turbo_strengths"},
		{"GPT-3.5 Turbo", "2024-09", map[string]float64{"mmlu": 0.813, "gsm8k": 0.832},
			[]string{"Chat"},
			"gpt-3.5-turbo_recommend", false, "",
			"gpt-3.5-turbo_strengths"},
		{"Gemini 3.0 Pro", "2025-05", map[string]float64{"mmlu": 0.915, "human_eval": 0.903, "gsm8k": 0.962},
			[]string{"Chat", "Vision", "Code Generation", "Reasoning", "Audio/Video Understanding"},
			"gemini-3.0-pro_recommend", false, "",
			"gemini-3.0-pro_strengths"},
	}

	updated := 0
	for _, d := range details {
		benchJSON, benchErr := json.Marshal(d.benchmarks)
		if benchErr != nil {
			benchJSON = []byte("{}")
		}
		capsJSON, capsErr := json.Marshal(d.caps)
		if capsErr != nil {
			capsJSON = []byte("[]")
		}

		result := DB.Model(&ModelMetadata{}).
			Where("model_name = ? AND (knowledge_cutoff = '' OR knowledge_cutoff IS NULL)", d.name).
			Updates(map[string]interface{}{
				"knowledge_cutoff": d.cutoff,
				"benchmark_scores": string(benchJSON),
				"capabilities":     string(capsJSON),
				"recommended_for":  d.recommend,
				"open_source":      d.openSrc,
				"license":          d.license,
				"strengths":        d.strengths,
				"updated_time":     now,
			})
		if result.Error == nil {
			updated += int(result.RowsAffected)
		}
	}

	if updated > 0 {
		logger.SysLog(fmt.Sprintf("backfillModelDetails: %d rows updated with model details", updated))
	}

	// Migration: convert existing Chinese recommend/strengths to i18n keys
	// This is a one-time operation for rows seeded before the i18n key migration.
	migrated := 0
	for _, d := range details {
		recKey := d.name + "_recommend"
		strKey := d.name + "_strengths"
		// Only update rows that still have Chinese text (old seed format)
		r := DB.Model(&ModelMetadata{}).
			Where("model_name = ?", d.name).
			Where("(recommended_for != ? AND recommended_for != '')", recKey).
			Updates(map[string]interface{}{
				"recommended_for": recKey,
				"strengths":       strKey,
			})
		if r.Error == nil {
			migrated += int(r.RowsAffected)
		}
	}
	if migrated > 0 {
		logger.SysLog(fmt.Sprintf("backfillModelDetails: %d rows migrated from Chinese to i18n keys", migrated))
	}
}

// SeedModelMetadata ensures the table exists and seeds initial data.
func SeedModelMetadata() {
	// Ensure table exists first
	if err := DB.AutoMigrate(&ModelMetadata{}); err != nil {
		logger.SysError("SeedModelMetadata AutoMigrate failed: " + err.Error())
		return
	}

	// Backfill details for existing models (new fields added by migration)
	backfillModelDetails()

	var count int64
	DB.Model(&ModelMetadata{}).Count(&count)
	if count > 0 {
		return // already seeded
	}

	now := time.Now().Unix()

	// Base definitions: model_name → capabilities (language-independent)
	type baseDef struct {
		DisplayName     string
		Description     map[string]string // languages_type → description
		UseCase         string
		ContextWindow   int
		InputModalities string
		Series          string
		Provider        string
	}

	languages := []string{"中文简体", "中文繁体", "English", "Français", "日本語", "Русский", "Tiếng Việt"}

	defs := []baseDef{
		{
			DisplayName: "GPT-4o", UseCase: "chat", ContextWindow: 128000,
			InputModalities: `["Text","Image"]`, Series: "GPT-4", Provider: "OpenAI",
			Description: map[string]string{
				"中文简体":  "OpenAI 旗舰多模态模型，支持视觉理解、高智能和快速响应。",
				"中文繁体":  "OpenAI 旗艦多模態模型，支援視覺理解、高智能和快速回應。",
				"English":  "OpenAI flagship multimodal model with vision, high intelligence and fast responses.",
				"Français": "Modèle multimodal phare d'OpenAI avec vision, haute intelligence et réponses rapides.",
				"日本語":  "OpenAIの旗艦マルチモーダルモデル。ビジョン対応、高インテリジェンス、高速応答。",
				"Русский": "Флагманская мультимодальная модель OpenAI с поддержкой изображений.",
				"Tiếng Việt": "Mô hình đa phương thức hàng đầu của OpenAI với khả năng nhìn, trí thông minh cao.",
			},
		},
		{
			DisplayName: "GPT-4o-mini", UseCase: "chat", ContextWindow: 128000,
			InputModalities: `["Text","Image"]`, Series: "GPT-4", Provider: "OpenAI",
			Description: map[string]string{
				"中文简体": "快速、经济的小模型，适合轻量级任务，表现强劲。",
				"English": "Fast, affordable small model for lightweight tasks with strong performance.",
				"Français": "Petit modèle rapide et abordable pour les tâches légères.",
				"日本語": "高速で手頃な小型モデル。軽量タスクに最適。",
				"Русский": "Быстрая и доступная малая модель для лёгких задач.",
			},
		},
		{
			DisplayName: "Claude 3.5 Sonnet", UseCase: "chat", ContextWindow: 200000,
			InputModalities: `["Text","Image","File"]`, Series: "Claude 3.5", Provider: "Anthropic",
			Description: map[string]string{
				"中文简体": "Anthropic 的速度与智能平衡最佳模型，支持文件和图像。",
				"English":  "Best balance of speed and intelligence from Anthropic.",
				"Français": "Meilleur équilibre entre vitesse et intelligence chez Anthropic.",
				"日本語":  "Anthropicの速度と知能の最適バランスモデル。",
			},
		},
		{
			DisplayName: "Claude 3.5 Haiku", UseCase: "chat", ContextWindow: 200000,
			InputModalities: `["Text","Image","File"]`, Series: "Claude 3.5", Provider: "Anthropic",
			Description: map[string]string{
				"中文简体": "快速经济的 Claude 模型，适合日常任务。",
				"English":  "Fast and affordable Claude model for everyday tasks.",
				"Français": "Modèle Claude rapide et abordable pour les tâches quotidiennes.",
				"日本語":  "日常タスクに最適な高速で手頃なClaudeモデル。",
			},
		},
		{
			DisplayName: "Claude Opus 4", UseCase: "chat", ContextWindow: 200000,
			InputModalities: `["Text","Image","File"]`, Series: "Claude", Provider: "Anthropic",
			Description: map[string]string{
				"中文简体": "Anthropic 最强大的模型，适合复杂的企业级工作负载。",
				"English":  "Anthropic most powerful model for complex enterprise workloads.",
				"Français": "Le modèle le plus puissant d'Anthropic pour les charges de travail complexes.",
				"日本語":  "複雑なエンタープライズワークロード向けのAnthropic最強モデル。",
			},
		},
		{
			DisplayName: "Gemini 2.0 Flash", UseCase: "chat", ContextWindow: 1000000,
			InputModalities: `["Text","Image","Audio","Video","File"]`, Series: "Gemini 2.0", Provider: "Google",
			Description: map[string]string{
				"中文简体": "Google 快速多模态模型，1M 上下文，支持音频和视频输入。",
				"English":  "Google fast multimodal model with 1M context, supports audio/video.",
				"Français": "Modèle multimodal rapide de Google avec contexte 1M, support audio/vidéo.",
				"日本語":  "Googleの高速マルチモーダルモデル。1Mコンテキスト、音声/動画対応。",
			},
		},
		{
			DisplayName: "Gemini 2.0 Pro", UseCase: "chat", ContextWindow: 1000000,
			InputModalities: `["Text","Image","Audio","Video","File"]`, Series: "Gemini 2.0", Provider: "Google",
			Description: map[string]string{
				"中文简体": "Google 高质量模型，完整的多模态输入支持。",
				"English":  "Google high-quality model with full multimodal input support.",
				"Français": "Modèle haute qualité de Google avec support multimodal complet.",
				"日本語":  "完全なマルチモーダル入力をサポートするGoogle高品質モデル。",
			},
		},
		{
			DisplayName: "DeepSeek Chat", UseCase: "chat", ContextWindow: 128000,
			InputModalities: `["Text"]`, Series: "DeepSeek", Provider: "DeepSeek",
			Description: map[string]string{
				"中文简体": "DeepSeek 通用对话模型，性价比极高。",
				"English":  "DeepSeek general chat model offering excellent value.",
				"Français": "Modèle de discussion général DeepSeek offrant un excellent rapport qualité-prix.",
				"日本語":  "優れたコストパフォーマンスを提供するDeepSeek汎用チャットモデル。",
			},
		},
		{
			DisplayName: "DeepSeek R1", UseCase: "reasoning", ContextWindow: 128000,
			InputModalities: `["Text"]`, Series: "DeepSeek", Provider: "DeepSeek",
			Description: map[string]string{
				"中文简体": "DeepSeek 推理模型，具有透明的逐步思考链。",
				"English":  "DeepSeek reasoning model with transparent step-by-step thinking.",
				"Français": "Modèle de raisonnement DeepSeek avec chaîne de pensée transparente.",
				"日本語":  "透明なステップバイステップ思考を備えたDeepSeek推論モデル。",
			},
		},
		{
			DisplayName: "Qwen Max", UseCase: "chat", ContextWindow: 32000,
			InputModalities: `["Text"]`, Series: "Qwen", Provider: "Alibaba",
			Description: map[string]string{
				"中文简体": "阿里云最强通用语言模型。",
				"English":  "Alibaba strongest general-purpose language model.",
				"Français": "Le modèle de langage général le plus puissant d'Alibaba.",
				"日本語":  "Alibaba最強の汎用言語モデル。",
			},
		},
		{
			DisplayName: "Mistral Large", UseCase: "chat", ContextWindow: 128000,
			InputModalities: `["Text"]`, Series: "Mistral", Provider: "Mistral",
			Description: map[string]string{
				"中文简体": "Mistral AI 旗舰模型，适合复杂推理任务。",
				"English":  "Mistral AI flagship model for complex reasoning tasks.",
				"Français": "Modèle phare de Mistral AI pour les tâches de raisonnement complexes.",
				"日本語":  "複雑な推論タスク向けのMistral AI旗艦モデル。",
			},
		},
		{
			DisplayName: "Llama 3.1 70B", UseCase: "chat", ContextWindow: 128000,
			InputModalities: `["Text"]`, Series: "Llama 3.1", Provider: "Meta",
			Description: map[string]string{
				"中文简体": "Meta 开源模型，综合表现强劲。",
				"English":  "Meta open-source model with strong general performance.",
				"Français": "Modèle open-source de Meta avec de solides performances générales.",
				"日本語":  "強力な汎用パフォーマンスを備えたMetaのオープンソースモデル。",
			},
		},
		{
			DisplayName: "o1", UseCase: "reasoning", ContextWindow: 200000,
			InputModalities: `["Text","Image"]`, Series: "o1", Provider: "OpenAI",
			Description: map[string]string{
				"中文简体": "OpenAI 推理模型，专为复杂问题解决和分析设计。",
				"English":  "OpenAI reasoning model designed for complex problem solving and analysis.",
				"Français": "Modèle de raisonnement OpenAI conçu pour la résolution de problèmes complexes.",
				"日本語":  "複雑な問題解決と分析のために設計されたOpenAI推論モデル。",
			},
		},
		{
			DisplayName: "o3", UseCase: "reasoning", ContextWindow: 200000,
			InputModalities: `["Text","Image"]`, Series: "o3", Provider: "OpenAI",
			Description: map[string]string{
				"中文简体": "高级推理模型，具备扩展思考能力。",
				"English":  "Advanced reasoning model with extended thinking capabilities.",
				"Français": "Modèle de raisonnement avancé avec capacités de réflexion étendues.",
				"日本語":  "拡張思考機能を備えた高度な推論モデル。",
			},
		},
		{
			DisplayName: "Codestral", UseCase: "coding", ContextWindow: 32000,
			InputModalities: `["Text"]`, Series: "Mistral", Provider: "Mistral",
			Description: map[string]string{
				"中文简体": "Mistral 专用代码生成模型。",
				"English":  "Mistral dedicated code generation model for developers.",
				"Français": "Modèle de génération de code dédié de Mistral pour les développeurs.",
				"日本語":  "開発者向けMistral専用コード生成モデル。",
			},
		},
		{
			DisplayName: "Llama 3.2 Vision", UseCase: "vision", ContextWindow: 128000,
			InputModalities: `["Text","Image"]`, Series: "Llama 3.2", Provider: "Meta",
			Description: map[string]string{
				"中文简体": "Meta 多模态开源模型，支持视觉语言任务。",
				"English":  "Meta multimodal open-source model for vision-language tasks.",
				"Français": "Modèle open-source multimodal de Meta pour les tâches vision-langage.",
				"日本語":  "視覚言語タスク向けMetaマルチモーダルオープンソースモデル。",
			},
		},
		{
			DisplayName: "Qwen 2.5 VL", UseCase: "vision", ContextWindow: 128000,
			InputModalities: `["Text","Image"]`, Series: "Qwen 2.5", Provider: "Alibaba",
			Description: map[string]string{
				"中文简体": "阿里云多模态视觉语言模型，支持图像理解。",
				"English":  "Alibaba multimodal vision-language model for image understanding.",
				"Français": "Modèle vision-langage multimodal d'Alibaba pour la compréhension d'images.",
				"日本語":  "画像理解のためのAlibabaマルチモーダル視覚言語モデル。",
			},
		},
		{
			DisplayName: "Mixtral 8x7B", UseCase: "chat", ContextWindow: 32000,
			InputModalities: `["Text"]`, Series: "Mistral", Provider: "Mistral",
			Description: map[string]string{
				"中文简体": "Mistral MoE 架构，性价比优秀。",
				"English":  "Mistral MoE architecture offering good quality-to-cost ratio.",
				"Français": "Architecture MoE de Mistral offrant un bon rapport qualité-prix.",
				"日本語":  "優れた品質対コスト比を提供するMistral MoEアーキテクチャ。",
			},
		},
		{
			DisplayName: "DeepSeek V3", UseCase: "coding", ContextWindow: 64000,
			InputModalities: `["Text"]`, Series: "DeepSeek", Provider: "DeepSeek",
			Description: map[string]string{
				"中文简体": "DeepSeek 最新模型，针对代码生成优化。",
				"English":  "DeepSeek latest model optimized for code generation tasks.",
				"Français": "Dernier modèle DeepSeek optimisé pour la génération de code.",
				"日本語":  "コード生成タスクに最適化されたDeepSeek最新モデル。",
			},
		},
		{
			DisplayName: "GPT-4 Turbo", UseCase: "chat", ContextWindow: 128000,
			InputModalities: `["Text","Image"]`, Series: "GPT-4", Provider: "OpenAI",
			Description: map[string]string{
				"中文简体": "支持视觉的 GPT-4，扩展上下文窗口。",
				"English":  "GPT-4 with vision support and extended context window.",
				"Français": "GPT-4 avec support visuel et fenêtre de contexte étendue.",
				"日本語":  "ビジョン対応、拡張コンテキストウィンドウを備えたGPT-4。",
			},
		},
		{
			DisplayName: "GPT-3.5 Turbo", UseCase: "chat", ContextWindow: 16385,
			InputModalities: `["Text"]`, Series: "GPT-3.5", Provider: "OpenAI",
			Description: map[string]string{
				"中文简体": "快速经济的对话任务模型。",
				"English":  "Fast and cost-effective for simple conversational tasks.",
				"Français": "Rapide et économique pour les tâches de conversation simples.",
				"日本語":  "シンプルな会話タスクに高速でコスト効率の良いモデル。",
			},
		},
		{
			DisplayName: "Gemini 3.0 Pro", UseCase: "reasoning", ContextWindow: 1000000,
			InputModalities: `["Text","Image","Audio","Video","File"]`, Series: "Gemini 3.0", Provider: "Google",
			Description: map[string]string{
				"中文简体": "Google 最新推理模型，增强分析能力。",
				"English":  "Google latest reasoning model with enhanced analytical capabilities.",
				"Français": "Dernier modèle de raisonnement Google avec capacités analytiques améliorées.",
				"日本語":  "強化された分析機能を備えたGoogle最新推論モデル。",
			},
		},
		{
			DisplayName: "Llama 3.1 405B", UseCase: "chat", ContextWindow: 128000,
			InputModalities: `["Text"]`, Series: "Llama 3.1", Provider: "Meta",
			Description: map[string]string{
				"中文简体": "Meta 最大开源模型，接近 GPT-4 水平。",
				"English":  "Meta largest open-source model approaching GPT-4 level quality.",
				"Français": "Plus grand modèle open-source de Meta approchant la qualité GPT-4.",
				"日本語":  "GPT-4レベルの品質に迫るMeta最大のオープンソースモデル。",
			},
		},
		{
			DisplayName: "Qwen Plus", UseCase: "chat", ContextWindow: 128000,
			InputModalities: `["Text"]`, Series: "Qwen", Provider: "Alibaba",
			Description: map[string]string{
				"中文简体": "阿里云升级版模型，增强上下文理解。",
				"English":  "Alibaba upgraded model with extended context understanding.",
				"Français": "Modèle amélioré d'Alibaba avec compréhension contextuelle étendue.",
				"日本語":  "拡張コンテキスト理解を備えたAlibabaアップグレードモデル。",
			},
		},
		// ── Quantum processor metadata ──
		{
			DisplayName: "IonQ Aria", UseCase: "quantum", ContextWindow: 0,
			InputModalities: `["Quantum Circuit"]`, Series: "IonQ", Provider: "IonQ",
			Description: map[string]string{
				"中文简体":  "IonQ Aria 是领先的离子阱量子处理器，具有 25 个算法量子比特和高保真门操作。",
				"English":  "IonQ Aria is a leading trapped-ion quantum processor with 25 algorithmic qubits and high-fidelity gates.",
				"日本語":  "IonQ Ariaは、25のアルゴリズム量子ビットと高忠実度ゲートを備えた最先端のイオントラップ量子プロセッサです。",
			},
		},
		{
			DisplayName: "IonQ Forte", UseCase: "quantum", ContextWindow: 0,
			InputModalities: `["Quantum Circuit"]`, Series: "IonQ", Provider: "IonQ",
			Description: map[string]string{
				"中文简体":  "IonQ Forte 提供 36 个算法量子比特，支持原生门集和快速电路执行。",
				"English":  "IonQ Forte offers 36 algorithmic qubits with native gate set and fast circuit execution.",
				"日本語":  "IonQ Forteは、36のアルゴリズム量子ビット、ネイティブゲートセット、高速回路実行を提供します。",
			},
		},
		{
			DisplayName: "IonQ Harmony", UseCase: "quantum", ContextWindow: 0,
			InputModalities: `["Quantum Circuit"]`, Series: "IonQ", Provider: "IonQ",
			Description: map[string]string{
				"中文简体":  "IonQ Harmony 拥有 11 个量子比特，稳定运行，适合教育和原型开发。",
				"English":  "IonQ Harmony has 11 qubits with stable operation, ideal for education and prototyping.",
				"日本語":  "IonQ Harmonyは11量子ビット、安定した動作で教育やプロトタイピングに最適です。",
			},
		},
		{
			DisplayName: "IBM Brisbane", UseCase: "quantum", ContextWindow: 0,
			InputModalities: `["Quantum Circuit"]`, Series: "IBM Quantum", Provider: "IBM",
			Description: map[string]string{
				"中文简体":  "IBM Brisbane 基于 127 量子比特 Eagle 处理器，支持动态电路。",
				"English":  "IBM Brisbane features 127 qubits (Eagle processor) with dynamic circuit support.",
				"日本語":  "IBM Brisbaneは127量子ビット（Eagleプロセッサ）を搭載し、動的回路をサポートします。",
			},
		},
		{
			DisplayName: "IBM Kyiv", UseCase: "quantum", ContextWindow: 0,
			InputModalities: `["Quantum Circuit"]`, Series: "IBM Quantum", Provider: "IBM",
			Description: map[string]string{
				"中文简体":  "IBM Kyiv 提供 127 个量子比特，接入 IBM Quantum Network。",
				"English":  "IBM Kyiv provides 127 qubits with IBM Quantum Network access.",
				"日本語":  "IBM Kyivは127量子ビットを提供し、IBM Quantum Networkにアクセスできます。",
			},
		},
		{
			DisplayName: "IBM Sherbrooke", UseCase: "quantum", ContextWindow: 0,
			InputModalities: `["Quantum Circuit"]`, Series: "IBM Quantum", Provider: "IBM",
			Description: map[string]string{
				"中文简体":  "IBM Sherbrooke 具备 127 量子比特，集成 Q-CTRL 错误抑制技术。",
				"English":  "IBM Sherbrooke features 127 qubits with Q-CTRL error suppression integration.",
				"日本語":  "IBM Sherbrookeは127量子ビットを搭載し、Q-CTRLエラー抑制技術を統合しています。",
			},
		},
		{
			DisplayName: "Rigetti Aspen", UseCase: "quantum", ContextWindow: 0,
			InputModalities: `["Quantum Circuit"]`, Series: "Rigetti", Provider: "Rigetti",
			Description: map[string]string{
				"中文简体":  "Rigetti Aspen 拥有 80+ 量子比特，可扩展架构，支持量子-经典混合计算。",
				"English":  "Rigetti Aspen has 80+ qubits with an extensible architecture and quantum-classical hybrid support.",
				"日本語":  "Rigetti Aspenは80+量子ビット、拡張可能なアーキテクチャ、量子-古典ハイブリッドをサポートします。",
			},
		},
		{
			DisplayName: "Rigetti Ankaa", UseCase: "quantum", ContextWindow: 0,
			InputModalities: `["Quantum Circuit"]`, Series: "Rigetti", Provider: "Rigetti",
			Description: map[string]string{
				"中文简体":  "Rigetti Ankaa 提供 84 个量子比特，改进的相干时间和门保真度。",
				"English":  "Rigetti Ankaa delivers 84 qubits with improved coherence times and gate fidelity.",
				"日本語":  "Rigetti Ankaaは84量子ビット、改善されたコヒーレンス時間とゲート忠実度を提供します。",
			},
		},

	}

	// Insert rows for each language
	for _, def := range defs {
		for _, lang := range languages {
			desc := def.Description[lang]
			if desc == "" {
				desc = def.Description["English"] // fallback to English
			}
			m := &ModelMetadata{
				ModelName:       def.DisplayName,
				LanguagesType:   lang,
				DisplayName:     def.DisplayName,
				Description:     desc,
				UseCase:         def.UseCase,
				ContextWindow:   def.ContextWindow,
				InputModalities: def.InputModalities,
				Series:          def.Series,
				Provider:        def.Provider,
				CreatedTime:     now,
				UpdatedTime:     now,
			}
			if err := DB.Create(m).Error; err != nil {
				logger.SysError("failed to seed model_metadata: " + def.DisplayName + "/" + lang + " - " + err.Error())
			}
		}
	}

	logger.SysLog("model_metadata seeded: " + formatInt(len(defs)) + " models × 7 languages")
}
