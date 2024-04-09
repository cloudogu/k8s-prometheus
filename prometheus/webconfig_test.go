package prometheus

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func Test_NewWebConfigFileReaderWriter(t *testing.T) {
	t.Run("should creat new WebConfigFileReaderWriter", func(t *testing.T) {
		sut := NewWebConfigFileReaderWriter("/some/file.yaml")

		assert.Equal(t, sut.configFile, "/some/file.yaml")
	})
}

func Test_ReadWebConfig(t *testing.T) {
	t.Run("should read WebConfig", func(t *testing.T) {
		webConfig := &WebConfig{BasicAuthUsers: map[string]string{"user1": "password1", "user2": "password2", "user3": "password3", "user4": "password4"}}

		sut := &WebConfigFileReaderWriter{configFile: "testdata/web.config.yaml"}

		conf, err := sut.ReadWebConfig()

		require.NoError(t, err)
		assert.Equal(t, webConfig, conf)
	})

	t.Run("should get new WebConfig for non existing file", func(t *testing.T) {
		webConfig := &WebConfig{BasicAuthUsers: make(map[string]string)}

		sut := &WebConfigFileReaderWriter{configFile: "testdata/no-exists.yaml"}

		conf, err := sut.ReadWebConfig()

		require.NoError(t, err)
		assert.Equal(t, webConfig, conf)
	})

	t.Run("should fail to read WebConfig for wrongly fromatted file", func(t *testing.T) {
		sut := &WebConfigFileReaderWriter{configFile: "testdata/web.config.err.yaml"}

		_, err := sut.ReadWebConfig()

		require.Error(t, err)
		assert.ErrorContains(t, err, "error parsing web-config: yaml: line 2: mapping values are not allowed in this context")
	})
}

func Test_WriteWebConfig(t *testing.T) {
	t.Run("should write WebConfig", func(t *testing.T) {
		webConfig := &WebConfig{BasicAuthUsers: map[string]string{"user1": "password1", "user2": "password2", "user3": "password3", "user4": "password4"}}
		tmpFile, err := os.CreateTemp(t.TempDir(), "web.config.yaml")
		require.NoError(t, err)

		sut := &WebConfigFileReaderWriter{configFile: tmpFile.Name()}

		err = sut.WriteWebConfig(webConfig)

		require.NoError(t, err)

		//read config
		webConfigFile, err := os.ReadFile(tmpFile.Name())
		require.NoError(t, err)
		conf := WebConfig{BasicAuthUsers: make(map[string]string)}
		err = yaml.Unmarshal(webConfigFile, &conf)
		require.NoError(t, err)

		assert.Equal(t, webConfig, &conf)
	})

	t.Run("should write WebConfig", func(t *testing.T) {
		webConfig := &WebConfig{BasicAuthUsers: map[string]string{"user1": "password1", "user2": "password2", "user3": "password3", "user4": "password4"}}

		sut := &WebConfigFileReaderWriter{configFile: "not_exists/foo.yaml"}

		err := sut.WriteWebConfig(webConfig)

		require.Error(t, err)
		assert.ErrorContains(t, err, "error writing web-config-file not_exists/foo.yaml: open not_exists/foo.yaml: no such file or directory")
	})
}
