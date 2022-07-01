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

package rest

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/internal/agent"
	"github.com/observiq/bindplane-op/internal/server"
)

// AddDownloadRoutes adds /download/* routes to the gin HTTP router
func AddDownloadRoutes(router gin.IRouter, bindplane server.BindPlane) {
	router.GET("/downloads/:agent/:version/:platform/:type/:file", func(c *gin.Context) { getAgentDownload(c, bindplane) })
}

// @Summary Get Agent Download
// @Description Get the agent download with the specified parameters
// @Produce octet-stream
// @Router /downloads/{agent}/{version}/{platform}/{type}/{file} [get]
// @Param agent 	  path	string	true "observiq-agent"
// @Param version 	path	string	true "2.1.1"
// @Param platform 	path	string	true "darwin-arm64"
// @Param type 	    path	string	true "installer"
// @Param file 	    path	string	true "observiq-agent-installer.sh"
// @Success 200 {file} octet-stream
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func getAgentDownload(c *gin.Context, bindplane server.BindPlane) {
	versions := bindplane.Versions()

	artifactType := agent.ArtifactType(c.Param("type"))

	agentType := c.Param("agent")
	if agentType != "observiq-agent" {
		c.Status(404)
		return
	}

	// find the specified version
	version, err := versions.Version(c.Param("version"))
	if err != nil {
		if errors.Is(err, agent.ErrVersionNotFound) {
			handleErrorResponse(c, http.StatusNotFound, err)
			return
		}
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}
	if version == nil {
		c.Status(http.StatusNotFound)
		return
	}

	bindplane.Logger().Info("got version", zap.Any("version", version))

	// find the installer for the specified version and platform
	installer := versions.Artifact(artifactType, version, c.Param("platform"))

	// get a reader from the installer
	reader, err := installer.Reader()
	if err != nil {
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	// copy to the output stream
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}
