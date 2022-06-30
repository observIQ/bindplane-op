package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/observiq/bindplane/internal/server"
)

func RequireLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authenticated, ok := c.Get("authenticated"); !ok || !(authenticated.(bool)) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

func Chain(server server.BindPlane) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		CheckBasic(server),
		CheckSession(server),
		RequireLogin(),
	}
}
