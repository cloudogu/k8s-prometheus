package prometheus

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type WebConfig struct {
	BasicAuthUsers map[string]string `yaml:"basic_auth_users"`
}

type WebConfigFileReaderWriter struct {
	configFile string
}

func NewWebConfigFileReaderWriter(configFile string) *WebConfigFileReaderWriter {
	return &WebConfigFileReaderWriter{configFile: configFile}
}

func (rw *WebConfigFileReaderWriter) ReadWebConfig() (*WebConfig, error) {
	var webConfig = WebConfig{BasicAuthUsers: make(map[string]string)}

	webConfigFile, err := os.ReadFile(rw.configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// use empty web-config
			return &webConfig, nil
		}

		return nil, fmt.Errorf("error reading web-config-file %s: %w", rw.configFile, err)
	}

	if err := yaml.Unmarshal(webConfigFile, &webConfig); err != nil {
		return nil, fmt.Errorf("error parsing web-config: %w", err)
	}

	return &webConfig, nil
}

func (rw *WebConfigFileReaderWriter) WriteWebConfig(webConfig *WebConfig) error {
	yamlData, err := yaml.Marshal(webConfig)
	if err != nil {
		return fmt.Errorf("error marshalling web-config: %w", err)
	}

	if err := os.WriteFile(rw.configFile, yamlData, 0644); err != nil {
		return fmt.Errorf("error writing web-config-file %s: %w", rw.configFile, err)
	}

	return nil
}
