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
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/observiq/bindplane/internal/server"
	"go.uber.org/zap"
)

const (
	// CookieName is the name of the cookie used for session authentication.
	CookieName = "BP_OP_AUTH"
)

func login(ctx *gin.Context, bindplane server.BindPlane) {
	session, err := bindplane.Store().UserSessions().Get(ctx.Request, CookieName)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("failed to retrieve session"))
		bindplane.Logger().Error("failed to retrieve session at login", zap.Error(err))
		return
	}

	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	if password != bindplane.Config().Password || username != bindplane.Config().Username {
		ctx.AbortWithError(http.StatusUnauthorized, errors.New("incorrect username or password"))
		return
	}

	// Set user as authenticated
	session.Values["authenticated"] = true

	bindplane.Logger().Info("logging in user.", zap.String("user", username))

	// Save and write the session
	if err := session.Save(ctx.Request, ctx.Writer); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("failed to save session"))
		bindplane.Logger().Error("failed to save session after login", zap.Error(err))
	}
}

func logout(ctx *gin.Context, bindplane server.BindPlane) {
	session, err := bindplane.Store().UserSessions().Get(ctx.Request, CookieName)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("failed to retrieve session"))
		bindplane.Logger().Error("failed to retrieve session for logout", zap.Error(err))
		return
	}

	// Revoke users authentication
	session.Values["authenticated"] = false
	// Delete the cookie
	session.Options.MaxAge = -1

	bindplane.Logger().Info("logging out user.", zap.Any("user", session.Values["user"]))
	// Save and write the session
	if err := session.Save(ctx.Request, ctx.Writer); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("failed to save session"))
		bindplane.Logger().Error("failed to save session after logout", zap.Error(err))
	}
}

func verify(c *gin.Context, bindplane server.BindPlane) {
	session, err := bindplane.Store().UserSessions().Get(c.Request, CookieName)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("failed to retrieve session"))
		bindplane.Logger().Error("failed to save session during verify", zap.Error(err))
		return
	}

	if session.Values["authenticated"] == true {
		return
	}

	c.AbortWithError(http.StatusUnauthorized, errors.New("unauthorized"))
}

// AddRoutes adds the login, logout, and verify route used for session authentication.
func AddRoutes(router gin.IRouter, bindplane server.BindPlane) {
	router.POST("/login", func(ctx *gin.Context) { login(ctx, bindplane) })
	router.PUT("/logout", func(ctx *gin.Context) { logout(ctx, bindplane) })
	router.GET("/verify", func(ctx *gin.Context) { verify(ctx, bindplane) })
}
