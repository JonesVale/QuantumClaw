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
			"通用对话、编程辅助、数据分析、多模态任务", false, "",
			"最快响应速度, 全模态输入(文本/图像/音频), 128K上下文, 性价比极高"},
		{"GPT-4o-mini", "2025-04", map[string]float64{"mmlu": 0.821, "human_eval": 0.872, "gsm8k": 0.877},
			[]string{"Chat", "Vision", "Code Generation", "Function Calling"},
			"轻量级对话、分类、提取、摘要", false, "",
			"极低成本, 快速响应, 适合大规模部署"},
		{"Claude 3.5 Sonnet", "2025-01", map[string]float64{"mmlu": 0.884, "human_eval": 0.896, "gsm8k": 0.942},
			[]string{"Chat", "Vision", "Code Generation", "Reasoning", "File Analysis"},
			"文档分析、代码审查、研究报告、复杂推理", false, "",
			"200K超长上下文, 文件输入支持, 出色的代码和文档理解能力"},
		{"Claude 3.5 Haiku", "2025-01", map[string]float64{"mmlu": 0.829, "gsm8k": 0.886},
			[]string{"Chat", "Vision", "Code Generation", "File Analysis"},
			"日常对话、内容生成、翻译、摘要", false, "",
			"快速响应, 低成本, 200K上下文"},
		{"Claude Opus 4", "2025-04", map[string]float64{"mmlu": 0.899, "human_eval": 0.927, "gsm8k": 0.958},
			[]string{"Chat", "Vision", "Code Generation", "Reasoning", "File Analysis", "Tool Use"},
			"复杂分析、企业应用、研究级推理", false, "",
			"Anthropic最强大模型, 顶级推理能力, 工具调用"},
		{"Gemini 2.0 Flash", "2025-05", map[string]float64{"mmlu": 0.886, "human_eval": 0.848, "gsm8k": 0.941},
			[]string{"Chat", "Vision", "Code Generation", "Audio/Video Understanding"},
			"视频分析、音频处理、多模态内容理解", false, "",
			"100万上下文窗口(业界最长), 原生多模态, 极速响应"},
		{"Gemini 2.0 Pro", "2025-05", map[string]float64{"mmlu": 0.907, "human_eval": 0.892, "gsm8k": 0.954},
			[]string{"Chat", "Vision", "Code Generation", "Reasoning", "Audio/Video Understanding"},
			"多模态研究、大规模数据分析、复杂推理", false, "",
			"顶级多模态理解, 1M上下文, 全模态输入"},
		{"DeepSeek Chat", "2025-03", map[string]float64{"mmlu": 0.843, "human_eval": 0.819, "gsm8k": 0.887},
			[]string{"Chat", "Code Generation", "Reasoning"},
			"通用对话、编程辅助、知识问答", true, "MIT",
			"极高性价比, 开源可自部署, 中英双语能力优秀"},
		{"DeepSeek R1", "2025-03", map[string]float64{"mmlu": 0.901, "human_eval": 0.967, "gsm8k": 0.959, "math": 0.976},
			[]string{"Reasoning", "Code Generation", "Math", "Problem Solving"},
			"数学推理、编程竞赛、逻辑分析、科学计算", true, "MIT",
			"开源中最强推理能力, 逐步思考链, 数学和代码表现领先"},
		{"DeepSeek V3", "2025-03", map[string]float64{"mmlu": 0.878, "human_eval": 0.893, "gsm8k": 0.925},
			[]string{"Chat", "Code Generation", "Reasoning"},
			"代码生成、编程辅助、复杂推理", true, "MIT",
			"最新代码优化模型, 64K上下文"},
		{"Qwen Max", "2025-03", map[string]float64{"mmlu": 0.861, "human_eval": 0.783, "gsm8k": 0.878},
			[]string{"Chat", "Code Generation", "Multilingual"},
			"中文对话、客户服务、内容创作", false, "",
			"中文理解能力优秀, 阿里生态集成"},
		{"Qwen Plus", "2025-03", map[string]float64{"mmlu": 0.845, "gsm8k": 0.872},
			[]string{"Chat", "Code Generation", "Multilingual"},
			"中文对话、内容生成", false, "",
			"128K上下文, 增强上下文理解"},
		{"Qwen 2.5 VL", "2025-03", map[string]float64{"mmlu": 0.847, "gsm8k": 0.892},
			[]string{"Vision", "Chat", "OCRs"},
			"图像理解、文档处理、多模态对话", true, "Apache 2.0",
			"中文图像理解领先, 文档OCR能力强"},
		{"Mistral Large", "2025-02", map[string]float64{"mmlu": 0.853, "human_eval": 0.835, "gsm8k": 0.872},
			[]string{"Chat", "Code Generation", "Multilingual", "Reasoning"},
			"多语言内容生成、结构化输出", false, "",
			"多语言支持出色, 原生函数调用, 128K上下文"},
		{"Mixtral 8x7B", "2024-10", map[string]float64{"mmlu": 0.826, "human_eval": 0.751, "gsm8k": 0.851},
			[]string{"Chat", "Multilingual"},
			"多语言对话、内容生成", true, "Apache 2.0",
			"MoE架构性价比高, 开源"},
		{"Llama 3.1 70B", "2024-12", map[string]float64{"mmlu": 0.858, "human_eval": 0.801, "gsm8k": 0.901},
			[]string{"Chat", "Code Generation", "Reasoning", "Tool Use"},
			"私有部署、定制微调、企业自托管", true, "Llama 3.1 License",
			"领先的开源模型, 128K上下文, 支持工具调用"},
		{"Llama 3.1 405B", "2024-12", map[string]float64{"mmlu": 0.874, "human_eval": 0.841, "gsm8k": 0.921},
			[]string{"Chat", "Code Generation", "Reasoning", "Tool Use"},
			"最高质量开源模型, 研究、企业部署", true, "Llama 3.1 License",
			"接近GPT-4水平, 最大开源模型"},
		{"Llama 3.2 Vision", "2024-12", map[string]float64{"mmlu": 0.832, "gsm8k": 0.879},
			[]string{"Vision", "Chat", "Code Generation"},
			"图像理解、文档OCR、视觉问答", true, "Llama 3.2 License",
			"开源多模态视觉模型, 可定制微调"},
		{"o1", "2025-04", map[string]float64{"mmlu": 0.921, "human_eval": 0.924, "gsm8k": 0.968, "math": 0.942},
			[]string{"Reasoning", "Code Generation", "Math", "Problem Solving", "Vision"},
			"高难度推理、科学研究、数学证明", false, "",
			"深度推理能力, 200K上下文, 思考链推理"},
		{"o3", "2025-04", map[string]float64{"mmlu": 0.933, "human_eval": 0.971, "gsm8k": 0.989, "math": 0.967, "arc": 0.957},
			[]string{"Reasoning", "Code Generation", "Math", "Problem Solving", "Vision", "Scientific Research"},
			"前沿AI研究、代码竞赛、科学发现", false, "",
			"OpenAI最强推理能力, 领先所有基准"},
		{"Codestral", "2025-01", map[string]float64{"human_eval": 0.882, "mbpp": 0.856},
			[]string{"Code Generation", "Code Completion", "Code Review"},
			"代码补全、代码生成、代码审查", false, "",
			"专注代码任务, 支持80+编程语言"},
		{"GPT-4 Turbo", "2024-12", map[string]float64{"mmlu": 0.864, "human_eval": 0.832, "gsm8k": 0.915},
			[]string{"Chat", "Vision", "Code Generation"},
			"复杂任务、长文档分析", false, "",
			"支持视觉输入, 128K上下文"},
		{"GPT-3.5 Turbo", "2024-09", map[string]float64{"mmlu": 0.813, "gsm8k": 0.832},
			[]string{"Chat"},
			"简单对话、分类、提取", false, "",
			"快速经济, 适合简单对话任务"},
		{"Gemini 3.0 Pro", "2025-05", map[string]float64{"mmlu": 0.915, "human_eval": 0.903, "gsm8k": 0.962},
			[]string{"Chat", "Vision", "Code Generation", "Reasoning", "Audio/Video Understanding"},
			"先进推理、多模态分析", false, "",
			"1M上下文, 增强分析能力"},
	}

	updated := 0
	for _, d := range details {
		benchJSON, _ := json.Marshal(d.benchmarks)
		capsJSON, _ := json.Marshal(d.caps)

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
