package serviceaccount

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_ValidateAPIKey(t *testing.T) {
	t.Run("should validate api-key", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: map[string][]string{},
		}
		c.Request.Header.Set("X-CES-SA-API-KEY", "secretApiKey")

		handlerFunc := ValidateAPIKey("secretApiKey")
		handlerFunc(c)

		require.NoError(t, c.Err())
		require.False(t, c.IsAborted())
	})

	t.Run("should fail to validate api-key for wrong api-key", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: map[string][]string{},
		}
		c.Request.Header.Set("X-CES-SA-API-KEY", "secretApiKey")

		handlerFunc := ValidateAPIKey("superSecretApiKey")
		handlerFunc(c)

		require.NoError(t, c.Err())
		require.True(t, c.IsAborted())
		response := w.Result()
		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
		assert.Equal(t, "401 Unauthorized", response.Status)
		assert.Equal(t, "{\"message\":\"ApiKey not valid\",\"status\":401}", w.Body.String())
	})
}
