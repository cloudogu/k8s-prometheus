package proxy

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
)

func Test_NewController(t *testing.T) {
	t.Run("should successfully create new controller", func(t *testing.T) {
		ctrl, err := NewController("http://localhost:9090")

		require.NoError(t, err)
		assert.IsType(t, &httputil.ReverseProxy{}, ctrl.proxy)
	})

	t.Run("should fail to create new controller for error in url parsing", func(t *testing.T) {
		_, err := NewController("+:/''#someßThing")

		require.Error(t, err)
		assert.ErrorContains(t, err, "error parsing proxy-url: ")
	})
}

func Test_Proxy(t *testing.T) {
	t.Run("should proxy request", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, ginEngine := gin.CreateTestContext(w)

		testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/query", r.URL.Path)
			assert.Equal(t, "b", r.URL.Query().Get("a"))
			assert.Equal(t, "1", r.URL.Query().Get("c"))
			assert.Equal(t, "foo", r.Header.Get("Test"))
			assert.Equal(t, "bar", r.Header.Get("Foo"))
		}))

		ctx, _ := context.WithCancel(context.TODO())
		getReq, err := http.NewRequestWithContext(ctx, "GET", "http://prometheus:5050/api/query?a=b&c=1", nil)
		require.NoError(t, err)
		getReq.Header.Add("Test", "foo")
		getReq.Header.Set("Foo", "bar")
		c.Request = getReq

		ctrl, err := NewController(testSrv.URL)
		require.NoError(t, err)

		ginEngine.GET("/*proxyPath", ctrl.Proxy)

		ginEngine.ServeHTTP(w, getReq)

		require.NoError(t, c.Err())
		require.False(t, c.IsAborted())
	})
}
