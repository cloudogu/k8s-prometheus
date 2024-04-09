package serviceaccount

import (
	"github.com/cloudogu/k8s-prometheus/auth/configuration"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateServer(config configuration.Configuration, manager Manager) *http.Server {
	r := gin.Default()

	serviceAccountCtrl := NewController(manager)

	serviceAccountGroup := r.Group("/serviceaccounts")
	serviceAccountGroup.Use(ValidateAPIKey(config.ApiKey))
	serviceAccountGroup.POST("/", serviceAccountCtrl.CreateAccount)
	serviceAccountGroup.DELETE("/:consumer", serviceAccountCtrl.DeleteAccount)

	return &http.Server{
		Addr:    ":8087",
		Handler: r,
	}
}
