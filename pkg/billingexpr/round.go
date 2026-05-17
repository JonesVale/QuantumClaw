package billingexpr

import "math"

func RoundUp(value float64, decimals int) float64 {
	if decimals < 0 {
		decimals = 0
	}
	shift := math.Pow(10, float64(decimals))
	return math.Ceil(value*shift) / shift
}

func RoundDown(value float64, decimals int) float64 {
	if decimals < 0 {
		decimals = 0
	}
	shift := math.Pow(10, float64(decimals))
	return math.Floor(value*shift) / shift
}

func RoundNearest(value float64, decimals int) float64 {
	if decimals < 0 {
		decimals = 0
	}
	shift := math.Pow(10, float64(decimals))
	return math.Round(value*shift) / shift
}