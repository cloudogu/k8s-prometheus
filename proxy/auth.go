package proxy

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"strings"
)

type AccountValidator interface {
	ValidateAccount(username string, password string) error
}

func BasicAuth(av AccountValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, err := decodeAuthHeader(c.Request.Header.Get("Authorization"))
		if err != nil {
			slog.Error("error decoding auth-header", "err", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if err := av.ValidateAccount(username, password); err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

func decodeAuthHeader(authHeader string) (string, string, error) {
	// Split the header into its components
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		return "", "", fmt.Errorf("error splitting auth-header")
	}

	// Decode the base64 encoded credentials
	credentials, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", fmt.Errorf("error decoding auth-header: %w", err)
	}

	// Extract username and password
	auth := strings.SplitN(string(credentials), ":", 2)
	if len(auth) != 2 {
		return "", "", fmt.Errorf("error extracting user/password from auth-header")
	}

	return auth[0], auth[1], err
}
