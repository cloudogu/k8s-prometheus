package prometheus

import (
	"fmt"
	"strings"
)

type Manager struct {
	rw        *WebConfigReaderWriter
	WebConfig *WebConfig
}

func NewManager(configFile string) *Manager {
	return &Manager{rw: NewWebConfigReaderWriter(configFile)}
}

func (m *Manager) getWebConfig() (*WebConfig, error) {
	if m.WebConfig != nil {
		return m.WebConfig, nil
	}

	config, err := m.rw.ReadWebConfig()
	if err != nil {
		return nil, err
	}
	m.WebConfig = config

	return m.WebConfig, nil
}

func (m *Manager) CreateServiceAccount(consumer string, params []string) (credentials map[string]string, err error) {
	config, err := m.getWebConfig()
	if err != nil {
		return nil, err
	}

	username, password := createNewAccount(consumer)
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

	for user, hashedPassword := range config.BasicAuthUsers {
		if user == username {
			return compareHashAndPassword(hashedPassword, password)
		}
	}

	return fmt.Errorf("cloud not find user with name: %s", username)
}
