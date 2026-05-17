package helper

import billratio "github.com/quantumclaw/quantumclaw/relay/billing/ratio"

type PriceResult struct {
	Ratio            float64
	Price            float64
	CompletionRatio  float64
}

func GetModelPrice(modelName string, channelType int) *PriceResult {
	modelRatio := billratio.GetModelRatio(modelName, channelType)
	groupRatio := billratio.GetGroupRatio("default")
	completionRatio := billratio.GetCompletionRatio(modelName, channelType)
	return &PriceResult{
		Ratio:           modelRatio * groupRatio,
		Price:           0,
		CompletionRatio: completionRatio,
	}
}
