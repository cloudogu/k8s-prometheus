package main

import (
	"log/slog"
	"os"

	"github.com/cloudogu/k8s-prometheus/tools/internal/logging"
	"github.com/cloudogu/k8s-prometheus/tools/internal/storageresize"
)

func main() {
	config, err := storageresize.ReadConfigFromEnv()
	if err != nil {
		slog.Error("read configuration from env", "error", err)
		os.Exit(1)
	}

	logging.ConfigureLogger(config.LogLevel)

	k8sClient, err := storageresize.CreateK8sClient()
	if err != nil {
		slog.Error("create k8s client", "error", err)
		os.Exit(1)
	}

	resizer := storageresize.NewResizeSteps(config, k8sClient)
	err = storageresize.ResizeStorage(resizer)
	if err != nil {
		slog.Error("resize storage", "error", err)
		os.Exit(1)
	}
}
