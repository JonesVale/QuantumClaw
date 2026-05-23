package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// SeedQuantumModels 种子数据：写入常用量子模型到 model_metadata
// GET /api/models/seed-quantum
func SeedQuantumModels(c *gin.Context) {
	languages := []string{"中文简体", "中文繁体", "English", "Français", "日本語", "Русский", "Tiếng Việt"}

	type quantumModel struct {
		Name     string
		Provider string
		Series   string
	}

	quantumModels := []quantumModel{
		{Name: "ionq/harmony", Provider: "IonQ", Series: "Harmony"},
		{Name: "ionq/aria-1", Provider: "IonQ", Series: "Aria"},
		{Name: "ionq/forte-1", Provider: "IonQ", Series: "Forte"},
		{Name: "ibm/sherbrooke", Provider: "IBM Quantum", Series: "Sherbrooke"},
		{Name: "ibm/kyiv", Provider: "IBM Quantum", Series: "Kyiv"},
		{Name: "ibm/fez", Provider: "IBM Quantum", Series: "Fez"},
		{Name: "rigetti/aspen-m-3", Provider: "Rigetti", Series: "Aspen"},
		{Name: "rigetti/ankaa-2", Provider: "Rigetti", Series: "Ankaa"},
		{Name: "aws/braket-sv1", Provider: "AWS Braket", Series: "SV1"},
		{Name: "aws/braket-tn1", Provider: "AWS Braket", Series: "TN1"},
		{Name: "aws/braket-dm1", Provider: "AWS Braket", Series: "DM1"},
		{Name: "azure/quantum-sim", Provider: "Azure Quantum", Series: "Simulator"},
		{Name: "azure/quantum-ionq", Provider: "Azure Quantum", Series: "IonQ"},
		{Name: "azure/quantum-rigetti", Provider: "Azure Quantum", Series: "Rigetti"},
	}

	now := time.Now().Unix()
	count := 0

	for _, qm := range quantumModels {
		// 检查是否已存在
		var existing int64
		model.DB.Model(&model.ModelMetadata{}).
			Where("model_name = ? AND languages_type = 'English'", qm.Name).
			Count(&existing)
		if existing > 0 {
			continue
		}

		enDesc := qm.Name + " is a quantum computing backend from " + qm.Provider + " " + qm.Series + " series."
		cnDesc := qm.Name + " 是 " + qm.Provider + " " + qm.Series + " 系列的量子计算后端。"

		for _, lang := range languages {
			desc := enDesc
			switch lang {
			case "中文简体", "中文繁体":
				desc = cnDesc
			}
			m := &model.ModelMetadata{
				ModelName:       qm.Name,
				LanguagesType:   lang,
				DisplayName:     qm.Name,
				Description:     desc,
				UseCase:         "quantum",
				ContextWindow:   1000000,
				InputModalities: `["Text"]`,
				Series:          qm.Series,
				Provider:        qm.Provider,
				CreatedTime:     now,
				UpdatedTime:     now,
			}
			if err := model.DB.Create(m).Error; err == nil {
				count++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "quantum models seeded",
		"inserted": count,
	})
}
