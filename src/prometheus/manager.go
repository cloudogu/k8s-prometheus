package prometheus

import (
	"fmt"
	"strings"
)

type Manager struct {
	rw *WebConfigReaderWriter
}

func NewManager(configFile string) *Manager {
	return &Manager{rw: NewWebConfigReaderWriter(configFile)}
}

func (m *Manager) CreateServiceAccount(consumer string, params []string) (credentials map[string]string, err error) {
	config, err := m.rw.ReadWebConfig()
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
	config, err := m.rw.ReadWebConfig()
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
