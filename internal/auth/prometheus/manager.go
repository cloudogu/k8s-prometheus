package prometheus

import (
	"fmt"
	"strings"
)

type WebConfigReaderWriter interface {
	ReadWebConfig() (*WebConfig, error)
	WriteWebConfig(webConfig *WebConfig) error
}

type Manager struct {
	rw         WebConfigReaderWriter
	webConfig  *WebConfig
	webPresets *WebConfig
}

func NewManager(configFile string, webPresets *WebConfig) *Manager {
	return &Manager{
		rw:         NewWebConfigFileReaderWriter(configFile),
		webPresets: webPresets,
	}
}

func (m *Manager) getWebConfig() (*WebConfig, error) {
	if m.webConfig != nil {
		return m.webConfig, nil
	}

	config, err := m.rw.ReadWebConfig()
	if err != nil {
		return nil, err
	}
	m.webConfig = config

	return m.webConfig, nil
}

func (m *Manager) CreateServiceAccount(consumer string, params []string) (credentials map[string]string, err error) {
	config, err := m.getWebConfig()
	if err != nil {
		return nil, err
	}

	username, password, err := createNewAccount(consumer)
	if err != nil {
		return nil, err
	}
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	config.BasicAuthUsers[username] = hashedPassword

	if err := m.rw.WriteWebConfig(config); err != nil {
		return nil, err
	}

	return map[string]string{
		"username": username,
		"password": password,
	}, nil
}

func (m *Manager) DeleteServiceAccount(consumer string) error {
	config, err := m.getWebConfig()
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("%s-", consumer)
	for username := range config.BasicAuthUsers {
		if strings.HasPrefix(username, prefix) {
			delete(config.BasicAuthUsers, username)
			break
		}
	}

	return m.rw.WriteWebConfig(config)
}

func (m *Manager) ValidateAccount(username string, password string) error {
	config, err := m.getWebConfig()
	if err != nil {
		return err
	}

	for user, hashedPassword := range m.webPresets.BasicAuthUsers {
		if user == username {
			return compareHashAndPassword(hashedPassword, password)
		}
	}

	for user, hashedPassword := range config.BasicAuthUsers {
		if user == username {
			return compareHashAndPassword(hashedPassword, password)
		}
	}

	return fmt.Errorf("cloud not find user with name: %s", username)
}
