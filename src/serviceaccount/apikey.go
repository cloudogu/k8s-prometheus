package serviceaccount

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func ValidateAPIKey(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyFromRequest := c.Request.Header.Get("X-CES-SA-API-KEY")

		if apiKeyFromRequest != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"status": 401, "message": "ApiKey not valid"})
			c.Abort()
		}

		return
	}
}
