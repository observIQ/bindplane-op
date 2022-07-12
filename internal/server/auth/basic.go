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
	"github.com/gin-gonic/gin"

	"github.com/observiq/bindplane-op/internal/server"
)

// CheckBasic checks the basic authentication for a request and sets
// authenticated to true if it satisfies the basic auth.  If basic auth is not
// set or is incorrect it goes to the next handler.
func CheckBasic(server server.BindPlane) gin.HandlerFunc {
	configUsername := server.Config().Username
	configPassword := server.Config().Password

	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok || username != configUsername || password != configPassword {
			// Go to next middleware in chain, the final middleware will require authentication is set to true.
			c.Next()
			return
		}

		c.Set("authenticated", true)
	}
}
