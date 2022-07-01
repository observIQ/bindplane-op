// Copyright  observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
