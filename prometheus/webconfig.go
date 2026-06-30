package prometheus

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// WebConfig abstracts prometheus auth's web.config.yaml configuration file which contains user accounts.
type WebConfig struct {
	// BasicAuthUsers maps a username key to a hashed password.
	BasicAuthUsers map[string]string `yaml:"basic_auth_users"`
}

// ExistsUser returns true if a given user already exists, otherwise false.
func (wc *WebConfig) ExistsUser(user string) bool {
	_, ok := wc.BasicAuthUsers[user]
	return ok
}

// WebConfigFileReaderWriter can both read from and write to the provided web.config.yaml file.
type WebConfigFileReaderWriter struct {
	configFile string
}

// NewWebConfigFileReaderWriter creates a reader/writer for the given config file.
func NewWebConfigFileReaderWriter(configFile string) *WebConfigFileReaderWriter {
	return &WebConfigFileReaderWriter{configFile: configFile}
}

// ReadWebConfig reads the given config file and returns a respective web config object.
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

// WriteWebConfig takes a web config and writes it to the given config file.
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
