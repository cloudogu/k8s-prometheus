package serviceaccount

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const apiKeyHeader = "X-CES-SA-API-KEY"

func ValidateAPIKey(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyFromRequest := c.Request.Header.Get(apiKeyHeader)

		if apiKeyFromRequest != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"status": 401, "message": "ApiKey not valid"})
			c.Abort()
		}
	}
}
