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

package swagger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestAddRoutes(t *testing.T) {
	// Add routes
	router := gin.Default()
	AddRoutes(router)

	// Start server
	svr := httptest.NewServer(router)
	defer svr.Close()

	// create client
	client := resty.New()
	client.SetBaseURL(svr.URL)

	t.Run("hosts docs at /swagger/index.html", func(t *testing.T) {
		resp, err := client.R().Get("/swagger/index.html")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("redirects from /swagger", func(t *testing.T) {
		resp, err := client.R().Get("/swagger")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
	})
}
