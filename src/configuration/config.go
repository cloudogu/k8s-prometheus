package configuration

import (
	"fmt"
	"os"
)

const configFileEnv = "WEB_CONFIG_FILE"
const apiKeyEnv = "API_KEY"
const prometheusUrlEnv = "PROMETHEUS_URL"

type Configuration struct {
	WebConfigFile string
	ApiKey        string
	PrometheusUrl string
}

func ReadConfigFromEnv() (Configuration, error) {
	conf := Configuration{}

	conf.WebConfigFile = os.Getenv(configFileEnv)
	if conf.WebConfigFile == "" {
		return conf, fmt.Errorf("environement variable %s is not set", configFileEnv)
	}

	conf.ApiKey = os.Getenv(apiKeyEnv)
	if conf.ApiKey == "" {
		return conf, fmt.Errorf("environement variable %s is not set", apiKeyEnv)
	}

	conf.PrometheusUrl = os.Getenv(prometheusUrlEnv)
	if conf.PrometheusUrl == "" {
		return conf, fmt.Errorf("environement variable %s is not set", prometheusUrlEnv)
	}

	return conf, nil
}
