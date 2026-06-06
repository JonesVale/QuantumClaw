package model

import (
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ── 多语言翻译 ──

// Translation 多语言翻译键值对
type Translation struct {
	Id        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	LangKey   string    `json:"lang_key" gorm:"type:varchar(100);uniqueIndex:idx_translation;not null"`  // 翻译键名
	LangCode  string    `json:"lang_code" gorm:"type:varchar(10);uniqueIndex:idx_translation;not null"` // 语言代码 (en, zh-CN, ja...)
	Value     string    `json:"value" gorm:"type:text;not null"`                                          // 翻译值
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Translation) TableName() string {
	return "translations"
}

func InitTranslationTables() {
	if err := DB.AutoMigrate(&Translation{}); err != nil {
		logger.SysError("InitTranslationTables AutoMigrate failed: " + err.Error())
		return
	}
	logger.SysLog("translation tables initialized")
}

// ── CRUD ──

// GetTranslationsByCode 获取指定语言的所有翻译
func GetTranslationsByCode(langCode string) ([]Translation, error) {
	var list []Translation
	err := DB.Where("lang_code = ?", langCode).Find(&list).Error
	return list, err
}

// UpsertTranslation 创建或更新翻译
func UpsertTranslation(langKey, langCode, value string) error {
	var existing Translation
	result := DB.Where("lang_key = ? AND lang_code = ?", langKey, langCode).First(&existing)
	if result.Error != nil {
		// Insert
		return DB.Create(&Translation{
			LangKey:  langKey,
			LangCode: langCode,
			Value:    value,
		}).Error
	}
	// Update
	return DB.Model(&existing).Update("value", value).Error
}

// BatchUpsertTranslations 批量导入翻译
func BatchUpsertTranslations(translations []Translation) (int, error) {
	count := 0
	for _, t := range translations {
		if err := UpsertTranslation(t.LangKey, t.LangCode, t.Value); err != nil {
			logger.SysError("BatchUpsertTranslations failed: " + err.Error())
			continue
		}
		count++
	}
	return count, nil
}

// DeleteTranslation 删除翻译
func DeleteTranslation(id int) error {
	return DB.Delete(&Translation{}, id).Error
}

// SearchTranslations 搜索翻译
func SearchTranslations(query, langCode string, offset, limit int) ([]Translation, int64, error) {
	var list []Translation
	var total int64
	db := DB.Model(&Translation{})
	if langCode != "" {
		db = db.Where("lang_code = ?", langCode)
	}
	if query != "" {
		db = db.Where("lang_key LIKE ? OR value LIKE ?", "%"+query+"%", "%"+query+"%")
	}
	db.Count(&total)
	err := db.Order("lang_code, lang_key").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}
