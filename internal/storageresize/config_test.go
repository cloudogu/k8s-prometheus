package storageresize

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConfig(t *testing.T) {
	t.Run("should read config", func(t *testing.T) {
		err := os.Setenv(logLevelEnv, "ERROR")
		require.NoError(t, err)
		err = os.Setenv(promOpNameEnv, "aOpName")
		require.NoError(t, err)
		err = os.Setenv(storageSizeEnv, "10Gi")

		cfg, err := readConfig()

		assert.NoError(t, err)
		assert.Equal(t, &config{
			logLevel:         "ERROR",
			promOperatorName: "aOpName",
			storageSize:      "10Gi",
		}, cfg)
	})

	t.Run("should use the default log level if none is specified", func(t *testing.T) {
		err := os.Unsetenv(logLevelEnv)
		require.NoError(t, err)
		err = os.Setenv(promOpNameEnv, "aOpName")
		require.NoError(t, err)
		err = os.Setenv(storageSizeEnv, "10Gi")
		require.NoError(t, err)

		cfg, err := readConfig()

		assert.NoError(t, err)
		assert.Equal(t, defaultLogLevel, cfg.logLevel)
	})

	t.Run("should return an error if no prometheus operator name is specified", func(t *testing.T) {
		err := os.Setenv(logLevelEnv, "ERROR")
		require.NoError(t, err)
		err = os.Unsetenv(promOpNameEnv)
		require.NoError(t, err)
		err = os.Setenv(storageSizeEnv, "10Gi")
		require.NoError(t, err)

		_, err = readConfig()

		assert.Error(t, err)
		assert.ErrorContains(t, err, promOpNameEnv)
	})

	t.Run("should return an error if no storage size is specified", func(t *testing.T) {
		err := os.Unsetenv(logLevelEnv)
		require.NoError(t, err)
		err = os.Setenv(promOpNameEnv, "aOpName")
		require.NoError(t, err)
		err = os.Unsetenv(storageSizeEnv)
		require.NoError(t, err)

		_, err = readConfig()

		assert.Error(t, err)
		assert.ErrorContains(t, err, storageSizeEnv)
	})

	t.Run("should return an error if the storage size has the wrong format", func(t *testing.T) {
		err := os.Setenv(logLevelEnv, "ERROR")
		require.NoError(t, err)
		err = os.Setenv(promOpNameEnv, "aOpName")
		require.NoError(t, err)
		err = os.Setenv(storageSizeEnv, "10,45")

		_, err = readConfig()

		assert.Error(t, err)
		assert.ErrorContains(t, err, "storage quantity")
		assert.ErrorContains(t, err, storageSizeEnv)
	})
}
