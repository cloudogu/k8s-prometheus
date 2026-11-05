package storageresize

import (
	"fmt"

	"github.com/cloudogu/k8s-prometheus/auth/internal/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ResizeStorage() error {
	cfg, err := readConfig()
	if err != nil {
		return fmt.Errorf("read config %v", err)
	}

	logging.ConfigureLogger(cfg.logLevel)

	// patchPVCs

	return nil
}

func patchPVCs(client client.Client, size string) error {
	return nil
}
