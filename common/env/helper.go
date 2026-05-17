package env

import (
	"os"
	"strconv"
	"strings"
)

func Bool(env string, defaultValue bool) bool {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	return os.Getenv(env) == "true"
}

func Int(env string, defaultValue int) int {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	num, err := strconv.Atoi(os.Getenv(env))
	if err != nil {
		return defaultValue
	}
	return num
}

func Float64(env string, defaultValue float64) float64 {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	num, err := strconv.ParseFloat(os.Getenv(env), 64)
	if err != nil {
		return defaultValue
	}
	return num
}

func String(env string, defaultValue string) string {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	return os.Getenv(env)
}

// StringSlice reads an environment variable and splits it by comma into a string slice.
// If the environment variable is empty or not set, returns defaultValue.
func StringSlice(env string, defaultValue []string) []string {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	val := os.Getenv(env)
	// Split by comma and trim spaces
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
