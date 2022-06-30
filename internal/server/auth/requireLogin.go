package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/observiq/bindplane/internal/server"
)

// RequireLogin should be the last middleware in the middleware chain.
// It checks to see that "authenticated" has been set true by previous middleware.
func RequireLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authenticated, ok := c.Get("authenticated"); !ok || !(authenticated.(bool)) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

// Chain returns the ordered slice of authentication middleware.
func Chain(server server.BindPlane) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		CheckBasic(server),
		CheckSession(server),
		RequireLogin(),
	}
}
