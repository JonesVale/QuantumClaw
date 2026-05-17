package model

import (
	"errors"
	"fmt"
)

type LanguageType struct {
	ID           int    `gorm:"primaryKey;autoIncrement" json:"id"`
	LanguageCode string `gorm:"unique;size:10" json:"language_code"`
	LanguageName string `gorm:"size:50" json:"language_name"`
	IsDefault    bool   `gorm:"default:false" json:"is_default"`
	CreatedAt    int64  `gorm:"autoCreateTime" json:"created_at"`
}

type LanguageResource struct {
	ID           int    `gorm:"primaryKey;autoIncrement" json:"items"`
	LCode        string `gorm:"column:l_code;size:100" json:"lcode"`
	Display      string `gorm:"size:500" json:"display"`
	FromName     string `gorm:"size:100" json:"from_name"`
	LanguageType string `gorm:"size:10" json:"language_type"`
	CN           string `gorm:"size:500" json:"cn"`
}

var CurrentLanguage string = "en"

func InitLanguageTypes() {
	defaultLangs := []LanguageType{
		{LanguageCode: "en", LanguageName: "English", IsDefault: true},
		{LanguageCode: "zh-CN", LanguageName: "简体中文", IsDefault: false},
		{LanguageCode: "zh-TW", LanguageName: "繁体中文", IsDefault: false},
		{LanguageCode: "fr", LanguageName: "Français", IsDefault: false},
		{LanguageCode: "ja", LanguageName: "日本語", IsDefault: false},
		{LanguageCode: "ru", LanguageName: "Русский", IsDefault: false},
		{LanguageCode: "vi", LanguageName: "Tiếng Việt", IsDefault: false},
	}

	for _, lang := range defaultLangs {
		var existing LanguageType
		if err := DB.Where("language_code = ?", lang.LanguageCode).First(&existing).Error; err != nil {
			DB.Create(&lang)
		}
	}
}

func SetCurrentLanguage(langCode string) error {
	var langType LanguageType
	if err := DB.Where("language_code = ?", langCode).First(&langType).Error; err != nil {
		return fmt.Errorf("language not found: %w", err)
	}
	CurrentLanguage = langCode
	return nil
}

func GetCurrentLanguage() string {
	return CurrentLanguage
}

func GetAllLanguageTypes() ([]LanguageType, error) {
	var langs []LanguageType
	err := DB.Find(&langs).Error
	return langs, err
}

func AddLanguageType(languageCode, languageName string) error {
	var existing LanguageType
	if err := DB.Where("language_code = ?", languageCode).First(&existing).Error; err == nil {
		return errors.New("language already exists")
	}

	langType := LanguageType{
		LanguageCode: languageCode,
		LanguageName: languageName,
		IsDefault:    false,
	}

	return DB.Create(&langType).Error
}

func DeleteLanguageType(languageCode string) error {
	if languageCode == "en" {
		return errors.New("cannot delete default language")
	}
	return DB.Where("language_code = ?", languageCode).Delete(&LanguageType{}).Error
}

func GetLanguageResources(languageType string) ([]LanguageResource, error) {
	var resources []LanguageResource
	err := DB.Where("language_type = ?", languageType).Find(&resources).Error
	return resources, err
}

func GetLanguageResourceByLCode(languageType, lcode string) (*LanguageResource, error) {
	var resource LanguageResource
	err := DB.Where("language_type = ? AND l_code = ?", languageType, lcode).First(&resource).Error
	return &resource, err
}

func UpdateLanguageResourceDisplay(items int, display string) error {
	return DB.Model(&LanguageResource{}).Where("id = ?", items).Update("display", display).Error
}

func AddLanguageResource(resource LanguageResource) error {
	return DB.Create(&resource).Error
}

func GetResourceDisplay(lcode, fromName string) string {
	var resource LanguageResource
	err := DB.Where("language_type = ? AND l_code = ? AND from_name = ?", CurrentLanguage, lcode, fromName).First(&resource).Error
	if err != nil {
		return lcode
	}
	if resource.Display != "" {
		return resource.Display
	}
	return resource.CN
}

func InitChineseLanguageResources() {
	resources := []LanguageResource{
		{LCode: "QuantumClaw", Display: "QuantumClaw", FromName: "system", LanguageType: "zh-CN", CN: "QuantumClaw"},
		{LCode: "Sign In", Display: "登录", FromName: "auth", LanguageType: "zh-CN", CN: "登录"},
		{LCode: "Sign Up", Display: "注册", FromName: "auth", LanguageType: "zh-CN", CN: "注册"},
		{LCode: "Sign Out", Display: "退出登录", FromName: "auth", LanguageType: "zh-CN", CN: "退出登录"},
		{LCode: "Dashboard", Display: "仪表盘", FromName: "nav", LanguageType: "zh-CN", CN: "仪表盘"},
		{LCode: "Overview", Display: "概览", FromName: "nav", LanguageType: "zh-CN", CN: "概览"},
		{LCode: "Channels", Display: "渠道管理", FromName: "nav", LanguageType: "zh-CN", CN: "渠道管理"},
		{LCode: "API Keys", Display: "API 密钥", FromName: "nav", LanguageType: "zh-CN", CN: "API 密钥"},
		{LCode: "Models", Display: "模型管理", FromName: "nav", LanguageType: "zh-CN", CN: "模型管理"},
		{LCode: "Users", Display: "用户管理", FromName: "nav", LanguageType: "zh-CN", CN: "用户管理"},
		{LCode: "Usage Logs", Display: "使用日志", FromName: "nav", LanguageType: "zh-CN", CN: "使用日志"},
		{LCode: "Redemption Codes", Display: "兑换码", FromName: "nav", LanguageType: "zh-CN", CN: "兑换码"},
		{LCode: "Wallet", Display: "钱包", FromName: "nav", LanguageType: "zh-CN", CN: "钱包"},
		{LCode: "Profile", Display: "个人资料", FromName: "nav", LanguageType: "zh-CN", CN: "个人资料"},
		{LCode: "Settings", Display: "设置", FromName: "nav", LanguageType: "zh-CN", CN: "设置"},
		{LCode: "System Settings", Display: "系统设置", FromName: "nav", LanguageType: "zh-CN", CN: "系统设置"},
		{LCode: "Session expired!", Display: "会话已过期！", FromName: "msg", LanguageType: "zh-CN", CN: "会话已过期！"},
		{LCode: "Internal Server Error!", Display: "服务器内部错误！", FromName: "msg", LanguageType: "zh-CN", CN: "服务器内部错误！"},
		{LCode: "Content not modified!", Display: "内容未修改！", FromName: "msg", LanguageType: "zh-CN", CN: "内容未修改！"},
		{LCode: "Playground", Display: "在线测试", FromName: "page", LanguageType: "zh-CN", CN: "在线测试"},
		{LCode: "Chat", Display: "对话", FromName: "page", LanguageType: "zh-CN", CN: "对话"},
		{LCode: "General", Display: "常规", FromName: "cat", LanguageType: "zh-CN", CN: "常规"},
		{LCode: "Personal", Display: "个人", FromName: "cat", LanguageType: "zh-CN", CN: "个人"},
		{LCode: "Admin", Display: "管理员", FromName: "cat", LanguageType: "zh-CN", CN: "管理员"},
		{LCode: "Task Logs", Display: "任务日志", FromName: "nav", LanguageType: "zh-CN", CN: "任务日志"},
		{LCode: "Subscription Management", Display: "订阅管理", FromName: "nav", LanguageType: "zh-CN", CN: "订阅管理"},
		{LCode: "Username", Display: "用户名", FromName: "form", LanguageType: "zh-CN", CN: "用户名"},
		{LCode: "Password", Display: "密码", FromName: "form", LanguageType: "zh-CN", CN: "密码"},
		{LCode: "Cancel", Display: "取消", FromName: "btn", LanguageType: "zh-CN", CN: "取消"},
		{LCode: "Save", Display: "保存", FromName: "btn", LanguageType: "zh-CN", CN: "保存"},
		{LCode: "Create", Display: "创建", FromName: "btn", LanguageType: "zh-CN", CN: "创建"},
		{LCode: "Edit", Display: "编辑", FromName: "btn", LanguageType: "zh-CN", CN: "编辑"},
		{LCode: "Delete", Display: "删除", FromName: "btn", LanguageType: "zh-CN", CN: "删除"},
		{LCode: "Search", Display: "搜索", FromName: "btn", LanguageType: "zh-CN", CN: "搜索"},
		{LCode: "Refresh", Display: "刷新", FromName: "btn", LanguageType: "zh-CN", CN: "刷新"},
		{LCode: "Previous", Display: "上一页", FromName: "btn", LanguageType: "zh-CN", CN: "上一页"},
		{LCode: "Next", Display: "下一页", FromName: "btn", LanguageType: "zh-CN", CN: "下一页"},
		{LCode: "Active", Display: "正常", FromName: "status", LanguageType: "zh-CN", CN: "正常"},
		{LCode: "Disabled", Display: "已禁用", FromName: "status", LanguageType: "zh-CN", CN: "已禁用"},
		{LCode: "Enable", Display: "启用", FromName: "action", LanguageType: "zh-CN", CN: "启用"},
		{LCode: "Disable", Display: "禁用", FromName: "action", LanguageType: "zh-CN", CN: "禁用"},
		{LCode: "Close", Display: "关闭", FromName: "btn", LanguageType: "zh-CN", CN: "关闭"},
		{LCode: "Copy", Display: "复制", FromName: "btn", LanguageType: "zh-CN", CN: "复制"},
		{LCode: "Loading...", Display: "加载中...", FromName: "state", LanguageType: "zh-CN", CN: "加载中..."},
		{LCode: "Saving...", Display: "保存中...", FromName: "state", LanguageType: "zh-CN", CN: "保存中..."},
		{LCode: "Updating...", Display: "更新中...", FromName: "state", LanguageType: "zh-CN", CN: "更新中..."},
		{LCode: "Creating...", Display: "创建中...", FromName: "state", LanguageType: "zh-CN", CN: "创建中..."},
		{LCode: "Sign in successful", Display: "登录成功", FromName: "msg", LanguageType: "zh-CN", CN: "登录成功"},
		{LCode: "Sign in failed", Display: "登录失败", FromName: "msg", LanguageType: "zh-CN", CN: "登录失败"},
		{LCode: "Signing in...", Display: "登录中...", FromName: "state", LanguageType: "zh-CN", CN: "登录中..."},
		{LCode: "Please fill in all fields", Display: "请填写所有字段", FromName: "msg", LanguageType: "zh-CN", CN: "请填写所有字段"},
		{LCode: "Enter your credentials to access your account", Display: "输入您的凭据以访问您的账户", FromName: "msg", LanguageType: "zh-CN", CN: "输入您的凭据以访问您的账户"},
		{LCode: "Please save your API key, it will not be shown again", Display: "请保存您的 API 密钥，它不会再次显示", FromName: "msg", LanguageType: "zh-CN", CN: "请保存您的 API 密钥，它不会再次显示"},
		{LCode: "Copied to clipboard", Display: "已复制到剪贴板", FromName: "msg", LanguageType: "zh-CN", CN: "已复制到剪贴板"},
		{LCode: "Are you sure you want to delete this channel?", Display: "确定要删除此渠道吗？", FromName: "msg", LanguageType: "zh-CN", CN: "确定要删除此渠道吗？"},
		{LCode: "Are you sure?", Display: "确定吗？", FromName: "msg", LanguageType: "zh-CN", CN: "确定吗？"},
		{LCode: "Operation failed", Display: "操作失败", FromName: "msg", LanguageType: "zh-CN", CN: "操作失败"},
		{LCode: "No channels found", Display: "未找到渠道", FromName: "msg", LanguageType: "zh-CN", CN: "未找到渠道"},
		{LCode: "No tokens found", Display: "未找到令牌", FromName: "msg", LanguageType: "zh-CN", CN: "未找到令牌"},
		{LCode: "No users found", Display: "未找到用户", FromName: "msg", LanguageType: "zh-CN", CN: "未找到用户"},
		{LCode: "No logs found", Display: "未找到日志", FromName: "msg", LanguageType: "zh-CN", CN: "未找到日志"},
		{LCode: "No models available", Display: "暂无可用模型", FromName: "msg", LanguageType: "zh-CN", CN: "暂无可用模型"},
		{LCode: "No redemption codes", Display: "暂无兑换码", FromName: "msg", LanguageType: "zh-CN", CN: "暂无兑换码"},
		{LCode: "Total Users", Display: "总用户数", FromName: "stat", LanguageType: "zh-CN", CN: "总用户数"},
		{LCode: "Active Channels", Display: "活跃渠道", FromName: "stat", LanguageType: "zh-CN", CN: "活跃渠道"},
		{LCode: "Today Requests", Display: "今日请求", FromName: "stat", LanguageType: "zh-CN", CN: "今日请求"},
		{LCode: "Active Tokens", Display: "活跃令牌", FromName: "stat", LanguageType: "zh-CN", CN: "活跃令牌"},
		{LCode: "Registered accounts", Display: "已注册账户", FromName: "stat", LanguageType: "zh-CN", CN: "已注册账户"},
		{LCode: "Of {{total}} total", Display: "共 {{total}} 个", FromName: "stat", LanguageType: "zh-CN", CN: "共 {{total}} 个"},
		{LCode: "API calls today", Display: "今日 API 调用", FromName: "stat", LanguageType: "zh-CN", CN: "今日 API 调用"},
		{LCode: "API keys issued", Display: "已发放密钥", FromName: "stat", LanguageType: "zh-CN", CN: "已发放密钥"},
		{LCode: "System Status", Display: "系统状态", FromName: "stat", LanguageType: "zh-CN", CN: "系统状态"},
		{LCode: "Backend API", Display: "后端 API", FromName: "stat", LanguageType: "zh-CN", CN: "后端 API"},
		{LCode: "Online", Display: "在线", FromName: "stat", LanguageType: "zh-CN", CN: "在线"},
		{LCode: "System Version", Display: "系统版本", FromName: "stat", LanguageType: "zh-CN", CN: "系统版本"},
		{LCode: "Account Type", Display: "账户类型", FromName: "stat", LanguageType: "zh-CN", CN: "账户类型"},
		{LCode: "Quick Actions", Display: "快捷操作", FromName: "stat", LanguageType: "zh-CN", CN: "快捷操作"},
		{LCode: "Common tasks and shortcuts", Display: "常用任务和快捷方式", FromName: "desc", LanguageType: "zh-CN", CN: "常用任务和快捷方式"},
		{LCode: "Welcome back, {{name}}", Display: "欢迎回来，{{name}}", FromName: "msg", LanguageType: "zh-CN", CN: "欢迎回来，{{name}}"},
		{LCode: "System overview", Display: "系统概览", FromName: "desc", LanguageType: "zh-CN", CN: "系统概览"},
		{LCode: "Channel Name", Display: "渠道名称", FromName: "form", LanguageType: "zh-CN", CN: "渠道名称"},
		{LCode: "Channel Type", Display: "渠道类型", FromName: "form", LanguageType: "zh-CN", CN: "渠道类型"},
		{LCode: "API Key", Display: "API 密钥", FromName: "form", LanguageType: "zh-CN", CN: "API 密钥"},
		{LCode: "Base URL", Display: "基础 URL", FromName: "form", LanguageType: "zh-CN", CN: "基础 URL"},
		{LCode: "Name", Display: "名称", FromName: "form", LanguageType: "zh-CN", CN: "名称"},
		{LCode: "Type", Display: "类型", FromName: "form", LanguageType: "zh-CN", CN: "类型"},
		{LCode: "Status", Display: "状态", FromName: "form", LanguageType: "zh-CN", CN: "状态"},
		{LCode: "Models", Display: "模型", FromName: "form", LanguageType: "zh-CN", CN: "模型"},
		{LCode: "Response Time", Display: "响应时间", FromName: "form", LanguageType: "zh-CN", CN: "响应时间"},
		{LCode: "Updated", Display: "更新时间", FromName: "form", LanguageType: "zh-CN", CN: "更新时间"},
		{LCode: "Test", Display: "测试", FromName: "btn", LanguageType: "zh-CN", CN: "测试"},
		{LCode: "Channel created", Display: "渠道已创建", FromName: "msg", LanguageType: "zh-CN", CN: "渠道已创建"},
		{LCode: "Channel updated", Display: "渠道已更新", FromName: "msg", LanguageType: "zh-CN", CN: "渠道已更新"},
		{LCode: "Channel deleted", Display: "渠道已删除", FromName: "msg", LanguageType: "zh-CN", CN: "渠道已删除"},
		{LCode: "Channel test passed", Display: "渠道测试通过", FromName: "msg", LanguageType: "zh-CN", CN: "渠道测试通过"},
		{LCode: "Channel test failed", Display: "渠道测试失败", FromName: "msg", LanguageType: "zh-CN", CN: "渠道测试失败"},
		{LCode: "Failed to create channel", Display: "创建渠道失败", FromName: "msg", LanguageType: "zh-CN", CN: "创建渠道失败"},
		{LCode: "Failed to update channel", Display: "更新渠道失败", FromName: "msg", LanguageType: "zh-CN", CN: "更新渠道失败"},
		{LCode: "Failed to delete channel", Display: "删除渠道失败", FromName: "msg", LanguageType: "zh-CN", CN: "删除渠道失败"},
		{LCode: "Create Channel", Display: "创建渠道", FromName: "page", LanguageType: "zh-CN", CN: "创建渠道"},
		{LCode: "Edit Channel", Display: "编辑渠道", FromName: "page", LanguageType: "zh-CN", CN: "编辑渠道"},
		{LCode: "Update channel configuration", Display: "更新渠道配置", FromName: "desc", LanguageType: "zh-CN", CN: "更新渠道配置"},
		{LCode: "Add a new API provider channel", Display: "添加新的 API 提供商渠道", FromName: "desc", LanguageType: "zh-CN", CN: "添加新的 API 提供商渠道"},
		{LCode: "Manage API provider channels", Display: "管理 API 提供商渠道", FromName: "desc", LanguageType: "zh-CN", CN: "管理 API 提供商渠道"},
		{LCode: "Search channels...", Display: "搜索渠道...", FromName: "hint", LanguageType: "zh-CN", CN: "搜索渠道..."},
		{LCode: "Token Name", Display: "令牌名称", FromName: "form", LanguageType: "zh-CN", CN: "令牌名称"},
		{LCode: "Key", Display: "密钥", FromName: "form", LanguageType: "zh-CN", CN: "密钥"},
		{LCode: "Quota", Display: "配额", FromName: "form", LanguageType: "zh-CN", CN: "配额"},
		{LCode: "Created", Display: "创建时间", FromName: "form", LanguageType: "zh-CN", CN: "创建时间"},
		{LCode: "Unlimited Quota", Display: "无限配额", FromName: "form", LanguageType: "zh-CN", CN: "无限配额"},
		{LCode: "Unlimited", Display: "无限", FromName: "form", LanguageType: "zh-CN", CN: "无限"},
		{LCode: "Leave empty for all models", Display: "留空表示所有模型", FromName: "hint", LanguageType: "zh-CN", CN: "留空表示所有模型"},
		{LCode: "Token updated", Display: "令牌已更新", FromName: "msg", LanguageType: "zh-CN", CN: "令牌已更新"},
		{LCode: "Token created", Display: "令牌已创建", FromName: "msg", LanguageType: "zh-CN", CN: "令牌已创建"},
		{LCode: "Token deleted", Display: "令牌已删除", FromName: "msg", LanguageType: "zh-CN", CN: "令牌已删除"},
		{LCode: "Token status updated", Display: "令牌状态已更新", FromName: "msg", LanguageType: "zh-CN", CN: "令牌状态已更新"},
		{LCode: "Failed to create token", Display: "创建令牌失败", FromName: "msg", LanguageType: "zh-CN", CN: "创建令牌失败"},
		{LCode: "Failed to update token", Display: "更新令牌失败", FromName: "msg", LanguageType: "zh-CN", CN: "更新令牌失败"},
		{LCode: "Create Token", Display: "创建令牌", FromName: "page", LanguageType: "zh-CN", CN: "创建令牌"},
		{LCode: "Edit Token", Display: "编辑令牌", FromName: "page", LanguageType: "zh-CN", CN: "编辑令牌"},
		{LCode: "Update API key configuration", Display: "更新 API 密钥配置", FromName: "desc", LanguageType: "zh-CN", CN: "更新 API 密钥配置"},
		{LCode: "Create a new API key", Display: "创建新的 API 密钥", FromName: "desc", LanguageType: "zh-CN", CN: "创建新的 API 密钥"},
		{LCode: "Manage API access tokens", Display: "管理 API 访问令牌", FromName: "desc", LanguageType: "zh-CN", CN: "管理 API 访问令牌"},
		{LCode: "Search tokens...", Display: "搜索令牌...", FromName: "hint", LanguageType: "zh-CN", CN: "搜索令牌..."},
		{LCode: "Display Name", Display: "显示名称", FromName: "form", LanguageType: "zh-CN", CN: "显示名称"},
		{LCode: "Role", Display: "角色", FromName: "form", LanguageType: "zh-CN", CN: "角色"},
		{LCode: "Requests", Display: "请求数", FromName: "form", LanguageType: "zh-CN", CN: "请求数"},
		{LCode: "User", Display: "用户", FromName: "form", LanguageType: "zh-CN", CN: "用户"},
		{LCode: "Super Admin", Display: "超级管理员", FromName: "form", LanguageType: "zh-CN", CN: "超级管理员"},
		{LCode: "User updated", Display: "用户已更新", FromName: "msg", LanguageType: "zh-CN", CN: "用户已更新"},
		{LCode: "User created", Display: "用户已创建", FromName: "msg", LanguageType: "zh-CN", CN: "用户已创建"},
		{LCode: "User deleted", Display: "用户已删除", FromName: "msg", LanguageType: "zh-CN", CN: "用户已删除"},
		{LCode: "User status updated", Display: "用户状态已更新", FromName: "msg", LanguageType: "zh-CN", CN: "用户状态已更新"},
		{LCode: "Failed to update user", Display: "更新用户失败", FromName: "msg", LanguageType: "zh-CN", CN: "更新用户失败"},
		{LCode: "Create User", Display: "创建用户", FromName: "page", LanguageType: "zh-CN", CN: "创建用户"},
		{LCode: "Edit User", Display: "编辑用户", FromName: "page", LanguageType: "zh-CN", CN: "编辑用户"},
		{LCode: "Update user information", Display: "更新用户信息", FromName: "desc", LanguageType: "zh-CN", CN: "更新用户信息"},
		{LCode: "Create a new user account", Display: "创建新的用户账户", FromName: "desc", LanguageType: "zh-CN", CN: "创建新的用户账户"},
		{LCode: "Manage user accounts", Display: "管理用户账户", FromName: "desc", LanguageType: "zh-CN", CN: "管理用户账户"},
		{LCode: "Search users...", Display: "搜索用户...", FromName: "hint", LanguageType: "zh-CN", CN: "搜索用户..."},
		{LCode: "Leave empty to keep unchanged", Display: "留空保持不变", FromName: "hint", LanguageType: "zh-CN", CN: "留空保持不变"},
		{LCode: "Email", Display: "电子邮箱", FromName: "form", LanguageType: "zh-CN", CN: "电子邮箱"},
		{LCode: "Available AI models across channels", Display: "所有渠道的可用 AI 模型", FromName: "desc", LanguageType: "zh-CN", CN: "所有渠道的可用 AI 模型"},
		{LCode: "Search models...", Display: "搜索模型...", FromName: "hint", LanguageType: "zh-CN", CN: "搜索模型..."},
		{LCode: "API request history and audit logs", Display: "API 请求历史和审计日志", FromName: "desc", LanguageType: "zh-CN", CN: "API 请求历史和审计日志"},
		{LCode: "Search logs...", Display: "搜索日志...", FromName: "hint", LanguageType: "zh-CN", CN: "搜索日志..."},
		{LCode: "All Types", Display: "所有类型", FromName: "filter", LanguageType: "zh-CN", CN: "所有类型"},
		{LCode: "Chat Completion", Display: "对话补全", FromName: "filter", LanguageType: "zh-CN", CN: "对话补全"},
		{LCode: "Completion", Display: "文本补全", FromName: "filter", LanguageType: "zh-CN", CN: "文本补全"},
		{LCode: "Embedding", Display: "嵌入", FromName: "filter", LanguageType: "zh-CN", CN: "嵌入"},
		{LCode: "Image", Display: "图像", FromName: "filter", LanguageType: "zh-CN", CN: "图像"},
		{LCode: "Audio", Display: "音频", FromName: "filter", LanguageType: "zh-CN", CN: "音频"},
		{LCode: "Tokens", Display: "Token数", FromName: "form", LanguageType: "zh-CN", CN: "Token数"},
		{LCode: "Time", Display: "时间", FromName: "form", LanguageType: "zh-CN", CN: "时间"},
		{LCode: "Page {{page}}", Display: "第 {{page}} 页", FromName: "pagin", LanguageType: "zh-CN", CN: "第 {{page}} 页"},
		{LCode: "Manage quota redemption codes", Display: "管理配额兑换码", FromName: "desc", LanguageType: "zh-CN", CN: "管理配额兑换码"},
		{LCode: "Create Codes", Display: "创建兑换码", FromName: "page", LanguageType: "zh-CN", CN: "创建兑换码"},
		{LCode: "Redemption codes created", Display: "兑换码已创建", FromName: "msg", LanguageType: "zh-CN", CN: "兑换码已创建"},
		{LCode: "Failed to create codes", Display: "创建兑换码失败", FromName: "msg", LanguageType: "zh-CN", CN: "创建兑换码失败"},
		{LCode: "Count", Display: "数量", FromName: "form", LanguageType: "zh-CN", CN: "数量"},
		{LCode: "Used", Display: "已使用", FromName: "form", LanguageType: "zh-CN", CN: "已使用"},
		{LCode: "Batch Name", Display: "批次名称", FromName: "form", LanguageType: "zh-CN", CN: "批次名称"},
		{LCode: "Code Count", Display: "兑换码数量", FromName: "form", LanguageType: "zh-CN", CN: "兑换码数量"},
		{LCode: "Quota Per Code", Display: "每个兑换码配额", FromName: "form", LanguageType: "zh-CN", CN: "每个兑换码配额"},
		{LCode: "Generate quota codes for users to redeem", Display: "生成配额兑换码供用户兑换", FromName: "desc", LanguageType: "zh-CN", CN: "生成配额兑换码供用户兑换"},
		{LCode: "Start a conversation", Display: "开始对话", FromName: "hint", LanguageType: "zh-CN", CN: "开始对话"},
		{LCode: "Type a message...", Display: "输入消息...", FromName: "hint", LanguageType: "zh-CN", CN: "输入消息..."},
		{LCode: "Stop", Display: "停止", FromName: "btn", LanguageType: "zh-CN", CN: "停止"},
		{LCode: "Clear", Display: "清除", FromName: "btn", LanguageType: "zh-CN", CN: "清除"},
		{LCode: "Select model", Display: "选择模型", FromName: "hint", LanguageType: "zh-CN", CN: "选择模型"},
		{LCode: "Manage your account settings", Display: "管理您的账户设置", FromName: "desc", LanguageType: "zh-CN", CN: "管理您的账户设置"},
		{LCode: "Personal Information", Display: "个人信息", FromName: "section", LanguageType: "zh-CN", CN: "个人信息"},
		{LCode: "Update your display name and email", Display: "更新您的显示名称和邮箱", FromName: "desc", LanguageType: "zh-CN", CN: "更新您的显示名称和邮箱"},
		{LCode: "Your display name", Display: "您的显示名称", FromName: "hint", LanguageType: "zh-CN", CN: "您的显示名称"},
		{LCode: "Email address", Display: "电子邮箱地址", FromName: "hint", LanguageType: "zh-CN", CN: "电子邮箱地址"},
		{LCode: "Profile updated", Display: "个人资料已更新", FromName: "msg", LanguageType: "zh-CN", CN: "个人资料已更新"},
		{LCode: "Failed to update profile", Display: "更新个人资料失败", FromName: "msg", LanguageType: "zh-CN", CN: "更新个人资料失败"},
		{LCode: "Change Password", Display: "修改密码", FromName: "section", LanguageType: "zh-CN", CN: "修改密码"},
		{LCode: "Update your password to keep your account secure", Display: "修改您的密码以保护账户安全", FromName: "desc", LanguageType: "zh-CN", CN: "修改您的密码以保护账户安全"},
		{LCode: "Current Password", Display: "当前密码", FromName: "form", LanguageType: "zh-CN", CN: "当前密码"},
		{LCode: "New Password", Display: "新密码", FromName: "form", LanguageType: "zh-CN", CN: "新密码"},
		{LCode: "Confirm New Password", Display: "确认新密码", FromName: "form", LanguageType: "zh-CN", CN: "确认新密码"},
		{LCode: "Passwords do not match", Display: "两次输入的密码不一致", FromName: "msg", LanguageType: "zh-CN", CN: "两次输入的密码不一致"},
		{LCode: "Update Password", Display: "修改密码", FromName: "btn", LanguageType: "zh-CN", CN: "修改密码"},
		{LCode: "Account Details", Display: "账户详情", FromName: "section", LanguageType: "zh-CN", CN: "账户详情"},
		{LCode: "Group", Display: "用户组", FromName: "form", LanguageType: "zh-CN", CN: "用户组"},
		{LCode: "Used Quota", Display: "已用配额", FromName: "form", LanguageType: "zh-CN", CN: "已用配额"},
		{LCode: "Remaining Quota", Display: "剩余配额", FromName: "form", LanguageType: "zh-CN", CN: "剩余配额"},
		{LCode: "Request Count", Display: "请求次数", FromName: "form", LanguageType: "zh-CN", CN: "请求次数"},
		{LCode: "Configure system-wide settings", Display: "配置系统全局设置", FromName: "desc", LanguageType: "zh-CN", CN: "配置系统全局设置"},
		{LCode: "Basic system configuration", Display: "基础系统配置", FromName: "desc", LanguageType: "zh-CN", CN: "基础系统配置"},
		{LCode: "System Name", Display: "系统名称", FromName: "form", LanguageType: "zh-CN", CN: "系统名称"},
		{LCode: "Displayed in the browser tab and login page", Display: "显示在浏览器标签和登录页面", FromName: "hint", LanguageType: "zh-CN", CN: "显示在浏览器标签和登录页面"},
		{LCode: "Logo URL", Display: "Logo 地址", FromName: "form", LanguageType: "zh-CN", CN: "Logo 地址"},
		{LCode: "Footer HTML", Display: "页脚 HTML", FromName: "form", LanguageType: "zh-CN", CN: "页脚 HTML"},
		{LCode: "Settings saved", Display: "设置已保存", FromName: "msg", LanguageType: "zh-CN", CN: "设置已保存"},
		{LCode: "Failed to save settings", Display: "保存设置失败", FromName: "msg", LanguageType: "zh-CN", CN: "保存设置失败"},
		{LCode: "Security", Display: "安全设置", FromName: "section", LanguageType: "zh-CN", CN: "安全设置"},
		{LCode: "Authentication and registration settings", Display: "认证和注册设置", FromName: "desc", LanguageType: "zh-CN", CN: "认证和注册设置"},
		{LCode: "Allow Registration", Display: "允许注册", FromName: "form", LanguageType: "zh-CN", CN: "允许注册"},
		{LCode: "Allow new users to self-register", Display: "允许新用户自行注册", FromName: "hint", LanguageType: "zh-CN", CN: "允许新用户自行注册"},
		{LCode: "Email Verification", Display: "邮箱验证", FromName: "form", LanguageType: "zh-CN", CN: "邮箱验证"},
		{LCode: "Require email verification for new users", Display: "要求新用户验证邮箱", FromName: "hint", LanguageType: "zh-CN", CN: "要求新用户验证邮箱"},
		{LCode: "Turnstile Check", Display: "Turnstile 验证", FromName: "form", LanguageType: "zh-CN", CN: "Turnstile 验证"},
		{LCode: "Enable Cloudflare Turnstile captcha", Display: "启用 Cloudflare Turnstile 验证码", FromName: "hint", LanguageType: "zh-CN", CN: "启用 Cloudflare Turnstile 验证码"},
		{LCode: "API Configuration", Display: "API 配置", FromName: "section", LanguageType: "zh-CN", CN: "API 配置"},
		{LCode: "API rate limits and access control", Display: "API 速率限制和访问控制", FromName: "desc", LanguageType: "zh-CN", CN: "API 速率限制和访问控制"},
		{LCode: "Global API Rate Limit (RPM)", Display: "全局 API 速率限制（RPM）", FromName: "form", LanguageType: "zh-CN", CN: "全局 API 速率限制（RPM）"},
		{LCode: "Quota Per Unit (Tokens)", Display: "每单位配额（Token数）", FromName: "form", LanguageType: "zh-CN", CN: "每单位配额（Token数）"},
		{LCode: "View and manage your quota balance", Display: "查看和管理您的配额余额", FromName: "desc", LanguageType: "zh-CN", CN: "查看和管理您的配额余额"},
		{LCode: "Quota Balance", Display: "配额余额", FromName: "form", LanguageType: "zh-CN", CN: "配额余额"},
		{LCode: "Total Quota", Display: "总配额", FromName: "form", LanguageType: "zh-CN", CN: "总配额"},
		{LCode: "Redeem Code", Display: "兑换码", FromName: "form", LanguageType: "zh-CN", CN: "兑换码"},
		{LCode: "Enter a redemption code to add quota", Display: "输入兑换码以添加配额", FromName: "hint", LanguageType: "zh-CN", CN: "输入兑换码以添加配额"},
		{LCode: "Enter redemption code", Display: "输入兑换码", FromName: "hint", LanguageType: "zh-CN", CN: "输入兑换码"},
		{LCode: "Redeem", Display: "兑换", FromName: "btn", LanguageType: "zh-CN", CN: "兑换"},
		{LCode: "Redemption successful! Quota added.", Display: "兑换成功！配额已添加。", FromName: "msg", LanguageType: "zh-CN", CN: "兑换成功！配额已添加。"},
		{LCode: "Redemption failed", Display: "兑换失败", FromName: "msg", LanguageType: "zh-CN", CN: "兑换失败"},
		{LCode: "Usage Summary", Display: "使用概要", FromName: "section", LanguageType: "zh-CN", CN: "使用概要"},
		{LCode: "Usage Rate", Display: "使用率", FromName: "form", LanguageType: "zh-CN", CN: "使用率"},
		{LCode: "Unused", Display: "未使用", FromName: "form", LanguageType: "zh-CN", CN: "未使用"},
		{LCode: "e.g. OpenAI Official", Display: "例如 OpenAI 官方", FromName: "hint", LanguageType: "zh-CN", CN: "例如 OpenAI 官方"},
		{LCode: "e.g. Production API Key", Display: "例如 生产环境 API 密钥", FromName: "hint", LanguageType: "zh-CN", CN: "例如 生产环境 API 密钥"},
	}

	for _, resource := range resources {
		var existing LanguageResource
		if err := DB.Where("language_type = ? AND l_code = ? AND from_name = ?", resource.LanguageType, resource.LCode, resource.FromName).First(&existing).Error; err != nil {
			DB.Create(&resource)
		}
	}
}
