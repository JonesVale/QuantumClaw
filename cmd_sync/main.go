package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/service"
)

func main() {
	os.Chdir("/opt/quantumclaw")
	if err := service.SyncPopularApps(context.Background()); err != nil {
		fmt.Println("ERR:", err)
		os.Exit(1)
	}
	fmt.Println("DONE")
}
