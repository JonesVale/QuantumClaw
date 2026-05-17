package common

import "fmt"

func GetQuotaText(quota int64) string {
	if quota <= 0 {
		return "0"
	}
	return Int64ToString(quota)
}

func Int64ToString(i int64) string {
	return fmt.Sprintf("%d", i)
}
