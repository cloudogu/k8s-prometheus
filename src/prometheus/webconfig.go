package prometheus

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type WebConfig struct {
	BasicAuthUsers map[string]string `yaml:"basic_auth_users"`
}

type WebConfigReaderWriter struct {
	configFile string
}

func NewWebConfigReaderWriter(configFile string) *WebConfigReaderWriter {
	return &WebConfigReaderWriter{configFile: configFile}
}

func (rw *WebConfigReaderWriter) ReadWebConfig() (WebConfig, error) {
	var webConfig WebConfig

	webConfigFile, err := os.ReadFile(rw.configFile)
	if err != nil {
		return webConfig, fmt.Errorf("error reading web-config-file %s: %w", rw.configFile, err)
	}

	if err := yaml.Unmarshal(webConfigFile, &webConfig); err != nil {
		return webConfig, fmt.Errorf("error parsing web-config: %w", err)
	}

	return webConfig, nil
}

func (rw *WebConfigReaderWriter) WriteWebConfig(webConfig WebConfig) error {
	yamlData, err := yaml.Marshal(&webConfig)
	if err != nil {
		return fmt.Errorf("error marshalling web-config: %w", err)
	}

	if err := os.WriteFile(rw.configFile, yamlData, 0644); err != nil {
		return fmt.Errorf("error writing web-config-file %s: %w", rw.configFile, err)
	}

	return nil
}
