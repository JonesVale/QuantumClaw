package controller

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/setting/performance_setting"
	"github.com/gin-gonic/gin"
)

// ==================== 性能监控 ====================

type PerformanceStats struct {
	Memory struct {
		Alloc      uint64 `json:"alloc"`
		TotalAlloc uint64 `json:"total_alloc"`
		Sys        uint64 `json:"sys"`
		NumGC      uint32 `json:"num_gc"`
		GoRoutines int    `json:"goroutines"`
	} `json:"memory"`
	Runtime struct {
		NumCPU    int `json:"num_cpu"`
		NumCgo    int `json:"num_cgo"`
		NumThread int `json:"num_thread"`
	} `json:"runtime"`
	Uptime struct {
		StartTime    int64  `json:"start_time"`
		UptimeSeconds int64 `json:"uptime_seconds"`
	} `json:"uptime"`
	Cache struct {
		ChannelAffinityStats any `json:"channel_affinity_stats"`
	} `json:"cache"`
}

// ==================== 获取性能统计 ====================

func GetPerformanceStats(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := PerformanceStats{}
	stats.Memory.Alloc = m.Alloc
	stats.Memory.TotalAlloc = m.TotalAlloc
	stats.Memory.Sys = m.Sys
	stats.Memory.NumGC = m.NumGC
	stats.Memory.GoRoutines = runtime.NumGoroutine()
	stats.Runtime.NumCPU = runtime.NumCPU()
	stats.Uptime.StartTime = startTimeUnix
	stats.Uptime.UptimeSeconds = time.Now().Unix() - startTimeUnix

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// ==================== Prometheus 格式指标 ====================

func GetPrometheusMetrics(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	setting := performance_setting.GetPerformanceSetting()
	if !setting.EnablePrometheusMetrics {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prometheus metrics not enabled"})
		return
	}

	var sb strings.Builder
	sb.WriteString("# HELP quantumclaw_memory_alloc_bytes Current bytes allocated by the Go process.\n")
	sb.WriteString("# TYPE quantumclaw_memory_alloc_bytes gauge\n")
	sb.WriteString(fmt.Sprintf("quantumclaw_memory_alloc_bytes %d\n", m.Alloc))

	sb.WriteString("# HELP quantumclaw_memory_total_alloc_bytes Total bytes allocated by the Go process (cumulative).\n")
	sb.WriteString("# TYPE quantumclaw_memory_total_alloc_bytes gauge\n")
	sb.WriteString(fmt.Sprintf("quantumclaw_memory_total_alloc_bytes %d\n", m.TotalAlloc))

	sb.WriteString("# HELP quantumclaw_memory_sys_bytes Total bytes of memory obtained from the OS.\n")
	sb.WriteString("# TYPE quantumclaw_memory_sys_bytes gauge\n")
	sb.WriteString(fmt.Sprintf("quantumclaw_memory_sys_bytes %d\n", m.Sys))

	sb.WriteString("# HELP quantumclaw_goroutines Number of goroutines.\n")
	sb.WriteString("# TYPE quantumclaw_goroutines gauge\n")
	sb.WriteString(fmt.Sprintf("quantumclaw_goroutines %d\n", runtime.NumGoroutine()))

	sb.WriteString("# HELP quantumclaw_gc_total Total number of GC cycles.\n")
	sb.WriteString("# TYPE quantumclaw_gc_total counter\n")
	sb.WriteString(fmt.Sprintf("quantumclaw_gc_total %d\n", m.NumGC))

	sb.WriteString("# HELP quantumclaw_uptime_seconds Process uptime in seconds.\n")
	sb.WriteString("# TYPE quantumclaw_uptime_seconds counter\n")
	sb.WriteString(fmt.Sprintf("quantumclaw_uptime_seconds %d\n", time.Now().Unix()-startTimeUnix))

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, sb.String())
}

var startTimeUnix = time.Now().Unix()
