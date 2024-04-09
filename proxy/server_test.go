package proxy

import (
	"github.com/cloudogu/k8s-prometheus/auth/configuration"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_CreateServer(t *testing.T) {
	t.Run("should successfully create new controller", func(t *testing.T) {
		mockAv := NewMockAccountValidator(t)
		conf := configuration.Configuration{PrometheusUrl: "http://localhost:9090"}

		srv := CreateServer(conf, mockAv)

		require.NotNil(t, srv)
		assert.Equal(t, ":8086", srv.Addr)

		routes := srv.Handler.(*gin.Engine).Routes()
		assert.Len(t, routes, 9)
		assert.NotNil(t, routes[0].HandlerFunc)
		assert.Equal(t, "GET", routes[0].Method)
		assert.NotNil(t, routes[1].HandlerFunc)
		assert.Equal(t, "POST", routes[1].Method)
		assert.NotNil(t, routes[2].HandlerFunc)
		assert.Equal(t, "PUT", routes[2].Method)
		assert.NotNil(t, routes[3].HandlerFunc)
		assert.Equal(t, "PATCH", routes[3].Method)
		assert.NotNil(t, routes[4].HandlerFunc)
		assert.Equal(t, "HEAD", routes[4].Method)
		assert.NotNil(t, routes[5].HandlerFunc)
		assert.Equal(t, "OPTIONS", routes[5].Method)
		assert.NotNil(t, routes[6].HandlerFunc)
		assert.Equal(t, "DELETE", routes[6].Method)
		assert.NotNil(t, routes[7].HandlerFunc)
		assert.Equal(t, "CONNECT", routes[7].Method)
		assert.NotNil(t, routes[8].HandlerFunc)
		assert.Equal(t, "TRACE", routes[8].Method)
	})

	t.Run("should panic while creating new controller for error in prometheus url", func(t *testing.T) {
		mockAv := NewMockAccountValidator(t)
		conf := configuration.Configuration{PrometheusUrl: "+:/foo-bar"}

		assert.Panics(t, func() {
			CreateServer(conf, mockAv)
		})
	})
}
