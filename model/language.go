package model

import (
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

	var count int64
	DB.Model(&LanguageType{}).Count(&count)
	if count > 0 {
		return
	}

	langs := []LanguageType{
		{LanguagesType: "中文简体", AddTime: "2026-05-20"},
		{LanguagesType: "English", AddTime: "2026-05-20"},
	}
	for _, l := range langs {
		DB.Create(&l)
	}
	logger.SysLog("language types seeded: 中文简体 + English")
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
