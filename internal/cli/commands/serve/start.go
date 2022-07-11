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

package serve

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-multierror"
	cors "github.com/itsjamie/gin-cors"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/common"
	swaggerdocs "github.com/observiq/bindplane-op/docs/swagger"
	"github.com/observiq/bindplane-op/internal/agent"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/internal/cli/commands/profile"
	"github.com/observiq/bindplane-op/internal/graphql"
	"github.com/observiq/bindplane-op/internal/opamp"
	"github.com/observiq/bindplane-op/internal/rest"
	"github.com/observiq/bindplane-op/internal/server"
	"github.com/observiq/bindplane-op/internal/server/auth"
	"github.com/observiq/bindplane-op/internal/server/sessions"
	"github.com/observiq/bindplane-op/internal/store"
	"github.com/observiq/bindplane-op/internal/store/search"
	"github.com/observiq/bindplane-op/ui"
)

// Server is the BindPlane web server, serving HTTP, Websocket and Graphql.
type Server struct {
	logger *zap.Logger
	http   *http.Server
}

// Start starts the BindPlane using the specified Config.
func (s *Server) Start(bindplane *cli.BindPlane, h profile.Helper, forceConsoleColor, skipSeed bool) error {
	config := &bindplane.Config.Server

	// ensure that we have a secret key
	err := s.ensureSecretKey(config, h)
	if err != nil {
		return err
	}

	// initialize the store which stores bindplane resources
	st, err := s.createStore(config)
	if err != nil {
		return err
	}

	// seed the store with the resourceTypes in /resources
	if !skipSeed {
		err := store.Seed(st, s.logger)
		if err != nil {
			s.logger.Error("failed to seed resourceTypes", zap.Error(err))
		}
	}

	// seed the search index
	s.seedSearchIndexes(st)

	// initialize the versions which provides agent versions for updates
	versions := s.createVersions(config)

	// initialize the server which provides access to everything
	server, err := server.NewBindPlane(config, s.logger, st, versions)
	if err != nil {
		return err
	}

	// Gin Routes setup
	setGinMode(config)
	if forceConsoleColor {
		gin.ForceConsoleColor()
	}

	router := gin.New()
	setGinLogging(bindplane, config, router)

	router.Use(cors.Middleware(cors.Config{
		// TODO(andy): This could use a configured variable that references BindPlane UI, e.g. "http://localhost:3000" https://github.com/observiq/bindplane/issues/250
		Origins:        "*",
		Methods:        "GET, PUT, POST, DELETE",
		RequestHeaders: "Origin, Authorization, Content-Type",
		Credentials:    true,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	sessions.AddRoutes(router, server)

	v1 := router.Group("/v1")
	v1.Use(otelgin.Middleware("bindplane"))

	authv1 := v1.Group("/", auth.Chain(server)...)
	rest.AddRestRoutes(authv1, server)

	// download routes do not require authorization
	rest.AddDownloadRoutes(router, server)

	graphql.AddRoutes(authv1, server)

	// Swagger documentation
	swaggerdocs.SwaggerInfo.BasePath = "/v1"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// opamp does its own authorization based on the OnConnecting callback
	err = opamp.AddRoutes(v1, server)
	if err != nil {
		return fmt.Errorf("failed to start OpAMP: %w", err)
	}

	ui.AddRoutes(router, server)

	// TODO(andy): Use a worker pattern here and shutdown cleanly https://github.com/observiq/bindplane/issues/251
	go server.Manager().Start(context.Background())

	s.http = &http.Server{
		Addr:              config.BindAddress(),
		Handler:           router,
		ReadTimeout:       time.Second * 20,
		ReadHeaderTimeout: time.Second * 20,
		WriteTimeout:      time.Second * 20,
		IdleTimeout:       time.Second * 60,
	}

	if config.EnableTLS() {
		c, err := configureTLS(config)
		if err != nil {
			return fmt.Errorf("failed to configure tls: %w", err)
		}
		s.http.TLSConfig = c
	}

	go func() {
		var err error

		if s.http.TLSConfig != nil {
			// Empty strings passed because http server is already
			// configured with the certificates.
			err = s.http.ListenAndServeTLS("", "")
		} else {
			err = s.http.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("listen error", zap.Error(err))
			fmt.Println("listen error:", err.Error())
			os.Exit(200)
		}
	}()

	// Clean shutdown when OS sends signal
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.logger.Info("caught shutdown signal, stopping server")

	return s.Stop()
}

// Stop will stop the server with a timeout
func (s *Server) Stop() error {
	if err := s.stop(); err != nil {
		s.logger.Error("error while trying to stop the server", zap.Error(err))
		return err
	}

	s.logger.Info("server stopped cleanly, exiting")
	return nil
}

// ----------------------------------------------------------------------

func (s *Server) stop() error {
	// TODO(jsirianni): Is 20 seconds the correct shutdown timeout? Generally
	// shutdown is instant, but this gives the server time to handle inflight requests.
	// https://github.com/observiq/bindplane/issues/252
	timeout := 20 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.http.Shutdown(ctx)
}

func (s *Server) createStore(config *common.Server) (store.Store, error) {
	if config.SessionsSecret == "" {
		return nil, errors.New("cannot create store with unset value for sessions-secret, run bindplane init server to set value")
	}

	switch config.StoreType {
	case common.StoreTypeMap:
		return store.NewMapStore(context.Background(), store.Options{
			SessionsSecret:   config.SessionsSecret,
			MaxEventsToMerge: 100,
		}, s.logger), nil

	case common.StoreTypeGoogleCloud:
		s.logger.Info("Using Google Cloud Datastore and Pub/Sub")
		return store.NewGoogleCloudStore(context.Background(), config, s.logger)

	default:
		// case common.StoreTypeBbolt:
		storageFilePath := config.BoltDatabasePath()

		db, err := store.InitDB(storageFilePath)
		// Exit if DB creation is unsuccessful.
		if err != nil {
			return nil, fmt.Errorf("BBolt storage file failed to open: %w", err)
		}

		s.logger.Info("Using BBolt Storage", zap.String("storageFilePath", storageFilePath))
		return store.NewBoltStore(context.Background(), db, store.Options{
			SessionsSecret:   config.SessionsSecret,
			MaxEventsToMerge: 100,
		}, s.logger), nil
	}
}

func (s *Server) createVersions(config *common.Server) agent.Versions {
	var client agent.Client
	if !config.Offline {
		client = agent.NewClient(agent.ClientSettings{
			AgentVersionsURL: config.AgentsServiceURL,
		})
	}
	var cache agent.Cache
	if !config.DisableDownloadsCache {
		cache = agent.NewCache(agent.CacheSettings{
			Directory: config.BindPlaneDownloadsPath(),
		})
	}
	return agent.NewVersions(client, cache, agent.VersionsSettings{
		Logger: s.logger.Named("versions"),
	})
}

func (s *Server) ensureSecretKey(config *common.Server, h profile.Helper) error {
	if config.SecretKey == "" {
		// TODO(andy): generate a new secret key and save it.
		//
		// * this needs to be handled by serve_test.go
		//
		// * it should create the secretKey and save it for the current profile. that means if we run with --profile
		// some-profile-name, we should save the secretKey to some-profile-name even if it isn't the current profile. we
		// don't have enough information to do that here right now.
	}
	return nil
}

func (s *Server) seedSearchIndexes(store store.Store) {
	// seed search indexes
	err := seedConfigurationsIndex(store)
	if err != nil {
		s.logger.Error("unable to seed configurations into the search index, search results will be empty", zap.Error(err))
	}
	err = seedAgentsIndex(store)
	if err != nil {
		s.logger.Error("unable to seed agents into the search index, search results will be empty", zap.Error(err))
	}
}

func seedConfigurationsIndex(s store.Store) error {
	configurations, err := s.Configurations()
	if err != nil {
		return err
	}
	return seedIndex(configurations, s.ConfigurationIndex())
}

func seedAgentsIndex(s store.Store) error {
	agents, err := s.Agents(context.TODO())
	if err != nil {
		return err
	}
	return seedIndex(agents, s.AgentIndex())
}

func seedIndex[T search.Indexed](indexed []T, index search.Index) error {
	var errs error
	for _, i := range indexed {
		err := index.Upsert(i)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

// ----------------------------------------------------------------------

func setGinMode(config *common.Server) {
	switch config.BindPlaneEnv() {
	case common.EnvDevelopment:
		gin.SetMode(gin.DebugMode)
	case common.EnvTest:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
}

func setGinLogging(bindplane *cli.BindPlane, config *common.Server, engine *gin.Engine) {
	// In development, we use the default logger to stdout, but we also use the JSON logger to the log file
	if config.BindPlaneEnv() == common.EnvDevelopment {
		engine.Use(gin.Logger())
	}
	logger := bindplane.Logger()

	// pass empty string for time format to use the zap timestamp
	engine.Use(ginzap.Ginzap(logger, "", false))
	engine.Use(ginzap.RecoveryWithZap(logger, true))
}
