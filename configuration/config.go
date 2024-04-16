package configuration

import (
	"fmt"
	"os"
)

const configFileEnv = "WEB_CONFIG_FILE"
const apiKeyEnv = "API_KEY"
const prometheusUrlEnv = "PROMETHEUS_URL"
const logLevelEnv = "LOG_LEVEL"

const errorFormat = "environment variable %s is not set"

type Configuration struct {
	WebConfigFile string
	ApiKey        string
	PrometheusUrl string
	LogLevel      string
}

func ReadConfigFromEnv() (Configuration, error) {
	conf := Configuration{}

	conf.LogLevel = os.Getenv(logLevelEnv)
	if conf.LogLevel == "" {
		conf.LogLevel = "INFO"
	}

	conf.WebConfigFile = os.Getenv(configFileEnv)
	if conf.WebConfigFile == "" {
		return conf, fmt.Errorf(errorFormat, configFileEnv)
	}

	conf.ApiKey = os.Getenv(apiKeyEnv)
	if conf.ApiKey == "" {
		return conf, fmt.Errorf(errorFormat, apiKeyEnv)
	}

	conf.PrometheusUrl = os.Getenv(prometheusUrlEnv)
	if conf.PrometheusUrl == "" {
		return conf, fmt.Errorf(errorFormat, prometheusUrlEnv)
	}

	return conf, nil
}
