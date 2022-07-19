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
	"context"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/server"
	"github.com/observiq/bindplane-op/internal/server/mocks"
	"github.com/observiq/bindplane-op/internal/store"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/model/observiq"
	"github.com/open-telemetry/opamp-go/protobufs"
	opamp "github.com/open-telemetry/opamp-go/server/types"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testServer(manager server.Manager) *opampServer {
	return newServer(manager, zap.NewNop())
}

func testResource[T model.Resource](t *testing.T, name string) T {
	return fileResource[T](t, filepath.Join("testfiles", name))
}
func fileResource[T model.Resource](t *testing.T, path string) T {
	resources, err := model.ResourcesFromFile(path)
	require.NoError(t, err)

	parsed, err := model.ParseResources(resources)
	require.NoError(t, err)
	require.Len(t, parsed, 1)

	resource, ok := parsed[0].(T)
	require.True(t, ok)
	return resource
}

func TestServerSendHeartbeat(t *testing.T) {
	manager := &mocks.Manager{}
	conn := &mocks.Connection{}
	server := testServer(manager)
	server.connections.connect(conn, "known")

	conn.On("Send", mock.Anything, mock.Anything).Return(nil)

	err := server.SendHeartbeat("known")
	require.NoError(t, err)

	err = server.SendHeartbeat("unknown")
	require.NoError(t, err)

	conn.AssertExpectations(t)
}

type TestAddr struct {
	network string
	address string
}

var _ net.Addr = (*TestAddr)(nil)

func (addr *TestAddr) Network() string {
	return addr.network
}
func (addr *TestAddr) String() string {
	return addr.address
}

func TestUpdateOpAmpAgentDetails(t *testing.T) {
	agent := model.Agent{}
	conn := &mocks.Connection{}
	conn.On("RemoteAddr").Return(&TestAddr{network: "tcp", address: "0.0.0.0:0"})

	kv := func(key, value string) *protobufs.KeyValue {
		return &protobufs.KeyValue{Key: key, Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: value}}}
	}

	desc := &protobufs.AgentDescription{
		IdentifyingAttributes: []*protobufs.KeyValue{
			kv("service.instance.id", "instance.id"),
			kv("service.instance.name", "instance.name"),
			kv("service.name", "name"),
			kv("service.version", "version"),
			kv("service.labels", "x=y"),
		},
		NonIdentifyingAttributes: []*protobufs.KeyValue{
			kv("os.arch", "arch"),
			kv("os.details", "details"),
			kv("os.family", "family"),
			kv("host.name", "host"),
			kv("host.mac_address", "mac_address"),
		},
	}

	updateOpAmpAgentDetails(&agent, conn, desc)

	require.Nil(t, agent.DisconnectedAt)
	require.Equal(t, "instance.id", agent.ID)
	require.Equal(t, "name", agent.Type)
	require.Equal(t, "arch", agent.Architecture)
	require.Equal(t, "instance.name", agent.Name)
	require.Equal(t, "host", agent.HostName)
	require.Equal(t, "family", agent.Platform)
	require.Equal(t, "details", agent.OperatingSystem)
	require.Equal(t, model.LabelsFromValidatedMap(map[string]string{
		model.LabelBindPlaneAgentID:      "instance.id",
		model.LabelBindPlaneAgentName:    "instance.name",
		model.LabelBindPlaneAgentVersion: "version",
		model.LabelBindPlaneAgentHost:    "host",
		model.LabelBindPlaneAgentOS:      "family",
		model.LabelBindPlaneAgentArch:    "arch",
		"x":                              "y",
	}), agent.Labels)
	require.Equal(t, "version", agent.Version)
	require.Equal(t, "0.0.0.0:0", agent.RemoteAddress)
	require.Equal(t, "mac_address", agent.MacAddress)
}

// slightly different (no address and labels in non-identifying)
func TestUpdateOpAmpAgentDetails2(t *testing.T) {
	agent := model.Agent{}
	conn := &mocks.Connection{}
	conn.On("RemoteAddr").Return(nil)

	kv := func(key, value string) *protobufs.KeyValue {
		return &protobufs.KeyValue{Key: key, Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: value}}}
	}

	desc := &protobufs.AgentDescription{
		IdentifyingAttributes: []*protobufs.KeyValue{
			kv("service.instance.id", "instance.id"),
			kv("service.instance.name", "instance.name"),
			kv("service.name", "name"),
			kv("service.version", "version"),
		},
		NonIdentifyingAttributes: []*protobufs.KeyValue{
			kv("service.labels", "x=y"),
			kv("os.arch", "arch"),
			kv("os.details", "details"),
			kv("os.family", "family"),
			kv("host.name", "host"),
			kv("host.mac_address", "mac_address"),
		},
	}

	updateOpAmpAgentDetails(&agent, conn, desc)

	require.Nil(t, agent.DisconnectedAt)
	require.Equal(t, "instance.id", agent.ID)
	require.Equal(t, "name", agent.Type)
	require.Equal(t, "arch", agent.Architecture)
	require.Equal(t, "instance.name", agent.Name)
	require.Equal(t, "host", agent.HostName)
	require.Equal(t, "family", agent.Platform)
	require.Equal(t, "details", agent.OperatingSystem)
	require.Equal(t, model.LabelsFromValidatedMap(map[string]string{
		model.LabelBindPlaneAgentID:      "instance.id",
		model.LabelBindPlaneAgentName:    "instance.name",
		model.LabelBindPlaneAgentVersion: "version",
		model.LabelBindPlaneAgentHost:    "host",
		model.LabelBindPlaneAgentOS:      "family",
		model.LabelBindPlaneAgentArch:    "arch",
		"x":                              "y",
	}), agent.Labels)
	require.Equal(t, "version", agent.Version)
	require.Equal(t, "", agent.RemoteAddress)
	require.Equal(t, "mac_address", agent.MacAddress)
}

// bad labels
func TestUpdateOpAmpAgentDetails3(t *testing.T) {
	agent := model.Agent{}
	conn := &mocks.Connection{}
	conn.On("RemoteAddr").Return(nil)

	kv := func(key, value string) *protobufs.KeyValue {
		return &protobufs.KeyValue{Key: key, Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: value}}}
	}

	desc := &protobufs.AgentDescription{
		IdentifyingAttributes: []*protobufs.KeyValue{
			kv("service.instance.id", "instance.id"),
			kv("service.instance.name", "instance.name"),
			kv("service.name", "name"),
			kv("service.version", "version"),
		},
		NonIdentifyingAttributes: []*protobufs.KeyValue{
			kv("service.labels", "=="),
			kv("os.arch", "arch"),
			kv("os.details", "details"),
			kv("os.family", "family"),
			kv("host.name", "host"),
			kv("host.mac_address", "mac_address"),
		},
	}

	updateOpAmpAgentDetails(&agent, conn, desc)

	require.Nil(t, agent.DisconnectedAt)
	require.Equal(t, "instance.id", agent.ID)
	require.Equal(t, "name", agent.Type)
	require.Equal(t, "arch", agent.Architecture)
	require.Equal(t, "instance.name", agent.Name)
	require.Equal(t, "host", agent.HostName)
	require.Equal(t, "family", agent.Platform)
	require.Equal(t, "details", agent.OperatingSystem)
	require.Equal(t, model.LabelsFromValidatedMap(map[string]string{
		model.LabelBindPlaneAgentID:      "instance.id",
		model.LabelBindPlaneAgentName:    "instance.name",
		model.LabelBindPlaneAgentVersion: "version",
		model.LabelBindPlaneAgentHost:    "host",
		model.LabelBindPlaneAgentOS:      "family",
		model.LabelBindPlaneAgentArch:    "arch",
	}), agent.Labels)
	require.Equal(t, "version", agent.Version)
	require.Equal(t, "", agent.RemoteAddress)
	require.Equal(t, "mac_address", agent.MacAddress)
}

func TestServerOnConnecting(t *testing.T) {
	goodKey := "secret"
	badKey := "other"
	noKey := ""
	tests := []struct {
		name          string
		authorization string
		expect        opamp.ConnectionResponse
	}{
		{
			name:          "no key",
			authorization: "",
			expect: opamp.ConnectionResponse{
				Accept:         false,
				HTTPStatusCode: http.StatusUnauthorized,
			},
		},
		{
			name:          "bad key",
			authorization: fmt.Sprintf("Secret-Key %s", badKey),
			expect: opamp.ConnectionResponse{
				Accept:         false,
				HTTPStatusCode: http.StatusUnauthorized,
			},
		},
		{
			name:          "good key",
			authorization: fmt.Sprintf("Secret-Key %s", goodKey),
			expect: opamp.ConnectionResponse{
				Accept:         true,
				HTTPStatusCode: http.StatusOK,
			},
		},
		{
			name:          "bad format",
			authorization: badKey,
			expect: opamp.ConnectionResponse{
				Accept:         false,
				HTTPStatusCode: http.StatusUnauthorized,
			},
		},
		{
			name:          "bad format 2",
			authorization: fmt.Sprintf("Secret-Key: %s", goodKey),
			expect: opamp.ConnectionResponse{
				Accept:         false,
				HTTPStatusCode: http.StatusUnauthorized,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := &mocks.Manager{}
			manager.On("VerifySecretKey", mock.Anything, goodKey).Return(true)
			manager.On("VerifySecretKey", mock.Anything, badKey).Return(false)
			manager.On("VerifySecretKey", mock.Anything, noKey).Return(false)
			server := testServer(manager)
			server.compatibleOpAMPVersions = []string{"v0.2.0"}
			request := &http.Request{
				Header: http.Header{
					"Opamp-Version": []string{"v0.2.0"},
				},
			}
			if test.authorization != "" {
				request.Header["Authorization"] = []string{test.authorization}
			}
			response := server.OnConnecting(request)
			require.Equal(t, test.expect.Accept, response.Accept)
			require.Equal(t, test.expect.HTTPStatusCode, response.HTTPStatusCode)
		})
	}
}

func makeAgentDescription(version string) *protobufs.AgentDescription {
	return &protobufs.AgentDescription{
		IdentifyingAttributes: []*protobufs.KeyValue{
			{
				Key:   "service.version",
				Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: version}},
			},
		},
		NonIdentifyingAttributes: []*protobufs.KeyValue{
			{
				Key:   "service.labels",
				Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "a=b,c=d,configuration=api-test"}},
			},
		},
	}
}

func TestServerOnMessage(t *testing.T) {
	agentID := "a4013625-30f4-489e-a0ca-ef1c97d2ae3f"
	agentCapabilities := protobufs.AgentCapabilities_ReportsEffectiveConfig |
		protobufs.AgentCapabilities_ReportsPackageStatuses |
		protobufs.AgentCapabilities_AcceptsRemoteConfig |
		protobufs.AgentCapabilities_AcceptsPackages |
		protobufs.AgentCapabilities_ReportsStatus

	agentWithConfigurationRaw := &observiq.RawAgentConfiguration{
		Collector: []byte(strings.TrimLeft(`
receivers:
  otlp:
    protocols:
      grpc:
      http:
exporters:
  otlp:
    endpoint: otelcol:4317
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
`, "\n")),
		Manager: []byte(`labels: a=b,c=d,configuration=api-test`),
	}
	agentWithConfigurationRawHash := agentWithConfigurationRaw.Hash()

	// setup initial state
	logger := zap.NewNop()

	testMapStore := store.NewMapStore(context.TODO(), store.Options{
		SessionsSecret:   "supersecret-key",
		MaxEventsToMerge: 1000,
	}, zap.NewNop())
	testManager, err := server.NewManager(
		&common.Server{
			SecretKey: "secret",
		},
		testMapStore,
		logger,
	)
	require.NoError(t, err)

	conn := &testConnection{
		addr: testAddr{"127.0.0.1"},
	}
	server := testServer(testManager)
	testManager.EnableProtocol(server)

	agentDescription := makeAgentDescription("1.0")
	agentDescription2 := makeAgentDescription("2.0")
	agentDescription3 := makeAgentDescription("3.0")

	var sequenceNum uint64
	nextSequenceNum := func() uint64 {
		sequenceNum++
		return sequenceNum
	}

	// these tests are expected to run in order and may have side-affects that are tested in subsequent tests
	tests := []struct {
		name    string
		message *protobufs.AgentToServer
		expect  *protobufs.ServerToAgent
		verify  func(t *testing.T, server *opampServer, result *protobufs.ServerToAgent)
	}{
		{
			name: "status report with no contents, should request details",
			message: &protobufs.AgentToServer{
				SequenceNum:  nextSequenceNum(),
				InstanceUid:  agentID,
				Capabilities: agentCapabilities,
			},
			expect: &protobufs.ServerToAgent{
				InstanceUid:  agentID,
				Capabilities: capabilities,
				Flags:        protobufs.ServerToAgent_ReportFullState,
			},
			verify: func(t *testing.T, server *opampServer, result *protobufs.ServerToAgent) {
				require.ElementsMatch(t, []string{agentID}, server.connections.agentIDs())
				agent, err := server.manager.Agent(context.TODO(), agentID)
				require.NoError(t, err)
				require.Equal(t, "Connected", agent.StatusDisplayText())
			},
		},
		{
			name: "malformed config causes error",
			message: &protobufs.AgentToServer{
				SequenceNum:  nextSequenceNum(),
				InstanceUid:  agentID,
				Capabilities: agentCapabilities,
				EffectiveConfig: &protobufs.EffectiveConfig{
					ConfigMap: &protobufs.AgentConfigMap{
						ConfigMap: map[string]*protobufs.AgentConfigFile{
							observiq.ManagerFilename:   {Body: []byte("[]bad yaml")},
							observiq.CollectorFilename: {Body: []byte("collector")},
							observiq.LoggingFilename:   {Body: []byte("")},
						},
					},
				},
			},
			expect: &protobufs.ServerToAgent{
				InstanceUid:  agentID,
				Capabilities: capabilities,
				Flags:        protobufs.ServerToAgent_ReportFullState,
				ErrorResponse: &protobufs.ServerErrorResponse{
					Type:         protobufs.ServerErrorResponse_Unknown,
					ErrorMessage: "unable to parse the current agent configuration: unable to parse manager config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into observiq.ManagerConfig",
				},
			},
		},
		{
			name: "store valid config",
			message: &protobufs.AgentToServer{
				SequenceNum:  nextSequenceNum(),
				InstanceUid:  agentID,
				Capabilities: agentCapabilities,
				EffectiveConfig: &protobufs.EffectiveConfig{
					ConfigMap: &protobufs.AgentConfigMap{
						ConfigMap: map[string]*protobufs.AgentConfigFile{
							observiq.ManagerFilename:   {Body: []byte("labels: a=b,c=d,configuration=api-test")},
							observiq.CollectorFilename: {Body: []byte("pipelines:")},
							observiq.LoggingFilename:   {Body: []byte("")},
						},
					},
				},
				AgentDescription: agentDescription,
			},
			expect: &protobufs.ServerToAgent{
				InstanceUid:  agentID,
				Capabilities: capabilities,
				Flags:        protobufs.ServerToAgent_ReportFullState,
			},
			verify: func(t *testing.T, server *opampServer, result *protobufs.ServerToAgent) {
				agent, err := server.manager.Agent(context.TODO(), agentID)
				require.NoError(t, err)
				require.Equal(t, "a=b,c=d,configuration=api-test", agent.Labels.Custom().String())
			},
		},
		{
			name: "new description hash requests details",
			message: &protobufs.AgentToServer{
				SequenceNum:  nextSequenceNum(),
				InstanceUid:  agentID,
				Capabilities: agentCapabilities,
			},
			expect: &protobufs.ServerToAgent{
				InstanceUid:  agentID,
				Capabilities: capabilities,
				Flags:        protobufs.ServerToAgent_ReportFullState,
			},
		},
		{
			name: "store new agent description details",
			message: &protobufs.AgentToServer{
				SequenceNum:  nextSequenceNum(),
				InstanceUid:  agentID,
				Capabilities: agentCapabilities,
				EffectiveConfig: &protobufs.EffectiveConfig{
					ConfigMap: &protobufs.AgentConfigMap{
						ConfigMap: map[string]*protobufs.AgentConfigFile{
							observiq.ManagerFilename:   {Body: []byte("labels: a=b,c=d,configuration=api-test")},
							observiq.CollectorFilename: {Body: []byte("pipelines:")},
							observiq.LoggingFilename:   {Body: []byte("")},
						},
					},
				},
				AgentDescription: agentDescription2,
			},
			expect: &protobufs.ServerToAgent{
				InstanceUid:  agentID,
				Capabilities: capabilities,
				Flags:        protobufs.ServerToAgent_ReportFullState,
			},
			verify: func(t *testing.T, server *opampServer, result *protobufs.ServerToAgent) {
				agent, err := server.manager.Agent(context.TODO(), agentID)
				require.NoError(t, err)
				require.Equal(t, "2.0", agent.Version)
			},
		},
		{
			name: "same description does not request details",
			message: &protobufs.AgentToServer{
				SequenceNum:  nextSequenceNum(),
				InstanceUid:  agentID,
				Capabilities: agentCapabilities,
				EffectiveConfig: &protobufs.EffectiveConfig{
					ConfigMap: &protobufs.AgentConfigMap{
						ConfigMap: map[string]*protobufs.AgentConfigFile{
							observiq.ManagerFilename:   {Body: []byte("labels: a=b,c=d,configuration=api-test")},
							observiq.CollectorFilename: {Body: []byte("pipelines:")},
							observiq.LoggingFilename:   {Body: []byte("")},
						},
					},
				},
				AgentDescription: agentDescription,
			},
			expect: &protobufs.ServerToAgent{
				InstanceUid:  agentID,
				Capabilities: capabilities,
				Flags:        protobufs.ServerToAgent_ReportFullState,
			},
			verify: func(t *testing.T, server *opampServer, result *protobufs.ServerToAgent) {
				// gross! inserting a new configuration here and making sure we get it in the next test
				raw := testResource[*model.Configuration](t, "configuration-raw.yaml")
				statuses, err := testMapStore.ApplyResources([]model.Resource{raw})
				require.Equal(t, model.StatusCreated, statuses[0].Status)
				require.NoError(t, err)
			},
		},
		{
			name: "another message with no changes, but new configuration",
			message: &protobufs.AgentToServer{
				SequenceNum:  nextSequenceNum(),
				InstanceUid:  agentID,
				Capabilities: agentCapabilities,
				EffectiveConfig: &protobufs.EffectiveConfig{
					ConfigMap: &protobufs.AgentConfigMap{
						ConfigMap: map[string]*protobufs.AgentConfigFile{
							observiq.ManagerFilename:   {Body: []byte("labels: a=b,c=d,configuration=api-test")},
							observiq.CollectorFilename: {Body: []byte("pipelines:")},
							observiq.LoggingFilename:   {Body: []byte("")},
						},
					},
				},
				AgentDescription: agentDescription,
			},
			expect: &protobufs.ServerToAgent{
				InstanceUid:  agentID,
				Capabilities: capabilities,
				Flags:        protobufs.ServerToAgent_ReportFullState,
				RemoteConfig: &protobufs.AgentRemoteConfig{
					Config: &protobufs.AgentConfigMap{
						ConfigMap: map[string]*protobufs.AgentConfigFile{
							observiq.CollectorFilename: {
								Body: agentWithConfigurationRaw.Collector,
							},
						},
					},
					ConfigHash: agentWithConfigurationRawHash,
				},
			},
		},
		{
			name: "another message with no changes and matching hashes but no config to store",
			message: &protobufs.AgentToServer{
				SequenceNum:      nextSequenceNum(),
				InstanceUid:      agentID,
				Capabilities:     agentCapabilities,
				EffectiveConfig:  &protobufs.EffectiveConfig{},
				AgentDescription: &protobufs.AgentDescription{},
				RemoteConfigStatus: &protobufs.RemoteConfigStatus{
					LastRemoteConfigHash: agentWithConfigurationRawHash,
					Status:               protobufs.RemoteConfigStatus_APPLIED,
				},
			},
			expect: &protobufs.ServerToAgent{
				InstanceUid:  agentID,
				Capabilities: capabilities,
				Flags:        protobufs.ServerToAgent_ReportFullState,
			},
		},
		{
			name: "another message with no changes with configuration to store",
			message: &protobufs.AgentToServer{
				SequenceNum:  nextSequenceNum(),
				InstanceUid:  agentID,
				Capabilities: agentCapabilities,
				EffectiveConfig: &protobufs.EffectiveConfig{
					ConfigMap: &protobufs.AgentConfigMap{
						ConfigMap: map[string]*protobufs.AgentConfigFile{
							observiq.CollectorFilename: {
								Body: agentWithConfigurationRaw.Collector,
							},
							observiq.ManagerFilename: {
								Body: []byte("labels: a=b,c=d,configuration=api-test"),
							},
						},
					},
				},
				AgentDescription:   agentDescription2,
				RemoteConfigStatus: &protobufs.RemoteConfigStatus{},
			},
			expect: &protobufs.ServerToAgent{
				InstanceUid:  agentID,
				Capabilities: capabilities,
				Flags:        protobufs.ServerToAgent_ReportFullState,
			},
		},
		{
			name: "skipped message ignores contents and requests full information",
			message: &protobufs.AgentToServer{
				SequenceNum:  nextSequenceNum() + 1,
				InstanceUid:  agentID,
				Capabilities: agentCapabilities,
				EffectiveConfig: &protobufs.EffectiveConfig{
					ConfigMap: &protobufs.AgentConfigMap{
						ConfigMap: map[string]*protobufs.AgentConfigFile{
							observiq.CollectorFilename: {
								Body: agentWithConfigurationRaw.Collector,
							},
							observiq.ManagerFilename: {
								Body: []byte("labels: a=b,c=d,configuration=api-test"),
							},
						},
					},
				},
				AgentDescription:   agentDescription3,
				RemoteConfigStatus: &protobufs.RemoteConfigStatus{},
			},
			expect: &protobufs.ServerToAgent{
				InstanceUid:  agentID,
				Capabilities: capabilities,
				Flags:        protobufs.ServerToAgent_ReportFullState,
			},
			verify: func(t *testing.T, server *opampServer, result *protobufs.ServerToAgent) {
				agent, err := server.manager.Agent(context.TODO(), agentID)
				require.NoError(t, err)
				require.Equal(t, "2.0", agent.Version)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := server.OnMessage(conn, test.message)

			// compare messages
			require.True(t, proto.Equal(test.expect, result), "protobufs must be equal\nexpect: %v\nactual: %v\n", test.expect, result)
			// test anything additional
			if test.verify != nil {
				test.verify(t, server, result)
			}
		})
	}
}

func TestUpdateAgentStatus(t *testing.T) {
	tests := []struct {
		name                string
		initialStatus       model.AgentStatus
		initialErrorMessage string
		remoteStatus        *protobufs.RemoteConfigStatus
		expectStatus        model.AgentStatus
		expectErrorMessage  string
	}{
		{
			name:          "nil status, preserve Connected",
			initialStatus: model.Connected,
			expectStatus:  model.Connected,
		},
		{
			name:          "nil status, set Connected",
			initialStatus: model.Disconnected,
			expectStatus:  model.Connected,
		},
		{
			name:          "nil status, preserve Error",
			initialStatus: model.Error,
			expectStatus:  model.Error,
		},
		{
			name:                "UNSET status, preserve Error",
			initialStatus:       model.Error,
			initialErrorMessage: "error",
			remoteStatus: &protobufs.RemoteConfigStatus{
				Status: protobufs.RemoteConfigStatus_UNSET,
			},
			expectStatus:       model.Error,
			expectErrorMessage: "error",
		},
		{
			name:                "FAILED status, set Error",
			initialStatus:       model.Connected,
			initialErrorMessage: "",
			remoteStatus: &protobufs.RemoteConfigStatus{
				Status:       protobufs.RemoteConfigStatus_FAILED,
				ErrorMessage: "error",
			},
			expectStatus:       model.Error,
			expectErrorMessage: "error",
		},
		{
			name:                "FAILED status, change Error",
			initialStatus:       model.Error,
			initialErrorMessage: "old error",
			remoteStatus: &protobufs.RemoteConfigStatus{
				Status:       protobufs.RemoteConfigStatus_FAILED,
				ErrorMessage: "new error",
			},
			expectStatus:       model.Error,
			expectErrorMessage: "new error",
		},
		{
			name:                "APPLIED status, clear Error",
			initialStatus:       model.Error,
			initialErrorMessage: "error",
			remoteStatus: &protobufs.RemoteConfigStatus{
				Status: protobufs.RemoteConfigStatus_APPLIED,
			},
			expectStatus:       model.Connected,
			expectErrorMessage: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agent := &model.Agent{
				Status:       test.initialStatus,
				ErrorMessage: test.initialErrorMessage,
			}
			updateAgentStatus(zap.NewNop(), agent, test.remoteStatus)
			require.Equal(t, test.expectStatus, agent.Status)
			require.Equal(t, test.expectErrorMessage, agent.ErrorMessage)
		})
	}
}

func TestOnConnectingOpAMPCompatibility(t *testing.T) {
	tests := []struct {
		name    string
		request http.Request
		expect  opamp.ConnectionResponse
	}{
		{
			name: "no version",
			request: http.Request{
				Header: http.Header{
					"Authorization":         []string{"Secret-Key a0f1db77-818a-4f1a-81a3-7b6a9613ef41"},
					"Connection":            []string{"Upgrade"},
					"Sec-Websocket-Key":     []string{"xx=="},
					"Sec-Websocket-Version": []string{"13"},
					"Upgrade":               []string{"websocket"},
					"User-Agent":            []string{"observiq-otel-collector/v1.2.0"}},
			},
			expect: opamp.ConnectionResponse{
				Accept:         false,
				HTTPStatusCode: http.StatusUpgradeRequired,
				HTTPResponseHeader: map[string]string{
					"Upgrade": "OpAMP/v0.2.0",
				},
			},
		},
		{
			name: "ok version",
			request: http.Request{
				Header: http.Header{
					"Agent-Hostname":        []string{"arm.localdomain"},
					"Agent-Id":              []string{"4ec02b0f-3cb7-498d-9172-bfaa28718ee8"},
					"Agent-Version":         []string{"v1.2.0"},
					"Authorization":         []string{"Secret-Key a0f1db77-818a-4f1a-81a3-7b6a9613ef41"},
					"Connection":            []string{"Upgrade"},
					"Opamp-Version":         []string{"v0.2.0"},
					"Sec-Websocket-Key":     []string{"xx=="},
					"Sec-Websocket-Version": []string{"13"},
					"Upgrade":               []string{"websocket"},
					"User-Agent":            []string{"observiq-otel-collector/v1.2.0"},
				},
			},
			expect: opamp.ConnectionResponse{
				Accept:         true,
				HTTPStatusCode: http.StatusOK,
			},
		},
		{
			name: "ok version, bad secret key",
			request: http.Request{
				Header: http.Header{
					"Agent-Hostname":        []string{"arm.localdomain"},
					"Agent-Id":              []string{"4ec02b0f-3cb7-498d-9172-bfaa28718ee8"},
					"Agent-Version":         []string{"v1.2.0"},
					"Authorization":         []string{"Secret-Key 6afd5cf2-2c3f-44f7-a2f6-6fc310ad69b8"},
					"Connection":            []string{"Upgrade"},
					"Opamp-Version":         []string{"v0.2.0"},
					"Sec-Websocket-Key":     []string{"xx=="},
					"Sec-Websocket-Version": []string{"13"},
					"Upgrade":               []string{"websocket"},
					"User-Agent":            []string{"observiq-otel-collector/v1.2.0"},
				},
			},
			expect: opamp.ConnectionResponse{
				Accept:         false,
				HTTPStatusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "future version",
			request: http.Request{
				Header: http.Header{
					"Agent-Hostname":        []string{"arm.localdomain"},
					"Agent-Id":              []string{"4ec02b0f-3cb7-498d-9172-bfaa28718ee8"},
					"Agent-Version":         []string{"v1.2.0"},
					"Authorization":         []string{"Secret-Key a0f1db77-818a-4f1a-81a3-7b6a9613ef41"},
					"Connection":            []string{"Upgrade"},
					"Opamp-Version":         []string{"v0.3.0"},
					"Sec-Websocket-Key":     []string{"xx=="},
					"Sec-Websocket-Version": []string{"13"},
					"Upgrade":               []string{"websocket"},
					"User-Agent":            []string{"observiq-otel-collector/v1.2.0"},
				},
			},
			expect: opamp.ConnectionResponse{
				Accept:         false,
				HTTPStatusCode: http.StatusUpgradeRequired,
				HTTPResponseHeader: map[string]string{
					"Upgrade": "OpAMP/v0.2.0",
				},
			},
		},
	}

	for _, test := range tests {
		testManager, err := server.NewManager(&common.Server{SecretKey: "a0f1db77-818a-4f1a-81a3-7b6a9613ef41"}, nil, zap.NewNop())
		require.NoError(t, err)
		testServer := newServer(testManager, zap.NewNop())
		testServer.compatibleOpAMPVersions = []string{"v0.2.0"}
		t.Run(test.name, func(t *testing.T) {
			response := testServer.OnConnecting(&test.request)
			require.Equal(t, test.expect.Accept, response.Accept)
			require.Equal(t, test.expect.HTTPStatusCode, response.HTTPStatusCode)
			require.Equal(t, test.expect.HTTPResponseHeader, response.HTTPResponseHeader)
		})
	}
}
