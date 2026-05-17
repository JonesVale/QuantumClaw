package dto

type PricingRequest struct {
	Model string `json:"model"`
}

type PricingResponse struct {
	Model           string  `json:"model"`
	InputPrice      float64 `json:"input_price"`
	OutputPrice     float64 `json:"output_price"`
	InputCachePrice float64 `json:"input_cache_price,omitempty"`
	Currency        string  `json:"currency"`
}
