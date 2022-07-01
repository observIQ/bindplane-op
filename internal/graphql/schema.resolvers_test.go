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

package graphql

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/observiq/bindplane/common"
	model1 "github.com/observiq/bindplane/internal/graphql/model"
	"github.com/observiq/bindplane/internal/server"
	"github.com/observiq/bindplane/internal/store"
	"github.com/observiq/bindplane/model"
)

func addAgent(s store.Store, agent *model.Agent) (*model.Agent, error) {
	_, err := s.UpsertAgent(context.TODO(), agent.ID, func(a *model.Agent) {
		*a = *agent
	})
	return agent, err
}

func TestQueryResolvers(t *testing.T) {
	mapstore := store.NewMapStore(zap.NewNop(), "super-secret-key")
	bindplane, err := server.NewBindPlane(&common.Server{}, zaptest.NewLogger(t), mapstore, nil)
	require.NoError(t, err)

	srv := newHandler(bindplane)
	c := client.New(srv)

	s := bindplane.Store()

	t.Run("agents returns all Agents in the store", func(t *testing.T) {
		s.Clear()

		var resp map[string]model1.Agents
		var err error

		// Shouldn't get any Agents before adding to the store
		err = c.Post(`query TestQuery { agents(selector: "") { agents { id } } }`, &resp)
		require.NoError(t, err)
		require.Len(t, resp["agents"].Agents, 0)

		xy, err := model.LabelsFromSelector("x=y")
		require.NoError(t, err)

		addAgent(s, &model.Agent{ID: "1", Name: "Fake Agent 1", Labels: xy})
		addAgent(s, &model.Agent{ID: "2", Name: "Fake Agent 2"})

		// Should get the two Agents back that we added
		err = c.Post(`query TestQuery { agents(selector: "") { agents { id } } }`, &resp)
		require.NoError(t, err)
		require.Len(t, resp["agents"].Agents, 2)

		// Should get the one Agent back that matches the selector
		err = c.Post(`query TestQuery { agents(selector: "x=y") { agents { id } } }`, &resp)
		require.NoError(t, err)
		require.Len(t, resp["agents"].Agents, 1)
	})

	t.Run("agent loads a specific Agent by ID", func(t *testing.T) {
		s.Clear()

		var resp map[string]*model.Agent
		var err error

		addAgent(s, &model.Agent{ID: "1", Name: "Fake Agent 1"})
		agent, err := addAgent(s, &model.Agent{ID: "2", Name: "Fake Agent 2"})
		require.NoError(t, err)

		err = c.Post("query TestQuery($id: ID!) { agent(id: $id) { id } }", &resp, client.Var("id", "2"))
		require.NoError(t, err)
		require.Equal(t, resp["agent"].ID, agent.ID)
	})
}

func TestConfigForAgent(t *testing.T) {
	mapstore := store.NewMapStore(zap.NewNop(), "super-secret-key")
	bindplane, err := server.NewBindPlane(&common.Server{}, zaptest.NewLogger(t), mapstore, nil)
	require.NoError(t, err)

	srv := newHandler(bindplane)
	c := client.New(srv)

	store := bindplane.Store()

	// SETUP
	labels := map[string]string{"env": "test", "app": "bindplane"}
	agent1Labels := model.Labels{Set: labels}

	otherLabels := map[string]string{"foo": "bar"}
	agent2labels := model.Labels{Set: otherLabels}

	addAgent(store, &model.Agent{ID: "1", Labels: agent1Labels})
	addAgent(store, &model.Agent{ID: "2", Labels: agent2labels})

	configLabels, _ := model.LabelsFromMap(map[string]string{"platform": "linux"})

	config := &model.Configuration{
		Spec: model.ConfigurationSpec{
			Raw:      "raw:",
			Selector: model.AgentSelector{MatchLabels: labels},
		},
		ResourceMeta: model.ResourceMeta{
			APIVersion: "",
			Kind:       "Configuration",
			Metadata: model.Metadata{
				Name:        "config",
				ID:          "config-123",
				Description: "should be used by agent 1",
				Labels:      configLabels,
			},
		},
	}

	_, err = bindplane.Store().ApplyResources([]model.Resource{config})
	require.NoError(t, err)

	resp := &struct {
		Agents struct {
			Agents []struct {
				ID                    string
				Name                  string
				ConfigurationResource *struct {
					Metadata struct {
						Name string
					}
				}
			}
		}
	}{}

	agentsQuery := `
	query TestAgents {
		agents {
			agents {
				id
				name
				configurationResource {
					metadata {
						name
					}
				}
			}
		}
	}
`

	err = c.Post(agentsQuery, &resp)
	require.NoError(t, err)

	for _, agent := range resp.Agents.Agents {
		switch agent.ID {
		case "1":
			require.Equal(t, "config", agent.ConfigurationResource.Metadata.Name)
		case "2":
			require.Nil(t, agent.ConfigurationResource)
		}
	}
}
