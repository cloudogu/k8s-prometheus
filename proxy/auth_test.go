package proxy

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_BasicAuth(t *testing.T) {
	t.Run("should successfully check basic auth user", func(t *testing.T) {
		mockAv := NewMockaccountValidator(t)
		mockAv.EXPECT().ValidateAccount("user", "password").Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: map[string][]string{},
		}
		c.Request.Header.Set("Authorization", "Basic dXNlcjpwYXNzd29yZA==")

		baHandlerFunc := BasicAuth(mockAv)
		baHandlerFunc(c)

		require.NoError(t, c.Err())
		require.False(t, c.IsAborted())
	})

	t.Run("should fail to check basic auth user for error in validation", func(t *testing.T) {
		mockAv := NewMockaccountValidator(t)
		mockAv.EXPECT().ValidateAccount("user", "password").Return(assert.AnError)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: map[string][]string{},
		}
		c.Request.Header.Set("Authorization", "Basic dXNlcjpwYXNzd29yZA==")

		baHandlerFunc := BasicAuth(mockAv)
		baHandlerFunc(c)

		response := w.Result()
		require.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
		assert.Equal(t, "401 Unauthorized", response.Status)
	})

	t.Run("should fail to check basic auth user for error decoding auth-header", func(t *testing.T) {
		mockAv := NewMockaccountValidator(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: map[string][]string{},
		}
		c.Request.Header.Set("Authorization", "something")

		baHandlerFunc := BasicAuth(mockAv)
		baHandlerFunc(c)

		response := w.Result()
		require.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
		assert.Equal(t, "401 Unauthorized", response.Status)
	})
}

func Test_decodeAuthHeader(t *testing.T) {
	t.Run("should successfully decode auth header", func(t *testing.T) {
		user, password, err := decodeAuthHeader("Basic dXNlcjpwYXNzd29yZA==")

		require.NoError(t, err)
		assert.Equal(t, "user", user)
		assert.Equal(t, "password", password)
	})

	t.Run("should fail decode auth header for error in splitting", func(t *testing.T) {
		_, _, err := decodeAuthHeader("something")

		require.Error(t, err)
		assert.ErrorContains(t, err, "error splitting auth-header")
	})

	t.Run("should fail decode auth header for error in decoding base64", func(t *testing.T) {
		_, _, err := decodeAuthHeader("Basic something")

		require.Error(t, err)
		assert.ErrorContains(t, err, "error decoding auth-header: illegal base64 data at input byte 8")
	})

	t.Run("should fail decode auth header for error in decoding base64", func(t *testing.T) {
		_, _, err := decodeAuthHeader("Basic c29tZXRoaW5n")

		require.Error(t, err)
		assert.ErrorContains(t, err, "error extracting user/password from auth-header")
	})
}
