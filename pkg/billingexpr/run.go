package billingexpr

import (
	"fmt"
	"time"

	"github.com/expr-lang/expr/vm"
)

type RuntimeVars struct {
	P     float64
	C     float64
	Len   float64
	Cr    float64
	Cc    float64
	Cc1h  float64
	Img   float64
	ImgO  float64
	Ai    float64
	Ao    float64
}

func Run(prog *vm.Program, vars RuntimeVars) (float64, error) {
	env := map[string]interface{}{
		"p":     vars.P,
		"c":     vars.C,
		"len":   vars.Len,
		"cr":    vars.Cr,
		"cc":    vars.Cc,
		"cc1h":  vars.Cc1h,
		"img":   vars.Img,
		"img_o": vars.ImgO,
		"ai":    vars.Ai,
		"ao":    vars.Ao,
		"tier": func(tierName string, defaultRate float64) float64 {
			return defaultRate
		},
		"header": func(name string) string {
			return ""
		},
		"param": func(name string) interface{} {
			return nil
		},
		"has": func(obj interface{}, key string) bool {
			return false
		},
		"hour": func(layout string) int {
			return time.Now().Hour()
		},
		"minute": func(layout string) int {
			return time.Now().Minute()
		},
		"weekday": func(layout string) int {
			return int(time.Now().Weekday())
		},
		"month": func(layout string) int {
			return int(time.Now().Month())
		},
		"day": func(layout string) int {
			return time.Now().Day()
		},
		"max":    func(a, b float64) float64 { return max(a, b) },
		"min":    func(a, b float64) float64 { return min(a, b) },
		"abs":    func(x float64) float64 { return abs(x) },
		"ceil":   func(x float64) float64 { return ceil(x) },
		"floor":  func(x float64) float64 { return floor(x) },
		"round":     func(value float64, decimals int) float64 { return RoundNearest(value, decimals) },
		"roundUp":   func(value float64, decimals int) float64 { return RoundUp(value, decimals) },
		"roundDown": func(value float64, decimals int) float64 { return RoundDown(value, decimals) },
	}

	result, err := vm.Run(prog, env)
	if err != nil {
		return 0, fmt.Errorf("expr run error: %w", err)
	}

	v, ok := result.(float64)
	if !ok {
		return 0, fmt.Errorf("expr returned non-float64 result")
	}

	return v, nil
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func ceil(x float64) float64 {
	return float64(int(x)) + 1
}

func floor(x float64) float64 {
	return float64(int(x))
}