package serviceaccount

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudogu/k8s-prometheus/auth/prometheus"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_NewController(t *testing.T) {
	t.Run("should successfully create new controller", func(t *testing.T) {
		mockManager := NewMockmanager(t)

		ctrl := NewController(mockManager)

		assert.Equal(t, mockManager, ctrl.manager)
	})
}

func Test_CreateAccount(t *testing.T) {
	t.Run("should create new account", func(t *testing.T) {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		req, err := http.NewRequest(http.MethodPost, "/serviceaccounts", strings.NewReader(`{"consumer": "grafana", "params": {"a":"", "b":"", "c":""}}`))
		require.NoError(t, err)
		ginCtx.Request = req

		mockManager := NewMockmanager(t)
		mockManager.EXPECT().CreateOrUpdateServiceAccount("grafana", map[string]string{"a": "", "b": "", "c": ""}, mock.Anything).Return(map[string]string{"username": "user", "password": "password"}, prometheus.CredentialCreated, nil)

		ctrl := NewController(mockManager)

		ctrl.CreateOrUpdateAccount(ginCtx)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, `{"password":"password","username":"user"}`, w.Body.String())
	})
	t.Run("should update existing account but the credentials do not change", func(t *testing.T) {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		req, err := http.NewRequest(http.MethodPost, "/serviceaccounts", strings.NewReader(`{"consumer": "grafana", "params": {"a":"", "b":"", "c":""}, "behaviorParams":{"rotateServiceAccountNow": false}}`))
		require.NoError(t, err)
		ginCtx.Request = req

		mockManager := NewMockmanager(t)
		mockManager.EXPECT().CreateOrUpdateServiceAccount("grafana", map[string]string{"a": "", "b": "", "c": ""}, prometheus.BehaviorParams{RotateServiceAccountNow: false}).Return(map[string]string{"username": "user", "password": "password"}, prometheus.CredentialNoChange, nil)

		ctrl := NewController(mockManager)

		// when
		ctrl.CreateOrUpdateAccount(ginCtx)

		// then
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())
	})
	t.Run("should update existing account but a credential rotation is forced", func(t *testing.T) {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		req, err := http.NewRequest(http.MethodPost, "/serviceaccounts", strings.NewReader(`{"consumer": "grafana", "params": {"a":"", "b":"", "c":""}, "behaviorParams":{"rotateServiceAccountNow": true}}`))
		require.NoError(t, err)
		ginCtx.Request = req

		mockManager := NewMockmanager(t)
		mockManager.EXPECT().CreateOrUpdateServiceAccount("grafana", map[string]string{"a": "", "b": "", "c": ""}, prometheus.BehaviorParams{RotateServiceAccountNow: true}).Return(map[string]string{"username": "user", "password": "password"}, prometheus.CredentialUpdated, nil)

		ctrl := NewController(mockManager)

		// when
		ctrl.CreateOrUpdateAccount(ginCtx)

		// then
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `{"password":"password","username":"user"}`, w.Body.String())
	})

	t.Run("should fail to create new account for bad request", func(t *testing.T) {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		req, err := http.NewRequest(http.MethodPost, "/serviceaccounts", strings.NewReader(`no valid JSON`))
		require.NoError(t, err)
		ginCtx.Request = req

		mockManager := NewMockmanager(t)

		ctrl := NewController(mockManager)

		ctrl.CreateOrUpdateAccount(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, `{"error":"invalid character 'o' in literal null (expecting 'u')"}`, w.Body.String())
	})

	t.Run("should fail to create new account for missing consumer", func(t *testing.T) {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		req, err := http.NewRequest(http.MethodPost, "/serviceaccounts", strings.NewReader(`{"no-consumer": "grafana", "params": {"a":"", "b":"", "c":""}}`))
		require.NoError(t, err)
		ginCtx.Request = req

		mockManager := NewMockmanager(t)

		ctrl := NewController(mockManager)

		ctrl.CreateOrUpdateAccount(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, `{"error":"consumer must not be empty"}`, w.Body.String())
	})

	t.Run("should fail to create new account for error in manager", func(t *testing.T) {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		req, err := http.NewRequest(http.MethodPost, "/serviceaccounts", strings.NewReader(`{"consumer": "grafana", "params": {"a":"", "b":"", "c":""}}`))
		require.NoError(t, err)
		ginCtx.Request = req

		mockManager := NewMockmanager(t)
		mockManager.EXPECT().CreateOrUpdateServiceAccount("grafana", map[string]string{"a": "", "b": "", "c": ""}, mock.Anything).Return(nil, prometheus.CredentialNoChange, assert.AnError)

		ctrl := NewController(mockManager)

		ctrl.CreateOrUpdateAccount(ginCtx)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, `{"error":"assert.AnError general error for testing"}`, w.Body.String())
	})
}

func Test_DeleteAccount(t *testing.T) {
	t.Run("should delete account", func(t *testing.T) {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		req, err := http.NewRequest(http.MethodDelete, "/serviceaccounts/grafana", nil)
		require.NoError(t, err)
		ginCtx.Request = req
		ginCtx.AddParam("consumer", "grafana")

		mockManager := NewMockmanager(t)
		mockManager.EXPECT().DeleteServiceAccount("grafana").Return(nil)

		ctrl := NewController(mockManager)

		ctrl.DeleteAccount(ginCtx)

		ginCtx.Writer.WriteHeaderNow()
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "", w.Body.String())
	})

	t.Run("should fail to delete account with no consumer given", func(t *testing.T) {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		req, err := http.NewRequest(http.MethodDelete, "/serviceaccounts/grafana", nil)
		require.NoError(t, err)
		ginCtx.Request = req

		mockManager := NewMockmanager(t)

		ctrl := NewController(mockManager)

		ctrl.DeleteAccount(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, `{"error":"consumer cannot be empty"}`, w.Body.String())
	})

	t.Run("should fail to delete account for error in manager", func(t *testing.T) {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		req, err := http.NewRequest(http.MethodDelete, "/serviceaccounts/grafana", nil)
		require.NoError(t, err)
		ginCtx.Request = req
		ginCtx.AddParam("consumer", "grafana")

		mockManager := NewMockmanager(t)
		mockManager.EXPECT().DeleteServiceAccount("grafana").Return(assert.AnError)

		ctrl := NewController(mockManager)

		ctrl.DeleteAccount(ginCtx)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, `{"error":"assert.AnError general error for testing"}`, w.Body.String())
	})
}
