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

package opamp

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/open-telemetry/opamp-go/protobufs"
	opampSvr "github.com/open-telemetry/opamp-go/server"
	opamp "github.com/open-telemetry/opamp-go/server/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/observiq/bindplane-op/internal/server"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/model/observiq"
)

var tracer = otel.Tracer("bindplane/opamp")

// ProtocolName is "opamp"
const ProtocolName = "opamp"

var compatibleOpAMPVersions = []string{"v0.2.0"}

const (
	headerAuthorization = "Authorization"
	headerUserAgent     = "User-Agent"
	headerOpAMPVersion  = "OpAMP-Version"
	headerAgentID       = "Agent-ID"
	headerAgentVersion  = "Agent-Version"
	headerAgentHostname = "Agent-Hostname"
)

// AddRoutes adds the routes used by opamp, currently /v1/opamp
func AddRoutes(router gin.IRouter, bindplane server.BindPlane) error {
	server := opampSvr.New(bindplane.Logger().Sugar())

	callbacks := newServer(bindplane.Manager(), bindplane.Logger())
	settings := opampSvr.Settings{
		Callbacks: callbacks,
	}

	handler, err := server.Attach(settings)
	if err != nil {
		return fmt.Errorf("error attempting to attach the OpAMP server: %w", err)
	}

	router.Any("/opamp", gin.WrapF(http.HandlerFunc(handler)))

	bindplane.Manager().EnableProtocol(callbacks)

	return nil
}

const (
	capabilities = protobufs.ServerCapabilities_AcceptsStatus | protobufs.ServerCapabilities_AcceptsEffectiveConfig | protobufs.ServerCapabilities_OffersRemoteConfig
)

type opampServer struct {
	manager                 server.Manager
	connections             *connections
	compatibleOpAMPVersions []string
	logger                  *zap.Logger
}

var _ server.Protocol = (*opampServer)(nil)
var _ opamp.Callbacks = (*opampServer)(nil)

func newServer(manager server.Manager, logger *zap.Logger) *opampServer {
	return &opampServer{
		manager:                 manager,
		connections:             newConnections(),
		compatibleOpAMPVersions: compatibleOpAMPVersions,
		logger:                  logger,
	}
}

// ----------------------------------------------------------------------
// The following callbacks will never be called concurrently for the same
// connection. They may be called concurrently for different connections.

// OnConnecting is called when there is a new incoming connection.
// The handler can examine the request and either accept or reject the connection.
// To accept:
//   Return ConnectionResponse with Accept=true.
//   HTTPStatusCode and HTTPResponseHeader are ignored.
//
// To reject:
//   Return ConnectionResponse with Accept=false. HTTPStatusCode MUST be set to
//   non-zero value to indicate the rejection reason (typically 401, 429 or 503).
//   HTTPResponseHeader may be optionally set (e.g. "Retry-After: 30").
func (s *opampServer) OnConnecting(request *http.Request) opamp.ConnectionResponse {
	ctx, span := tracer.Start(request.Context(), "opamp/connecting")
	defer span.End()

	s.logger.Info("OnConnecting", zap.Any("headers", request.Header), zap.String("RemoteAddr", request.RemoteAddr))

	// check for compatibility
	headers := parseAgentHeaders(request)
	if headers == nil || !slices.Contains(s.compatibleOpAMPVersions, headers.opampVersion) {
		// no version header, agent version is <= 1.2.0 or OpAMP version incompatible
		s.logger.Error("unable to connect to incompatible agent",
			zap.Any("headers", request.Header),
			zap.String("RemoteAddr", request.RemoteAddr),
			zap.Strings("compatibleOpAMPVersions", s.compatibleOpAMPVersions),
		)
		return opamp.ConnectionResponse{
			Accept:         false,
			HTTPStatusCode: http.StatusUpgradeRequired,
			HTTPResponseHeader: map[string]string{
				"Upgrade": fmt.Sprintf("OpAMP/%s", s.compatibleOpAMPVersions[0]),
			},
		}
	}

	accept := s.manager.VerifySecretKey(ctx, headers.secretKey)
	if !accept {
		return opamp.ConnectionResponse{
			Accept:         false,
			HTTPStatusCode: http.StatusUnauthorized,
		}
	}

	return opamp.ConnectionResponse{
		Accept:         true,
		HTTPStatusCode: http.StatusOK,
	}
}

type agentHeaders struct {
	opampVersion string
	id           string
	version      string
	hostname     string
	secretKey    string
}

func parseAgentHeaders(request *http.Request) *agentHeaders {
	authHeader := request.Header.Get(headerAuthorization)
	secretKey := strings.Replace(authHeader, "Secret-Key ", "", 1)
	if secretKey == authHeader {
		// check for missing Secret-Key identifier
		secretKey = ""
	}
	return &agentHeaders{
		opampVersion: request.Header.Get(headerOpAMPVersion),
		id:           request.Header.Get(headerAgentID),
		version:      request.Header.Get(headerAgentVersion),
		hostname:     request.Header.Get(headerAgentHostname),
		secretKey:    secretKey,
	}
}

// OnConnected is called when the WebSocket connection is successfully established after OnConnecting() returns and the
// HTTP connection is upgraded to WebSocket.
//
// opamp.Connection doesn't have much information that we can use here
func (s *opampServer) OnConnected(conn opamp.Connection) {
	_, span := tracer.Start(context.TODO(), "opamp/connected")
	defer span.End()
}

// OnMessage is called when a message is received from the connection. Can happen
// only after OnConnected().
func (s *opampServer) OnMessage(conn opamp.Connection, message *protobufs.AgentToServer) *protobufs.ServerToAgent {
	ctx, span := tracer.Start(context.TODO(), "opamp/message")
	defer span.End()

	agentID := message.InstanceUid
	hasConfiguration := message.GetEffectiveConfig().GetConfigMap() != nil

	span.SetAttributes(
		attribute.String("bindplane.agent.id", agentID),
		attribute.String("bindplane.component", "opamp"),
		attribute.Bool("bindplane.opamp.hasConfiguration", hasConfiguration),
	)

	s.logger.Info("OpAMP agent message", zap.String("agentID", agentID), zap.Strings("submessages", messageComponents(message)))
	s.connections.connect(conn, agentID)

	response := &protobufs.ServerToAgent{
		InstanceUid:  agentID,
		Capabilities: capabilities,
	}

	// verify the configuration and modify the response message
	err := s.verifyAgentConfig(ctx, conn, agentID, message, response)
	if err != nil {
		s.logger.Error("error verifying the agent configuration", zap.Error(err))
		// send an error response
		// TODO(andy): Ok to report the exact error?
		response.ErrorResponse = &protobufs.ServerErrorResponse{
			Type:         protobufs.ServerErrorResponse_Unknown,
			ErrorMessage: err.Error(),
		}
	}
	s.logger.Info("sending response to the agent", zap.Any("agentID", agentID), zap.Any("response", response))

	return response
}

// OnConnectionClose is called when the WebSocket connection is closed.
// Typically, preceded by OnDisconnect() unless the client misbehaves or the
// connection is lost.
func (s *opampServer) OnConnectionClose(conn opamp.Connection) {
	ctx, span := tracer.Start(context.TODO(), "opamp/disconnected")
	defer span.End()

	agentID := s.connections.agentID(conn)
	s.logger.Info("OpAMP agent disconnected", zap.String("AgentID", agentID))
	s.connections.disconnect(conn)
	if agentID == "" {
		return
	}
	_, err := s.manager.UpsertAgent(ctx, agentID, func(agent *model.Agent) {
		agent.Disconnect()
	})
	if err != nil {
		s.logger.Error("error trying to save disconnected state of agent", zap.String("agentID", agentID), zap.Error(err))
		return
	}
}

// ----------------------------------------------------------------------
// Protocol implementation

func (s *opampServer) Name() string {
	return ProtocolName
}

// ConnectedAgentIDs should return a slice of the currently connected agent IDs
func (s *opampServer) ConnectedAgentIDs(ctx context.Context) ([]string, error) {
	ctx, span := tracer.Start(ctx, "opamp/ConnectedAgentIDs")
	defer span.End()

	return s.connections.agentIDs(), nil
}

func (s *opampServer) Disconnect(agentID string) bool {
	conn := s.connections.connection(agentID)
	if conn != nil {
		s.connections.disconnect(conn)
		return true
	}
	return false
}

// Connected returns true if the specified agent ID is connected
func (s *opampServer) Connected(agentID string) bool {
	return s.connections.connected(agentID)
}

// UpdateAgent should send a message to the specified agent to update the configuration to match the
// specified configuration.
func (s *opampServer) UpdateAgent(ctx context.Context, agent *model.Agent, updates *server.AgentUpdates) error {
	conn := s.connections.connection(agent.ID)
	if conn == nil {
		// agent not connected, nothing to do
		return nil
	}
	ctx, span := tracer.Start(ctx, "opamp/UpdateAgent", trace.WithAttributes(
		attribute.String("bindplane.agent.id", agent.ID),
	))
	defer span.End()

	agentConfiguration, err := observiq.DecodeAgentConfiguration(agent.Configuration)
	if err != nil {
		// start with a blank configuration if the current isn't available
		agentConfiguration = &observiq.AgentConfiguration{}
	}

	newConfiguration, err := s.updatedConfiguration(ctx, agentConfiguration, updates)
	if err != nil {
		return fmt.Errorf("unable to get the new configuration for agent [%s]: %w", agent.ID, err)
	}

	serverToAgent := &protobufs.ServerToAgent{
		InstanceUid:  agent.ID,
		Capabilities: capabilities,
		Flags:        protobufs.ServerToAgent_ReportFullState,
	}

	if newConfiguration.Empty() {
		s.logger.Info("agent already has the correct configuration")
	} else {
		agentRawConfiguration := agentConfiguration.Raw()
		newRawConfiguration := newConfiguration.Raw()

		serverToAgent.RemoteConfig = agentRemoteConfig(&newRawConfiguration, &agentRawConfiguration)

		// use a separate goroutine to avoid blocking on the channel write
		go func() {
			// change the agent status to Configuring, but ignore any failure as this status is considered nice to have and not required to update the agent
			_, _ = s.manager.UpsertAgent(ctx, agent.ID, func(current *model.Agent) { current.Status = model.Configuring })
		}()
	}

	if updates.Version != "" {
		s.logger.Info("sending agent update to version", zap.String("version", updates.Version))
		// Dynamically figure out file based on OS
		downloadableFile := determineDownloadableFile(strings.ToLower(agent.OperatingSystem))
		serverToAgent.PackagesAvailable = &protobufs.PackagesAvailable{
			AllPackagesHash: []byte(updates.Version),
			Packages: map[string]*protobufs.PackageAvailable{
				"": {
					Type:    protobufs.PackageAvailable_TopLevelPackage,
					Version: updates.Version,
					File:    downloadableFile,
					Hash:    []byte(updates.Version),
				},
			},
		}

		// TODO(andy): Remove this before merging
		_, _ = s.manager.UpsertAgent(ctx, agent.ID, func(current *model.Agent) { current.VersionPending = "" })
	}

	// if the message doesn't have a new configuration or a new package available, do nothing
	if serverToAgent.RemoteConfig == nil && serverToAgent.PackagesAvailable == nil {
		return nil
	}

	return s.send(context.Background(), conn, serverToAgent)
}

func determineDownloadableFile(operatingSystem string) *protobufs.DownloadableFile {
	switch {
	case strings.Contains(operatingSystem, "mac"): // Mac
		return &protobufs.DownloadableFile{
			DownloadUrl: "",
			ContentHash: []byte{},
		}
	case strings.Contains(operatingSystem, "windows"): // Windows
		return &protobufs.DownloadableFile{
			DownloadUrl: "",
			ContentHash: []byte{},
		}
	default: // LINUX
		return &protobufs.DownloadableFile{
			DownloadUrl: "",
			ContentHash: []byte{},
		}
	}
}

// SendHeartbeat sends a heartbeat to the agent to keep the websocket open
func (s *opampServer) SendHeartbeat(agentID string) error {
	conn := s.connections.connection(agentID)
	if conn != nil {
		return s.send(context.Background(), conn, &protobufs.ServerToAgent{})
	}
	return nil
}

func (s *opampServer) send(ctx context.Context, conn opamp.Connection, msg *protobufs.ServerToAgent) error {
	lock := s.connections.sendLock(conn)
	lock.Lock()
	defer lock.Unlock()
	return conn.Send(ctx, msg)
}

// ----------------------------------------------------------------------

func (s *opampServer) verifyAgentConfig(ctx context.Context, conn opamp.Connection, agentID string, message *protobufs.AgentToServer, response *protobufs.ServerToAgent) error {
	ctx, span := tracer.Start(ctx, "opamp/verifyAgentConfig")
	defer span.End()

	// store the current configuration as reported by status
	agent, state, err := s.updateAgentState(ctx, agentID, conn, message, response)
	if err != nil {
		return fmt.Errorf("unable to update agent [%s]: %w", agentID, err)
	}

	return s.updateAgentConfig(ctx, agent, state, response)
}

// updateAgentConfig updates the current configuration by setting the RemoteConfig message if necessary
func (s *opampServer) updateAgentConfig(ctx context.Context, agent *model.Agent, state *agentState, response *protobufs.ServerToAgent) error {
	agentRawConfiguration := state.Configuration()
	if agentRawConfiguration == nil {
		s.logger.Info("no configuration available to verify, requesting from agent")
		return nil
	}

	agentConfiguration, err := agentRawConfiguration.Parse()
	if err != nil {
		// TODO(andy): ignore the current unparsable configuration and force new configuration?
		return fmt.Errorf("unable to parse the current agent configuration: %w", err)
	}

	// check the manager for any updates that should be applied to this agent
	updates, err := s.manager.AgentUpdates(ctx, agent)
	if err != nil {
		return fmt.Errorf("unable to get agent updates [%s]: %w", agent.ID, err)
	}

	serverConfiguration, err := s.updatedConfiguration(ctx, agentConfiguration, updates)
	if err != nil {
		return fmt.Errorf("unable to compute the updated agent configuration [%s]: %w", agent.ID, err)
	}

	// compare the configurations and compute a difference
	newConfiguration := observiq.ComputeConfigurationUpdates(&serverConfiguration, agentConfiguration)

	if newConfiguration.Empty() {
		// existing config is correct
		s.logger.Info("agent running with the correct config")
		return nil
	}

	rawNewConfiguration := newConfiguration.Raw()
	remoteConfig := agentRemoteConfig(&rawNewConfiguration, agentRawConfiguration)

	// check to see if we already tried this and received an error
	if bytes.Equal(state.Status.GetRemoteConfigStatus().GetLastRemoteConfigHash(), remoteConfig.GetConfigHash()) {
		s.logger.Info("already attempted to send this configuration")
		return nil
	}

	// change the agent status to Configuring, but ignore any failure as this status is considered nice to have and not
	// required to update the agent
	_, _ = s.manager.UpsertAgent(ctx, agent.ID, func(current *model.Agent) { current.Status = model.Configuring })

	s.logger.Info("agent running with outdated config", zap.Any("cur", agentConfiguration.Collector), zap.Any("new", serverConfiguration.Collector))
	response.RemoteConfig = remoteConfig

	return nil
}

func (s *opampServer) updatedConfiguration(ctx context.Context, agentConfiguration *observiq.AgentConfiguration, updates *server.AgentUpdates) (diff observiq.AgentConfiguration, err error) {
	// Configuration => collector.yaml
	if updates.Configuration != nil {
		newCollectorYAML, err := updates.Configuration.Render(ctx, s.manager.ResourceStore())
		if err != nil {
			return diff, err
		}
		diff.Collector = newCollectorYAML
	}

	// Labels => manager.yaml
	if updates.Labels != nil && !agentConfiguration.HasLabels(updates.Labels.String()) {
		diff.Manager = agentConfiguration.Manager
		diff.ReplaceLabels(updates.Labels.String())
	}

	return diff, nil
}

// agentRemoteConfig generates the protobuf for sending this Config to an agent using the OpAMP protocol
func agentRemoteConfig(updates *observiq.RawAgentConfiguration, agentRaw *observiq.RawAgentConfiguration) *protobufs.AgentRemoteConfig {
	// only store the configs that exist for the agent
	configMap := map[string]*protobufs.AgentConfigFile{}
	if updates.Collector != nil {
		configMap[observiq.CollectorFilename] = &protobufs.AgentConfigFile{Body: updates.Collector}
	}
	if updates.Logging != nil {
		configMap[observiq.LoggingFilename] = &protobufs.AgentConfigFile{Body: updates.Logging}
	}
	if updates.Manager != nil {
		configMap[observiq.ManagerFilename] = &protobufs.AgentConfigFile{Body: updates.Manager}
	}

	return &protobufs.AgentRemoteConfig{
		Config: &protobufs.AgentConfigMap{
			ConfigMap: configMap,
		},
		ConfigHash: computeHash(updates, agentRaw),
	}
}

func computeHash(updates *observiq.RawAgentConfiguration, agentRaw *observiq.RawAgentConfiguration) []byte {
	combined := agentRaw.ApplyUpdates(updates)
	return combined.Hash()
}
