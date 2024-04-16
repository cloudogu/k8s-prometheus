package prometheus

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func Test_NewManager(t *testing.T) {
	t.Run("should creat new Manager", func(t *testing.T) {
		sut := NewManager("/some/file.yaml")

		assert.Nil(t, sut.webConfig)
		assert.NotNil(t, sut.rw)
		assert.IsType(t, &WebConfigFileReaderWriter{}, sut.rw)
		assert.Equal(t, sut.rw.(*WebConfigFileReaderWriter).configFile, "/some/file.yaml")
	})
}

func Test_getWebConfig(t *testing.T) {
	t.Run("should get new WebConfig", func(t *testing.T) {
		webConfig := &WebConfig{BasicAuthUsers: map[string]string{"user1": "password1"}}
		mockRW := NewMockWebConfigReaderWriter(t)
		mockRW.EXPECT().ReadWebConfig().Return(webConfig, nil)

		sut := &Manager{rw: mockRW}

		conf, err := sut.getWebConfig()

		require.NoError(t, err)
		assert.Equal(t, webConfig, conf)
		assert.Equal(t, webConfig, sut.webConfig)
	})

	t.Run("should get existing WebConfig", func(t *testing.T) {
		webConfig := &WebConfig{BasicAuthUsers: map[string]string{"existingUser": "wxistingPassword"}}

		sut := &Manager{webConfig: webConfig}

		conf, err := sut.getWebConfig()

		require.NoError(t, err)
		assert.Equal(t, webConfig, conf)
	})

	t.Run("should fail to get WebConfig for error", func(t *testing.T) {
		mockRW := NewMockWebConfigReaderWriter(t)
		mockRW.EXPECT().ReadWebConfig().Return(nil, assert.AnError)

		sut := &Manager{rw: mockRW}

		_, err := sut.getWebConfig()

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func Test_CreateServiceAccount(t *testing.T) {
	t.Run("should create service account", func(t *testing.T) {
		webConfig := &WebConfig{BasicAuthUsers: map[string]string{"user1": "password1"}}
		mockRW := NewMockWebConfigReaderWriter(t)
		mockRW.EXPECT().WriteWebConfig(webConfig).Return(nil)

		sut := &Manager{rw: mockRW, webConfig: webConfig}

		credentials, err := sut.CreateServiceAccount("myConsumer", nil)

		require.NoError(t, err)
		assert.NotEqual(t, "", credentials["username"])
		assert.True(t, strings.HasPrefix(credentials["username"], "myConsumer-"))
		assert.NotEqual(t, "", credentials["password"])

		hashedPassword, exists := sut.webConfig.BasicAuthUsers[credentials["username"]]
		assert.True(t, exists)
		assert.Nil(t, compareHashAndPassword(hashedPassword, credentials["password"]))
	})

	t.Run("should fail to create service account on error reading config", func(t *testing.T) {
		mockRW := NewMockWebConfigReaderWriter(t)
		mockRW.EXPECT().ReadWebConfig().Return(nil, assert.AnError)

		sut := &Manager{rw: mockRW}

		_, err := sut.CreateServiceAccount("myConsumer", nil)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("should fail to create service account on error writing config", func(t *testing.T) {
		webConfig := &WebConfig{BasicAuthUsers: map[string]string{"user1": "password1"}}
		mockRW := NewMockWebConfigReaderWriter(t)
		mockRW.EXPECT().WriteWebConfig(webConfig).Return(assert.AnError)

		sut := &Manager{rw: mockRW, webConfig: webConfig}

		_, err := sut.CreateServiceAccount("myConsumer", nil)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func Test_DeleteServiceAccount(t *testing.T) {
	t.Run("should delete service account", func(t *testing.T) {
		webConfig := &WebConfig{BasicAuthUsers: map[string]string{"otherConsumer-user1": "password1", "myConsumer-user2": "password2"}}
		mockRW := NewMockWebConfigReaderWriter(t)
		mockRW.EXPECT().WriteWebConfig(&WebConfig{BasicAuthUsers: map[string]string{"otherConsumer-user1": "password1"}}).Return(nil)

		sut := &Manager{rw: mockRW, webConfig: webConfig}

		err := sut.DeleteServiceAccount("myConsumer")

		require.NoError(t, err)
		_, exists := sut.webConfig.BasicAuthUsers["myConsumer-user2"]
		assert.False(t, exists)
		_, exists = sut.webConfig.BasicAuthUsers["otherConsumer-user1"]
		assert.True(t, exists)
	})

	t.Run("should fail to delete service account on error reading config", func(t *testing.T) {
		mockRW := NewMockWebConfigReaderWriter(t)
		mockRW.EXPECT().ReadWebConfig().Return(nil, assert.AnError)

		sut := &Manager{rw: mockRW}

		err := sut.DeleteServiceAccount("myConsumer")

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("should fail to delete service account on error writing config", func(t *testing.T) {
		webConfig := &WebConfig{BasicAuthUsers: map[string]string{"otherConsumer-user1": "password1", "myConsumer-user2": "password2"}}
		mockRW := NewMockWebConfigReaderWriter(t)
		mockRW.EXPECT().WriteWebConfig(&WebConfig{BasicAuthUsers: map[string]string{"otherConsumer-user1": "password1"}}).Return(assert.AnError)

		sut := &Manager{rw: mockRW, webConfig: webConfig}

		err := sut.DeleteServiceAccount("myConsumer")

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func Test_ValidateAccount(t *testing.T) {
	t.Run("should validate service account", func(t *testing.T) {
		password1, err := hashPassword("password1")
		require.NoError(t, err)
		password2, err := hashPassword("password2")
		require.NoError(t, err)
		webConfig := &WebConfig{BasicAuthUsers: map[string]string{"user1": password1, "user2": password2}}

		sut := &Manager{webConfig: webConfig}

		err = sut.ValidateAccount("user2", "password2")

		require.NoError(t, err)
	})

	t.Run("should fail validate service account for non existing user", func(t *testing.T) {
		password1, err := hashPassword("password1")
		require.NoError(t, err)
		password2, err := hashPassword("password2")
		require.NoError(t, err)
		webConfig := &WebConfig{BasicAuthUsers: map[string]string{"user1": password1, "user2": password2}}

		sut := &Manager{webConfig: webConfig}

		err = sut.ValidateAccount("user3", "foo")

		require.Error(t, err)
		assert.ErrorContains(t, err, "cloud not find user with name: user3")
	})

	t.Run("should fail validate service account for wrong password", func(t *testing.T) {
		password1, err := hashPassword("password1")
		require.NoError(t, err)
		password2, err := hashPassword("password2")
		require.NoError(t, err)
		webConfig := &WebConfig{BasicAuthUsers: map[string]string{"user1": password1, "user2": password2}}

		sut := &Manager{webConfig: webConfig}

		err = sut.ValidateAccount("user2", "foo")

		require.Error(t, err)
		assert.ErrorContains(t, err, "crypto/bcrypt: hashedPassword is not the hash of the given password")
	})

	t.Run("should fail to validate service account on error reading config", func(t *testing.T) {
		mockRW := NewMockWebConfigReaderWriter(t)
		mockRW.EXPECT().ReadWebConfig().Return(nil, assert.AnError)

		sut := &Manager{rw: mockRW}

		err := sut.ValidateAccount("user2", "password2")

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
	})
}
