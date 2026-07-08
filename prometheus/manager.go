package prometheus

import (
	"fmt"
	"strings"
)

const (
	// CredentialNoChange means the update of the service account did not lead to a change of the underlying credentials.
	CredentialNoChange CredentialChangeType = iota
	// CredentialCreated means the service account created credentials.
	CredentialCreated
	// CredentialUpdated means the update of the service account did lead to a change of the underlying credentials.
	CredentialUpdated
)

// CredentialChangeType determines the type of change in actions where the credentials might change.
type CredentialChangeType int

// BehaviorParams may be used to by a consumer (via the service account request) to trigger actions towards the Service Account producer.
type BehaviorParams struct {
	// RotateServiceAccountNow indicates if a Service Account's credential should be rotated immediately.
	// This field must be ignored during the first creation of a Service Account (because there is nothing to rotate).
	// This field is optional and defaults to false.
	RotateServiceAccountNow bool `json:"rotateServiceAccountNow,omitempty"`
}

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

func (m *Manager) GetServiceAccount(consumer string) (user map[string]string, found bool, err error) {
	config, err := m.getWebConfig()
	if err != nil {
		return nil, false, err
	}

	_, ok := config.BasicAuthUsers[consumer]
	if !ok {
		return nil, false, nil
	}

	return map[string]string{
		"username": consumer,
	}, true, nil
}

func (m *Manager) CreateOrUpdateServiceAccount(consumer string, _ map[string]string, behaviorParams BehaviorParams) (credentials map[string]string, changed CredentialChangeType, err error) {
	config, err := m.getWebConfig()
	if err != nil {
		return nil, CredentialNoChange, err
	}
	userExists := config.ExistsUser(consumer)

	if !userExists || behaviorParams.RotateServiceAccountNow {
		username, password, err := createNewAccount(consumer)
		if err != nil {
			return nil, CredentialNoChange, err
		}
		hashedPassword, err := hashPassword(password)
		if err != nil {
			return nil, CredentialNoChange, err
		}

		config.BasicAuthUsers[username] = hashedPassword

		if err := m.rw.WriteWebConfig(config); err != nil {
			return nil, CredentialNoChange, err
		}

		credentialChangeResult := CredentialCreated
		if userExists {
			credentialChangeResult = CredentialUpdated
		}

		return map[string]string{
			"username": username,
			"password": password,
		}, credentialChangeResult, nil
	}

	return nil, CredentialNoChange, nil
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
