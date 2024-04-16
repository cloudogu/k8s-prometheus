package serviceaccount

import (
	"github.com/cloudogu/k8s-prometheus/auth/configuration"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_CreateServer(t *testing.T) {
	t.Run("should successfully create new controller", func(t *testing.T) {
		mockManager := NewMockmanager(t)
		conf := configuration.Configuration{ApiKey: "test123"}

		srv := CreateServer(conf, mockManager)

		require.NotNil(t, srv)
		assert.Equal(t, ":8087", srv.Addr)

		routes := srv.Handler.(*gin.Engine).Routes()
		assert.Len(t, routes, 2)

		assert.NotNil(t, routes[0].HandlerFunc)
		assert.Equal(t, "POST", routes[0].Method)
		assert.Equal(t, "/serviceaccounts/", routes[0].Path)

		assert.NotNil(t, routes[1].HandlerFunc)
		assert.Equal(t, "DELETE", routes[1].Method)
		assert.Equal(t, "/serviceaccounts/:consumer", routes[1].Path)
	})
}
