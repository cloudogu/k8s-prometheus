package main

import (
	"log/slog"
	"os"

	"github.com/cloudogu/k8s-prometheus/auth/internal/storageresize"
)

func main() {
	err := storageresize.ResizeStorage()
	if err != nil {
		slog.Error("Error while resize storage", "error", err)
		os.Exit(1)
	}
}
