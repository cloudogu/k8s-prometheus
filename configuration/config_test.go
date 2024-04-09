package configuration

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadConfigFromEnv(t *testing.T) {
	t.Run("should read config from env", func(t *testing.T) {
		err := os.Setenv("WEB_CONFIG_FILE", "/web-config.file")
		require.NoError(t, err)
		err = os.Setenv("API_KEY", "myApiKey")
		require.NoError(t, err)
		err = os.Setenv("PROMETHEUS_URL", "http://prom.url")
		require.NoError(t, err)
		err = os.Setenv("LOG_LEVEL", "DEBUG")
		require.NoError(t, err)

		conf, err := ReadConfigFromEnv()

		require.NoError(t, err)
		assert.Equal(t, "/web-config.file", conf.WebConfigFile)
		assert.Equal(t, "myApiKey", conf.ApiKey)
		assert.Equal(t, "http://prom.url", conf.PrometheusUrl)
		assert.Equal(t, "DEBUG", conf.LogLevel)
	})

	t.Run("should fail for missing web-configfile", func(t *testing.T) {
		err := os.Unsetenv("WEB_CONFIG_FILE")
		require.NoError(t, err)
		err = os.Unsetenv("API_KEY")
		require.NoError(t, err)
		err = os.Unsetenv("PROMETHEUS_URL")
		require.NoError(t, err)
		_, err = ReadConfigFromEnv()

		require.Error(t, err)
		assert.ErrorContains(t, err, "environment variable WEB_CONFIG_FILE is not set")
	})

	t.Run("should fail for missing api-key", func(t *testing.T) {
		err := os.Setenv("WEB_CONFIG_FILE", "/web-config.file")
		require.NoError(t, err)
		err = os.Unsetenv("API_KEY")
		require.NoError(t, err)
		err = os.Unsetenv("PROMETHEUS_URL")
		require.NoError(t, err)
		_, err = ReadConfigFromEnv()

		require.Error(t, err)
		assert.ErrorContains(t, err, "environment variable API_KEY is not set")
	})

	t.Run("should fail for missing prometheus-url", func(t *testing.T) {
		err := os.Setenv("WEB_CONFIG_FILE", "/web-config.file")
		require.NoError(t, err)
		err = os.Setenv("API_KEY", "myApiKey")
		require.NoError(t, err)
		err = os.Unsetenv("PROMETHEUS_URL")
		require.NoError(t, err)
		_, err = ReadConfigFromEnv()

		require.Error(t, err)
		assert.ErrorContains(t, err, "environment variable PROMETHEUS_URL is not set")
	})

	t.Run("should read config from env and set default log level if missing env", func(t *testing.T) {
		err := os.Setenv("WEB_CONFIG_FILE", "/web-config.file")
		require.NoError(t, err)
		err = os.Setenv("API_KEY", "myApiKey")
		require.NoError(t, err)
		err = os.Setenv("PROMETHEUS_URL", "http://prom.url")
		require.NoError(t, err)
		err = os.Unsetenv("LOG_LEVEL")
		require.NoError(t, err)

		conf, err := ReadConfigFromEnv()

		require.NoError(t, err)
		assert.Equal(t, "/web-config.file", conf.WebConfigFile)
		assert.Equal(t, "myApiKey", conf.ApiKey)
		assert.Equal(t, "http://prom.url", conf.PrometheusUrl)
		assert.Equal(t, "INFO", conf.LogLevel)
	})
}
