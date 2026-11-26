package storageresize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	t.Run("should read config", func(t *testing.T) {
		t.Setenv(logLevelEnv, "ERROR")
		t.Setenv(namespaceEnv, "ns1")
		t.Setenv(promOpNameEnv, "aOpName")
		t.Setenv(storageSizeEnv, "10Gi")

		cfg, err := ReadConfigFromEnv()

		assert.NoError(t, err)
		assert.Equal(t, &Config{
			LogLevel:         "ERROR",
			promOperatorName: "aOpName",
			storageSize:      "10Gi",
			namespace:        "ns1",
		}, cfg)
	})

	t.Run("should use the default log level if none is specified", func(t *testing.T) {
		t.Setenv(namespaceEnv, "ns1")
		t.Setenv(promOpNameEnv, "aOpName")
		t.Setenv(storageSizeEnv, "10Gi")

		cfg, err := ReadConfigFromEnv()

		assert.NoError(t, err)
		assert.Equal(t, defaultLogLevel, cfg.LogLevel)
	})

	t.Run("should return an error if no namespace is specified", func(t *testing.T) {
		t.Setenv(logLevelEnv, "ERROR")
		t.Setenv(promOpNameEnv, "aOpName")
		t.Setenv(storageSizeEnv, "10Gi")

		_, err := ReadConfigFromEnv()

		assert.ErrorContains(t, err, namespaceEnv)
	})

	t.Run("should return an error if no prometheus operator name is specified", func(t *testing.T) {
		t.Setenv(logLevelEnv, "ERROR")
		t.Setenv(namespaceEnv, "ns1")
		t.Setenv(storageSizeEnv, "10Gi")

		_, err := ReadConfigFromEnv()

		assert.ErrorContains(t, err, promOpNameEnv)
	})

	t.Run("should return an error if no storage size is specified", func(t *testing.T) {
		t.Setenv(logLevelEnv, "ERROR")
		t.Setenv(namespaceEnv, "ns1")
		t.Setenv(promOpNameEnv, "aOpName")

		_, err := ReadConfigFromEnv()

		assert.ErrorContains(t, err, storageSizeEnv)
	})

	t.Run("should return an error if the storage size has the wrong format", func(t *testing.T) {
		t.Setenv(logLevelEnv, "ERROR")
		t.Setenv(namespaceEnv, "ns1")
		t.Setenv(promOpNameEnv, "aOpName")
		t.Setenv(storageSizeEnv, "10,45")

		_, err := ReadConfigFromEnv()

		assert.ErrorContains(t, err, "storage quantity")
		assert.ErrorContains(t, err, storageSizeEnv)
	})
}
