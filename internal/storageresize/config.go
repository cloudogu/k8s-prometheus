package storageresize

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	errorFormat     = "environment variable %s is required but not set"
	logLevelEnv     = "LOG_LEVEL"
	namespaceEnv    = "NAMESPACE"
	promOpNameEnv   = "PROMETHUES_OPERATOR_NAME"
	storageSizeEnv  = "STORAGE_SIZE"
	defaultLogLevel = "INFO"
)

type config struct {
	logLevel         string
	promOperatorName string
	storageSize      string
	namespace        string
}

func readConfig() (*config, error) {
	cfg := &config{}
	cfg.logLevel = os.Getenv(logLevelEnv)
	if cfg.logLevel == "" {
		cfg.logLevel = defaultLogLevel
	}

	cfg.namespace = os.Getenv(namespaceEnv)
	if cfg.namespace == "" {
		return nil, fmt.Errorf(errorFormat, namespaceEnv)
	}

	cfg.promOperatorName = os.Getenv(promOpNameEnv)
	if cfg.promOperatorName == "" {
		return nil, fmt.Errorf(errorFormat, promOpNameEnv)
	}

	cfg.storageSize = os.Getenv(storageSizeEnv)
	if cfg.storageSize == "" {
		return nil, fmt.Errorf(errorFormat, storageSizeEnv)
	}
	_, err := resource.ParseQuantity(cfg.storageSize)
	if err != nil {
		return nil, fmt.Errorf(
			"parse the storage quantity '%s' from the %s environment variable: %w",
			cfg.storageSize,
			storageSizeEnv,
			err,
		)
	}

	return cfg, nil
}
