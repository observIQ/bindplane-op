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

package sessions

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/server"
	"github.com/observiq/bindplane-op/internal/store"
)

func TestAddRoutes(t *testing.T) {
	// Setup
	router := gin.Default()
	svr := httptest.NewServer(router)
	defer svr.Close()

	logger := zap.NewNop()
	s := store.NewMapStore(logger, "super-secret-key")
	bindplane, err := server.NewBindPlane(&common.Server{}, zap.NewNop(), s, nil)
	require.NoError(t, err)

	t.Run("adds /login /logout and /verify", func(t *testing.T) {
		AddRoutes(router, bindplane)

		routes := router.Routes()

		var hasLogin bool
		var hasLogout bool
		var hasVerify bool

		for _, r := range routes {
			switch r.Path {
			case "/login":
				hasLogin = true
			case "/logout":
				hasLogout = true
			case "/verify":
				hasVerify = true
			}
		}

		require.True(t, hasLogin)
		require.True(t, hasLogout)
		require.True(t, hasVerify)
	})
}

func TestHandleLogin(t *testing.T) {
	// Setup
	cfg := &common.Server{}
	cfg.Password = "secret"
	cfg.Username = "user"

	router := gin.Default()
	svr := httptest.NewServer(router)
	defer svr.Close()

	logger := zap.NewNop()
	s := store.NewMapStore(logger, "super-secret-key")
	bindplane, err := server.NewBindPlane(cfg, zap.NewNop(), s, nil)
	require.NoError(t, err)

	AddRoutes(router, bindplane)

	t.Run(fmt.Sprintf("sets the %s cookie with correct credentials", CookieName), func(t *testing.T) {
		client := resty.New()
		client.SetBaseURL(svr.URL)

		resp, err := client.R().SetFormData(
			map[string]string{
				"username": "user",
				"password": "secret",
			},
		).Post("/login")

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		var foundCookie bool
		for _, cookie := range resp.Cookies() {
			if strings.Contains(cookie.Name, CookieName) {
				foundCookie = true
			}
		}

		require.True(t, foundCookie)
	})

	t.Run("will not set a cookie for bad credentials", func(t *testing.T) {
		client := resty.New()
		client.SetBaseURL(svr.URL)

		resp, err := client.R().SetFormData(
			map[string]string{
				"username": "user",
				"password": "bad-secret",
			},
		).Post("/login")

		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode())
		require.Empty(t, resp.Cookies())
	})
}

func TestLogin(t *testing.T) {
	// Setup
	cfg := &common.Server{}
	cfg.Password = "secret"
	cfg.Username = "user"

	logger := zap.NewNop()

	s := store.NewMapStore(logger, "super-secret-key")

	bindplane, err := server.NewBindPlane(cfg, zap.NewNop(), s, nil)
	require.NoError(t, err)

	t.Run("will not set authenticated to true for invalid creds", func(t *testing.T) {
		// Create a Post Form Request with username and password
		req := httptest.NewRequest("POST", "/login", nil)
		req.PostForm = url.Values{
			"username": []string{"user"},
			"password": []string{"bad-secret"},
		}

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		// Log in with context
		login(ctx, bindplane)

		// Make sure authenticated is set
		session, err := bindplane.Store().UserSessions().Get(ctx.Request, CookieName)
		require.NoError(t, err)
		require.Nil(t, session.Values["authenticated"], "expect the authenticated key to not be set")
	})

	t.Run("sets authenticated to true on the cookie with valid creds", func(t *testing.T) {
		// Create a Post Form Request with username and password
		req := httptest.NewRequest("POST", "/login", nil)
		req.PostForm = url.Values{
			"username": []string{"user"},
			"password": []string{"secret"},
		}

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		// Log in with context
		login(ctx, bindplane)

		// Make sure authenticated is set
		session, err := bindplane.Store().UserSessions().Get(ctx.Request, CookieName)
		require.NoError(t, err)
		require.Equal(t, session.Values["authenticated"], true)
	})

}

func TestLogout(t *testing.T) {
	// Setup
	cfg := &common.Server{}
	cfg.Password = "secret"
	cfg.Username = "user"

	logger := zap.NewNop()

	s := store.NewMapStore(logger, "super-secret-key")

	bindplane, err := server.NewBindPlane(cfg, zap.NewNop(), s, nil)
	require.NoError(t, err)

	t.Run("will set authenticated to false for a logged in context", func(t *testing.T) {
		cookie := getLoggedInCookie(t, bindplane)

		// Make a logout request with the cookie we got from the login request
		logoutReq := httptest.NewRequest("PUT", "/logout", nil)
		logoutReq.AddCookie(cookie)

		logoutCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
		logoutCtx.Request = logoutReq

		// log the context out
		logout(logoutCtx, bindplane)

		// verify authenticated is set to false
		session, _ := bindplane.Store().UserSessions().Get(logoutCtx.Request, CookieName)
		require.False(t, session.Values["authenticated"].(bool))
	})
}

func TestVerify(t *testing.T) {
	// Setup
	cfg := &common.Server{}
	cfg.Password = "secret"
	cfg.Username = "user"

	logger := zap.NewNop()

	s := store.NewMapStore(logger, "super-secret-key")

	bindplane, err := server.NewBindPlane(cfg, zap.NewNop(), s, nil)
	require.NoError(t, err)

	t.Run("aborts with status 401 when authenticated is unset", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/verify", nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		verify(ctx, bindplane)

		require.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})

	t.Run("aborts with status 401 when authenticated is false", func(t *testing.T) {
		cookie := getLoggedOutCookie(t, bindplane)
		req := httptest.NewRequest("GET", "/verify", nil)
		req.AddCookie(cookie)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		verify(ctx, bindplane)
		require.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})

}

func getLoggedInContext(t *testing.T, bindplane server.BindPlane, w http.ResponseWriter) *gin.Context {
	// Make a login request
	loginRequest := httptest.NewRequest("POST", "/login", nil)
	loginRequest.PostForm = url.Values{
		"username": []string{"user"},
		"password": []string{"secret"},
	}

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = loginRequest

	login(ctx, bindplane)
	return ctx
}

func getLoggedInCookie(t *testing.T, bindplane server.BindPlane) *http.Cookie {
	w := httptest.NewRecorder()
	getLoggedInContext(t, bindplane, w)

	return w.Result().Cookies()[0]
}

func getLoggedOutCookie(t *testing.T, bindplane server.BindPlane) *http.Cookie {
	cookie := getLoggedInCookie(t, bindplane)

	// Make a logout request with the cookie we got from the login request
	logoutReq := httptest.NewRequest("PUT", "/logout", nil)
	logoutReq.AddCookie(cookie)

	w := httptest.NewRecorder()
	session, _ := bindplane.Store().UserSessions().Get(logoutReq, CookieName)
	session.Values["authenticated"] = false
	session.Save(logoutReq, w)

	return w.Result().Cookies()[0]
}
