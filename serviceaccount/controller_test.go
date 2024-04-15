package serviceaccount

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
		req, err := http.NewRequest(http.MethodPost, "/serviceaccounts", strings.NewReader(`{"consumer": "grafana", "params": ["a", "b", "c"]}`))
		require.NoError(t, err)
		ginCtx.Request = req

		mockManager := NewMockmanager(t)
		mockManager.EXPECT().CreateServiceAccount("grafana", []string{"a", "b", "c"}).Return(map[string]string{"username": "user", "password": "password"}, nil)

		ctrl := NewController(mockManager)

		ctrl.CreateAccount(ginCtx)

		assert.Equal(t, http.StatusCreated, w.Code)
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

		ctrl.CreateAccount(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, `{"error":"invalid character 'o' in literal null (expecting 'u')"}`, w.Body.String())
	})

	t.Run("should fail to create new account for missing consumer", func(t *testing.T) {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		req, err := http.NewRequest(http.MethodPost, "/serviceaccounts", strings.NewReader(`{"no-consumer": "grafana", "params": ["a", "b", "c"]}`))
		require.NoError(t, err)
		ginCtx.Request = req

		mockManager := NewMockmanager(t)

		ctrl := NewController(mockManager)

		ctrl.CreateAccount(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, `{"error":"consumer must not be empty"}`, w.Body.String())
	})

	t.Run("should fail to create new account for error in manager", func(t *testing.T) {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		req, err := http.NewRequest(http.MethodPost, "/serviceaccounts", strings.NewReader(`{"consumer": "grafana", "params": ["a", "b", "c"]}`))
		require.NoError(t, err)
		ginCtx.Request = req

		mockManager := NewMockmanager(t)
		mockManager.EXPECT().CreateServiceAccount("grafana", []string{"a", "b", "c"}).Return(nil, assert.AnError)

		ctrl := NewController(mockManager)

		ctrl.CreateAccount(ginCtx)

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
