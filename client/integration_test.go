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

//go:build integration
// +build integration

package client

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

func defaultServerEnv() map[string]string {
	return map[string]string{
		"BINDPLANE_CONFIG_USERNAME":        "oiq",
		"BINDPLANE_CONFIG_PASSWORD":        "password",
		"BINDPLANE_CONFIG_SESSIONS_SECRET": uuid.NewString(),
		"BINDPLANE_CONFIG_LOG_OUTPUT":      "stdout",
	}
}

func bindplaneContainer(t *testing.T, env map[string]string) (testcontainers.Container, int, error) {
	// Detect an open port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, 0, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	dir, err := os.Getwd()
	if err != nil {
		return nil, 0, err
	}

	mounts := map[string]string{
		"/tmp": path.Join(dir, "testdata"),
	}

	image := fmt.Sprintf("bindplane-%s:latest", runtime.GOARCH)

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        image,
		Env:          env,
		BindMounts:   mounts,
		ExposedPorts: []string{fmt.Sprintf("%d:%d", port, 3001)},
		WaitingFor:   wait.ForListeningPort("3001"),
	}

	require.NoError(t, req.Validate())

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	time.Sleep(time.Second * 3)

	return container, port, nil
}

// TestIntegrationHttp tests a simple client.Agents call
// using http / plain text.
func TestIntegrationHttp(t *testing.T) {
	env := defaultServerEnv()

	container, port, err := bindplaneContainer(t, env)
	if err != nil {
		require.NoError(t, err, "failed to build test container")
		return
	}
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
		time.Sleep(time.Second * 1)
	}()

	hostname, err := container.Host(context.Background())
	require.NoError(t, err, "failed to get container hostname")

	endpoint := url.URL{
		Host:   fmt.Sprintf("%s:%d", hostname, port),
		Scheme: "http",
	}

	defaultClientConfig := common.Client{
		Common: common.Common{
			Username:  "oiq",
			Password:  "password",
			ServerURL: endpoint.String(),
		},
	}

	client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
	require.NoError(t, err, "failed to create client config: %v", err)
	require.NotNil(t, client)

	_, err = client.Agents(context.Background())
	require.NoError(t, err)
}

// TestIntegrationHttps tests a simple client.Agents call
// using https / server side tls.
func TestIntegrationHttps(t *testing.T) {
	env := defaultServerEnv()
	env["BINDPLANE_CONFIG_TLS_CERT"] = "/tmp/bindplane.crt"
	env["BINDPLANE_CONFIG_TLS_KEY"] = "/tmp/bindplane.key"

	container, port, err := bindplaneContainer(t, env)
	if err != nil {
		require.NoError(t, err, "failed to build test container")
		return
	}
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
		time.Sleep(time.Second * 1)
	}()

	hostname, err := container.Host(context.Background())
	require.NoError(t, err, "failed to get container hostname")

	endpoint := url.URL{
		Host:   fmt.Sprintf("%s:%d", hostname, port),
		Scheme: "https",
	}

	defaultClientConfig := common.Client{
		Common: common.Common{
			Username:  "oiq",
			Password:  "password",
			ServerURL: endpoint.String(),
			TLSConfig: common.TLSConfig{
				CertificateAuthority: []string{
					"testdata/bindplane-ca.crt",
				},
			},
		},
	}

	client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
	require.NoError(t, err, "failed to create client config: %v", err)
	require.NotNil(t, client)

	_, err = client.Agents(context.Background())
	require.NoError(t, err)
}

// TestIntegrationHttpsMutualTLS tests all client api calls using
// client / server authentication (mutual TLS).
func TestIntegrationHttpsMutualTLS(t *testing.T) {
	env := defaultServerEnv()
	env["BINDPLANE_CONFIG_TLS_CERT"] = "/tmp/bindplane.crt"
	env["BINDPLANE_CONFIG_TLS_KEY"] = "/tmp/bindplane.key"
	env["BINDPLANE_CONFIG_TLS_CA"] = "/tmp/bindplane-ca.crt"

	container, port, err := bindplaneContainer(t, env)
	if err != nil {
		require.NoError(t, err, "failed to build test container")
		return
	}
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
		time.Sleep(time.Second * 1)
	}()

	hostname, err := container.Host(context.Background())
	require.NoError(t, err, "failed to get container hostname")

	endpoint := url.URL{
		Host:   fmt.Sprintf("%s:%d", hostname, port),
		Scheme: "https",
	}

	// Base config can be copied and modified by the test case
	defaultClientConfig := common.Client{
		Common: common.Common{
			Username:  "oiq",
			Password:  "password",
			ServerURL: endpoint.String(),
			TLSConfig: common.TLSConfig{
				Certificate: "testdata/bindplane.crt",
				PrivateKey:  "testdata/bindplane.key",
				CertificateAuthority: []string{
					"testdata/bindplane-ca.crt",
				},
			},
		},
	}

	cases := []struct {
		name      string
		apiCall   func() error
		expectErr string
	}{
		{
			"Agents",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Agents(context.Background())
				return err
			},
			"",
		},
		{
			"Agents Request Error",
			func() error {
				c := defaultClientConfig
				c.ServerURL = "xxx://invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Agents(context.Background())
				return err
			},
			"unsupported protocol scheme",
		},
		{
			"Agents Internal Server Error",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				opts := WithLimit(-1)
				_, err = client.Agents(context.Background(), opts)
				return err
			},
			"unable to get agents, got 500 Internal Server Error",
		},
		{
			"Agents Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Agents(context.Background())
				return err
			},
			"401 Unauthorized",
		},
		{
			"Agent",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Agent(context.Background(), "agent")
				return err
			},
			"unable to get agents, got 404 Not Found",
		},
		{
			"Agent Request Error",
			func() error {
				c := defaultClientConfig
				c.ServerURL = "xxx://invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Agent(context.Background(), "")
				return err
			},
			"unsupported protocol scheme",
		},
		{
			"Agent Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Agent(context.Background(), "agent")
				return err
			},
			"401 Unauthorized",
		},
		{
			"DeleteAgents",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.DeleteAgents(context.Background(), []string{"agent"})
				return err
			},
			"",
		},
		{
			"DeleteAgents",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.DeleteAgents(context.Background(), []string{"agent"})
				return err
			},
			"401 Unauthorized",
		},
		{
			"Configurations",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Configurations(context.Background())
				return err
			},
			"",
		},
		{
			"Configurations Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Configurations(context.Background())
				return err
			},
			"401 Unauthorized",
		},
		{
			"Configuration",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Configuration(context.Background(), "config")
				return err
			},
			"unable to get /configurations/config, got 404 Not Found",
		},
		{
			"Configuration Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Configuration(context.Background(), "config")
				return err
			},
			"401 Unauthorized",
		},
		{
			"DeleteConfiguration",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				return client.DeleteConfiguration(context.Background(), "config")
			},
			"/configurations/config not found",
		},
		{
			"DeleteConfiguration Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				return client.DeleteConfiguration(context.Background(), "config")
			},
			"401 Unauthorized",
		},
		{
			"RawConfiguration",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.RawConfiguration(context.Background(), "config")
				return err
			},
			"unable to get /configurations/config, got 404 Not Found",
		},
		{
			"RawConfiguration Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.RawConfiguration(context.Background(), "config")
				return err
			},
			"401 Unauthorized",
		},
		{
			"Source",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Source(context.Background(), "source")
				return err
			},
			"unable to get /sources/source, got 404 Not Found",
		},
		{
			"Source Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Source(context.Background(), "source")
				return err
			},
			"401 Unauthorized",
		},
		{
			"Sources",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Sources(context.Background())
				return err
			},
			"",
		},
		{
			"Sources Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Sources(context.Background())
				return err
			},
			"401 Unauthorized",
		},
		{
			"DeleteSource",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				return client.DeleteSource(context.Background(), "source")
			},
			"/sources/source not found",
		},
		{
			"DeleteSource Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				err = client.DeleteSource(context.Background(), "source")
				return err
			},
			"401 Unauthorized",
		},
		{
			"SourceTypes",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.SourceTypes(context.Background())
				return err
			},
			"",
		},
		{
			"SourceTypes Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.SourceTypes(context.Background())
				return err
			},
			"401 Unauthorized",
		},
		{
			"SourceType",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.SourceType(context.Background(), "source-type")
				return err
			},
			"unable to get /source-types/source-type, got 404 Not Found",
		},
		{
			"SourceType Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.SourceType(context.Background(), "source-type")
				return err
			},
			"401 Unauthorized",
		},
		{
			"DeleteSourceType",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				return client.DeleteSourceType(context.Background(), "source-type")
			},
			"/source-types/source-type not found",
		},
		{
			"DeleteSourceType Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				err = client.DeleteSourceType(context.Background(), "source-type")
				return err
			},
			"401 Unauthorized",
		},
		{
			"Destinations",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Destinations(context.Background())
				return err
			},
			"",
		},
		{
			"Destinations Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Destinations(context.Background())
				return err
			},
			"401 Unauthorized",
		},
		{
			"Destination",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Destination(context.Background(), "dest")
				return err
			},
			"unable to get /destinations/dest, got 404 Not Found",
		},
		{
			"Destination Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Destination(context.Background(), "dest")
				return err
			},
			"401 Unauthorized",
		},
		{
			"DeleteDestination",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				return client.DeleteDestination(context.Background(), "dest")
			},
			"/destinations/dest not found",
		},
		{
			"DeleteDestination Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				return client.DeleteDestination(context.Background(), "dest")
			},
			"401 Unauthorized",
		},
		{
			"DestinationTypes",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.DestinationTypes(context.Background())
				return err
			},
			"",
		},
		{
			"DestinationTypes Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.DestinationTypes(context.Background())
				return err
			},
			"401 Unauthorized",
		},
		{
			"DestinationType",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.DestinationType(context.Background(), "dest-type")
				return err
			},
			"unable to get /destination-types/dest-type, got 404 Not Found",
		},
		{
			"DestinationType Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.DestinationType(context.Background(), "dest-type")
				return err
			},
			"401 Unauthorized",
		},
		{
			"DeleteDestinationType",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				return client.DeleteDestinationType(context.Background(), "dest-type")
			},
			"/destination-types/dest-type not found",
		},
		{
			"DeleteDestinationType Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				return client.DeleteDestinationType(context.Background(), "dest-type")
			},
			"401 Unauthorized",
		},
		{
			"Apply",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Apply(context.Background(), nil)
				return err
			},
			"",
		},
		{
			"Apply Request Error",
			func() error {
				c := defaultClientConfig
				c.ServerURL = "xxx://invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Apply(context.Background(), nil)
				return err
			},
			"unsupported protocol scheme",
		},
		{
			"Apply Invalid Request Payload",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				resource := model.Configuration{}
				r := &model.AnyResource{
					ResourceMeta: resource.ResourceMeta,
				}
				_, err = client.Apply(context.Background(), []*model.AnyResource{r})
				return err
			},
			"unable to apply resources, got 400 Bad Request",
		},
		{
			"Apply Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Apply(context.Background(), nil)
				return err
			},
			"401 Unauthorized",
		},
		{
			"Delete",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Delete(context.Background(), nil)
				return err
			},
			"",
		},
		{
			"Delete Request Error",
			func() error {
				c := defaultClientConfig
				c.ServerURL = "invalid://invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Delete(context.Background(), nil)
				return err
			},
			"unsupported protocol scheme",
		},
		{
			"Delete Invalid Request Payload",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				resource := model.Configuration{}
				r := &model.AnyResource{
					ResourceMeta: resource.ResourceMeta,
				}
				_, err = client.Delete(context.Background(), []*model.AnyResource{r})
				return err
			},
			"bad request",
		},
		{
			"Delete Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Delete(context.Background(), nil)
				return err
			},
			"401 Unauthorized",
		},
		{
			"Delete_bad_request",
			func() error {
				r := model.AnyResource{
					ResourceMeta: model.ResourceMeta{
						APIVersion: "invalid",
					},
				}
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Delete(context.Background(), []*model.AnyResource{&r})
				return err
			},
			"bad request",
		},
		{
			"Version",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Version(context.Background())
				return err
			},
			"",
		},
		{
			"Version Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.Version(context.Background())
				return err
			},
			"401 Unauthorized",
		},
		{
			"AgentInstallCommand",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.AgentInstallCommand(context.Background(), AgentInstallOptions{})
				return err
			},
			"",
		},
		{
			"AgentInstallCommand Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.AgentInstallCommand(context.Background(), AgentInstallOptions{})
				return err
			},
			"401 Unauthorized",
		},
		{
			"AgentUpdate",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				return client.AgentUpdate(context.Background(), "id", "v1.3.0")
			},
			"",
		},
		{
			"AgentLabels",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.AgentLabels(context.Background(), "id")
				return err
			},
			"unable to get agent labels, got 404 Not Found",
		},
		{
			"AgentLabels Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.AgentLabels(context.Background(), "id")
				return err
			},
			"401 Unauthorized",
		},
		{
			"ApplyAgentLabels_not_found",
			func() error {
				l, err := model.LabelsFromMap(map[string]string{"a": "b"})
				if err != nil {
					return err
				}
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.ApplyAgentLabels(context.Background(), "id", &l, true)
				return err
			},
			"unable to apply labels, got 404 Not Found",
		},
		{
			"ApplyAgentLabels_bad_request",
			func() error {
				client, err := NewBindPlane(&defaultClientConfig, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.ApplyAgentLabels(context.Background(), "id", &model.Labels{}, true)
				return err
			},
			"unable to apply labels, got 400 Bad Request",
		},
		{
			"ApplyAgentLabels Unauthorized",
			func() error {
				c := defaultClientConfig
				c.Username = "invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.ApplyAgentLabels(context.Background(), "id", &model.Labels{}, true)
				return err
			},
			"401 Unauthorized",
		},
		{
			"ApplyAgentLabels Request Error",
			func() error {
				c := defaultClientConfig
				c.ServerURL = "invalid://invalid"
				client, err := NewBindPlane(&c, zap.NewNop())
				if err != nil {
					return err
				}
				_, err = client.ApplyAgentLabels(context.Background(), "id", &model.Labels{}, true)
				return err
			},
			"unsupported protocol scheme",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.apiCall()
			if tc.expectErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
