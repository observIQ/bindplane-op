// Copyright  observIQ, Inc
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

package graphql

import (
	"fmt"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/observiq/bindplane-op/internal/graphql/generated"
	"github.com/observiq/bindplane-op/internal/server"
)

// AddRoutes TODO(doc)
func AddRoutes(router gin.IRouter, bindplane server.BindPlane) {
	srv := newHandler(bindplane)

	// TODO(jsirianni) is playground required? https://github.com/observIQ/bindplane/issues/256
	router.GET("/playground", gin.WrapF(playground.Handler("GraphQL playground", "/v1/graphql")))

	// POST for queries and mutations and GET for subscriptions
	router.POST("/graphql", gin.WrapH(srv))
	router.GET("/graphql", gin.WrapH(srv))

	bindplane.Logger().Info(fmt.Sprintf("connect to %s/v1/playground for GraphQL playground", bindplane.Config().BindPlaneURL()))
}

// newHandler creates a *handler.Server configured for Post and Websocket
func newHandler(bindplane server.BindPlane) *handler.Server {
	srv := handler.New(
		generated.NewExecutableSchema(
			generated.Config{
				Resolvers: NewResolver(bindplane)}))

	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})
	srv.Use(extension.Introspection{})
	return srv
}
