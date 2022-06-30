package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/observiq/bindplane/internal/server"
	"github.com/observiq/bindplane/internal/server/sessions"
)

// CheckSession checks to see if the attached cookie session is authenticated
// and if so sets authenticated to true on the context.  If not authenticated it
// goes to the next handler.
func CheckSession(server server.BindPlane) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := server.Store().UserSessions().Get(c.Request, sessions.CookieName)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Check the authenticated value in the session storage - if unset or false go to next handler
		if session.Values["authenticated"] == nil || session.Values["authenticated"] == false {
			c.Next()
			return
		}

		c.Keys["authenticated"] = true
	}
}
