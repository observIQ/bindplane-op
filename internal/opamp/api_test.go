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
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/observiq/bindplane/common"
	"github.com/observiq/bindplane/internal/server"
	"github.com/observiq/bindplane/internal/server/mocks"
	"github.com/observiq/bindplane/model"
	"github.com/observiq/bindplane/model/observiq"
	"github.com/open-telemetry/opamp-go/protobufs"
	opamp "github.com/open-telemetry/opamp-go/server/types"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testServer(manager server.Manager) *opampServer {
	return newServer(manager, zap.NewNop())
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

func TestServerOnMessage(t *testing.T) {
	agentID := "a4013625-30f4-489e-a0ca-ef1c97d2ae3f"
	tests := []struct {
		name    string
		message *protobufs.AgentToServer
		expect  *protobufs.ServerToAgent
		verify  func(t *testing.T, server *opampServer, result *protobufs.ServerToAgent)
	}{
		{
			name: "status report with no contents, should request effective config",
			message: &protobufs.AgentToServer{
				InstanceUid: agentID,
			},
			expect: &protobufs.ServerToAgent{
				InstanceUid:  agentID,
				Capabilities: capabilities,
				Flags:        protobufs.ServerToAgent_ReportFullState,
			},
			verify: func(t *testing.T, server *opampServer, result *protobufs.ServerToAgent) {
				require.ElementsMatch(t, []string{agentID}, server.connections.agentIDs())
			},
		},
		{
			name: "malformed config causes error",
			message: &protobufs.AgentToServer{
				InstanceUid: agentID,
				EffectiveConfig: &protobufs.EffectiveConfig{
					ConfigMap: &protobufs.AgentConfigMap{
						ConfigMap: map[string]*protobufs.AgentConfigFile{
							observiq.ManagerFilename:   {Body: []byte("[]bad yaml")},
							observiq.CollectorFilename: {Body: []byte("collector")},
							observiq.LoggingFilename:   {Body: []byte("logging")},
						},
					},
				},
			},
			expect: &protobufs.ServerToAgent{
				InstanceUid:  agentID,
				Capabilities: capabilities,
				ErrorResponse: &protobufs.ServerErrorResponse{
					Type:         protobufs.ServerErrorResponse_Unknown,
					ErrorMessage: "unable to parse the current agent configuration: unable to parse manager config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into observiq.ManagerConfig",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := &mocks.Manager{}
			conn := &mocks.Connection{}
			server := testServer(manager)

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
			manager := &mocks.Manager{}
			server := testServer(manager)

			agent := &model.Agent{
				Status:       test.initialStatus,
				ErrorMessage: test.initialErrorMessage,
			}
			server.updateAgentStatus(agent, test.remoteStatus)
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
