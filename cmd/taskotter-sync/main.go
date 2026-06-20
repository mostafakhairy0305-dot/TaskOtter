package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/app"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	cfg, err := config.LoadFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "::error::%v\n", err)
		return 1
	}

	result, err := app.Run(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "::error::%v\n", err)
		return 1
	}

	if err := app.WriteActionOutputs(cfg, result); err != nil {
		fmt.Fprintf(os.Stderr, "::error::write outputs: %v\n", err)
		return 1
	}

	if result.Changed {
		fmt.Println("TaskOtter produced changes.")
		if cfg.FailOnChanges {
			app.ReportSyncRequired(result)
			return 1
		}
		return 0
	}

	fmt.Println("TaskOtter completed with no changes.")
	if cfg.FailOnChanges {
		app.ReportSyncUpToDate(result)
	}
	return 0
}
