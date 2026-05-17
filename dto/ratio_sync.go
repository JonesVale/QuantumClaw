package dto

type RatioSyncRequest struct {
	ModelRatios      map[string]float64 `json:"model_ratios"`
	GroupRatios      map[string]float64 `json:"group_ratios"`
	CompletionRatios map[string]float64 `json:"completion_ratios"`
}
