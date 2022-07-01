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

package get

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/model"
)

var tableOutput = "table"
var jsonOutput = "json"
var yamlOutput = "yaml"

func setupBindPlane(buffer *bytes.Buffer) *cli.BindPlane {
	bindplane := cli.NewBindPlane(common.InitConfig(""), buffer)
	bindplane.SetClient(&mockClient{})
	return bindplane
}

type mockClient struct {
	client.BindPlane
	mock.Mock
}

// Agents TODO(doc)
func (c *mockClient) Agents(ctx context.Context, options ...client.QueryOption) ([]*model.Agent, error) {
	return []*model.Agent{
		{
			ID:              "1",
			Architecture:    "amd64",
			HostName:        "local",
			Platform:        "linux",
			SecretKey:       "secret",
			Version:         "1.0.0",
			Name:            "Agent 1",
			Home:            "/stanza",
			OperatingSystem: "Ubuntu 20.10",
			MacAddress:      "00:00:ac:00:00:00",
			Type:            "stanza",
			Status:          model.Connected,
		},
		{
			ID:      "2",
			Name:    "Agent 2",
			Version: "1.0.0",
			Status:  model.Disconnected,
		},
	}, nil
}

// Agent TODO(doc)
func (c *mockClient) Agent(ctx context.Context, id string) (*model.Agent, error) {
	agents, _ := c.Agents(ctx)
	if id == agents[0].ID {
		return agents[0], nil
	}

	return nil, nil
}

func executeAndAssertOutput(t *testing.T, cmd *cobra.Command, buffer *bytes.Buffer, expected string) {
	executeErr := cmd.Execute()
	require.NoError(t, executeErr, "error while executing command")

	out, readErr := ioutil.ReadAll(buffer)
	require.NoError(t, readErr, "error while reading byte array")

	require.Equal(t, expected, string(out))
}
