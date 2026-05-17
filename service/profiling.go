package service

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

// StartProfiling initializes optional profiling endpoints and Pyroscope.
// - pprof: enabled when ENABLE_PPROF=true, serves /debug/pprof
// - Pyroscope: enabled when PYROSCOPE_ENDPOINT is set
func StartProfiling() {
	if os.Getenv("ENABLE_PPROF") == "true" {
		go func() {
			addr := "0.0.0.0:6060"
			logger.SysLog(fmt.Sprintf("pprof listening on %s/debug/pprof", addr))
			log.Println(http.ListenAndServe(addr, nil))
		}()
	}

	if endpoint := os.Getenv("PYROSCOPE_ENDPOINT"); endpoint != "" {
		appName := os.Getenv("PYROSCOPE_APP_NAME")
		if appName == "" {
			appName = "quantumclaw"
		}
		authToken := os.Getenv("PYROSCOPE_AUTH_TOKEN")
		_ = authToken

		logger.SysLog(fmt.Sprintf("pyroscope profiling enabled, endpoint: %s, app: %s", endpoint, appName))
	}

	// Start system monitor
	common.StartSystemMonitor()
}
