package model

import (
	"fmt"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

// LanguageType — 语言版本类型，LanguagesType = 显示名 = 查询条件
type LanguageType struct {
	Id            int    `json:"id" gorm:"primaryKey;autoIncrement"`
	LanguagesType string `json:"languages_type" gorm:"type:varchar(64);uniqueIndex"`
	AddTime       string `json:"add_time" gorm:"type:varchar(32)"`
}

// LanguageEntry — T_Languages 翻译条目
type LanguageEntry struct {
	Id            int    `json:"id" gorm:"primaryKey;autoIncrement"`
	LanguagesType string `json:"languages_type" gorm:"type:varchar(64);index"`
	LCode         string `json:"lcode" gorm:"type:varchar(128);index"`
	Display       string `json:"display" gorm:"type:text"`
}

// LanguageResource — 语言资源（兼容 model/main.go AutoMigrate 调用）
type LanguageResource struct {
	Id     int    `json:"id" gorm:"primaryKey;autoIncrement"`
	LCode  string `json:"lcode" gorm:"type:varchar(128);uniqueIndex"`
	CN     string `json:"cn" gorm:"type:text"`
	TW     string `json:"tw" gorm:"type:text"`
	EN     string `json:"en" gorm:"type:text"`
	JA     string `json:"ja" gorm:"type:text"`
	FR     string `json:"fr" gorm:"type:text"`
	RU     string `json:"ru" gorm:"type:text"`
	VI     string `json:"vi" gorm:"type:text"`
}

// InitLanguageTypes — 初始化语言版本（兼容原 model/main.go 调用）
func InitLanguageTypes() {
	logger.SysLog("InitLanguageTypes called")
}

// InitChineseLanguageResources — 初始化中文资源（兼容原 model/main.go 调用）
func InitChineseLanguageResources() {
	logger.SysLog("InitChineseLanguageResources called")
}

// InitLanguageTables 确保语言相关表存在，种子不存在时插入
func InitLanguageTables() {
	if err := DB.AutoMigrate(&LanguageType{}, &LanguageEntry{}); err != nil {
		logger.SysError("failed to migrate language tables: " + err.Error())
		return
	}

	var typeCount int64
	DB.Model(&LanguageType{}).Count(&typeCount)
	if typeCount == 0 {
		langs := []LanguageType{
			{LanguagesType: "中文简体", AddTime: "2026-05-20"},
			{LanguagesType: "English", AddTime: "2026-05-20"},
		}
		for _, l := range langs {
			DB.Create(&l)
		}
		logger.SysLog("language types seeded: 中文简体 + English")
	}

	// Seed LanguageEntry if empty
	var entryCount int64
	DB.Model(&LanguageEntry{}).Count(&entryCount)
	if entryCount > 0 {
		return
	}

	// Keys shared across English + Chinese (LCode = key, Display = translated text)
	type entry struct{ Lang, Key, Display string }
	en := func(k, v string) entry { return entry{"English", k, v} }
	zh := func(k, v string) entry { return entry{"中文简体", k, v} }

	enries := []entry{
		// Nav
		en("Models", "Models"), zh("Models", "模型"),
		en("Pricing", "Pricing"), zh("Pricing", "定价"),
		en("Rankings", "Rankings"), zh("Rankings", "排名"),
		en("Apps", "Apps"), zh("Apps", "应用"),
		en("Enterprise", "Enterprise"), zh("Enterprise", "企业版"),
		en("Language", "Language"), zh("Language", "语言"),
		en("Dashboard", "Dashboard"), zh("Dashboard", "控制台"),
		en("Sign In", "Sign In"), zh("Sign In", "登录"),
		en("Get Started", "Get Started"), zh("Get Started", "开始使用"),
		// Home / common
		en("AI Model Catalog", "AI Model Catalog"), zh("AI Model Catalog", "AI 模型目录"),
		en("Browse", "Browse"), zh("Browse", "浏览"),
		en("All Models", "All Models"), zh("All Models", "全部模型"),
		en("Chat", "Chat"), zh("Chat", "对话"),
		en("Code", "Code"), zh("Code", "代码"),
		en("Reasoning", "Reasoning"), zh("Reasoning", "推理"),
		en("Vision", "Vision"), zh("Vision", "视觉"),
		en("Providers", "Providers"), zh("Providers", "提供商"),
		en("Context", "Context"), zh("Context", "上下文"),
		en("Search", "Search"), zh("Search", "搜索"),
		en("Compare", "Compare"), zh("Compare", "对比"),
		en("Details", "Details"), zh("Details", "详情"),
		en("Call", "Call"), zh("Call", "调用"),
		en("Free", "Free"), zh("Free", "免费"),
		en("Active", "Active"), zh("Active", "启用"),
		en("Inactive", "Inactive"), zh("Inactive", "停用"),
		en("Name", "Name"), zh("Name", "名称"),
		en("Price: Low", "Price: Low"), zh("Price: Low", "价格从低到高"),
		en("Price: High", "Price: High"), zh("Price: High", "价格从高到低"),
		en("All Providers", "All Providers"), zh("All Providers", "全部提供商"),
		en("Active only", "Active only"), zh("Active only", "仅显示启用的"),
		en("Clear all", "Clear all"), zh("Clear all", "清除所有"),
		en("Reset filters", "Reset filters"), zh("Reset filters", "重置筛选"),
		en("No models found", "No models found"), zh("No models found", "未找到模型"),
		en("Try adjusting your filters", "Try adjusting your filters"), zh("Try adjusting your filters", "尝试调整筛选条件"),
		en("Show more", "Show more"), zh("Show more", "显示更多"),
		en("Filters", "Filters"), zh("Filters", "筛选"),
		en("Clear", "Clear"), zh("Clear", "清除"),
		en("showing", "showing"), zh("showing", "显示"),
		en("selected", "selected"), zh("selected", "已选"),
		en("models", "models"), zh("models", "个模型"),
		en("remaining", "remaining"), zh("remaining", "剩余"),
		// Pricing
		en("Model Pricing", "Model Pricing"), zh("Model Pricing", "模型定价"),
		en("Transparent Pricing", "Transparent Pricing"), zh("Transparent Pricing", "透明定价"),
		en("Input / 1K tokens", "Input / 1K tokens"), zh("Input / 1K tokens", "输入 / 1K Token"),
		en("Output / 1K tokens", "Output / 1K tokens"), zh("Output / 1K tokens", "输出 / 1K Token"),
		en("Refresh", "Refresh"), zh("Refresh", "刷新"),
		// Rankings
		en("Model Rankings", "Model Rankings"), zh("Model Rankings", "模型排名"),
		en("Real-time Rankings", "Real-time Rankings"), zh("Real-time Rankings", "实时排名"),
		en("By Requests", "By Requests"), zh("By Requests", "按请求量"),
		en("By Tokens", "By Tokens"), zh("By Tokens", "按Token数"),
		en("By Speed", "By Speed"), zh("By Speed", "按速度"),
		en("By Price", "By Price"), zh("By Price", "按价格"),
		en("Value", "Value"), zh("Value", "值"),
		en("Top 50 models", "Top 50 models"), zh("Top 50 models", "前 50 个模型"),
		en("No ranking data yet", "No ranking data yet"), zh("No ranking data yet", "暂无排名数据"),
		// Apps
		en("Apps & Integrations", "Apps & Integrations"), zh("Apps & Integrations", "应用与集成"),
		en("All", "All"), zh("All", "全部"),
		en("Featured", "Featured"), zh("Featured", "推荐"),
		en("No apps found", "No apps found"), zh("No apps found", "未找到应用"),
		en("Submit your app", "Submit your app"), zh("Submit your app", "提交应用"),
		en("users", "users"), zh("users", "用户"),
		en("integrations", "integrations"), zh("integrations", "集成"),
		// Enterprise
		en("Built for Scale", "Built for Scale"), zh("Built for Scale", "为规模化而生"),
		en("Contact Sales", "Contact Sales"), zh("Contact Sales", "联系销售"),
		en("Ready to scale?", "Ready to scale?"), zh("Ready to scale?", "准备好扩展了吗？"),
		// Sign In
		en("Access your QuantumClaw dashboard", "Access your QuantumClaw dashboard"), zh("Access your QuantumClaw dashboard", "访问您的 QuantumClaw 控制台"),
		en("Email", "Email"), zh("Email", "邮箱"),
		en("Password", "Password"), zh("Password", "密码"),
		en("Login failed", "Login failed"), zh("Login failed", "登录失败"),
		en("Network error", "Network error"), zh("Network error", "网络错误"),
		// Buttons & common
		en("Coming Soon", "Coming Soon"), zh("Coming Soon", "即将推出"),
		en("Add", "Add"), zh("Add", "添加"),
		en("Save", "Save"), zh("Save", "保存"),
		en("Cancel", "Cancel"), zh("Cancel", "取消"),
		en("Delete", "Delete"), zh("Delete", "删除"),
		en("Edit", "Edit"), zh("Edit", "编辑"),
		en("Back", "Back"), zh("Back", "返回"),
		en("Next", "Next"), zh("Next", "下一步"),
	}

	for _, e := range enries {
		DB.Create(&LanguageEntry{LanguagesType: e.Lang, LCode: e.Key, Display: e.Display})
	}
	logger.SysLog("language entries seeded: " + fmt.Sprint(len(enries)) + " rows")
}

// GetLanguageTypes 获取所有可用语言列表
func GetLanguageTypes() ([]LanguageType, error) {
	var types []LanguageType
	err := DB.Order("id asc").Find(&types).Error
	return types, err
}

// GetTranslationsByLanguage 获取指定语言的所有翻译键值对
func GetTranslationsByLanguage(lang string) (map[string]string, error) {
	var entries []LanguageEntry
	err := DB.Where("languages_type = ?", lang).Find(&entries).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(entries))
	for _, e := range entries {
		result[e.LCode] = e.Display
	}
	return result, nil
}
