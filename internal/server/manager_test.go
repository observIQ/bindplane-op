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

package server

import (
	"context"
	"testing"

	"github.com/observiq/bindplane-op/internal/store"
	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// what we want to test:
//
// add some agents with labels
// add some configurations with matchLabels
// * send labels changes (and associated configuration) to the agent
// * delete a configuration in use by an agent, revert to sample config
// * agent report an unknown configuration, no change

var (
	logger       = zap.NewNop()
	testMapstore = store.NewMapStore(context.Background(), store.Options{
		SessionsSecret:   "super-secret-key",
		MaxEventsToMerge: 1,
	}, logger)
	testProtocol = &mockProtocol{}
	testManager  = &manager{
		store:     testMapstore,
		logger:    logger,
		protocols: []Protocol{testProtocol},
	}
)

func makeTestAgent(agentID string) *model.Agent {
	agent, err := testMapstore.UpsertAgent(context.TODO(), agentID, func(agent *model.Agent) {})
	if err != nil {
		panic(err)
	}
	return agent
}

func makeTestAgentWithLabels(agentID string, labelSelector string) *model.Agent {
	agent, err := testMapstore.UpsertAgent(context.TODO(), agentID, func(agent *model.Agent) {
		// we are passing this labelSelector in during tests, so we can avoid an error here
		labels, _ := model.LabelsFromSelector(labelSelector)
		agent.Labels = labels
	})
	if err != nil {
		panic(err)
	}
	return agent
}

func makeTestConfiguration(t *testing.T, name, labelSelector, raw string) *model.Configuration {
	labels, err := model.LabelsFromSelector(labelSelector)
	require.NoError(t, err)
	return &model.Configuration{
		ResourceMeta: model.ResourceMeta{
			Kind: model.KindConfiguration,
			Metadata: model.Metadata{
				Name: name,
			},
		},
		Spec: model.ConfigurationSpec{
			Raw: raw,
			Selector: model.AgentSelector{
				MatchLabels: model.MatchLabels(labels.Set),
			},
		},
	}
}

func managerTestReset() {
	testMapstore.Clear()
	testProtocol = &mockProtocol{}
	testManager.protocols = []Protocol{testProtocol}
}

func TestHandleUpdatesEmpty(t *testing.T) {
	managerTestReset()
	updates := store.NewUpdates()
	testManager.handleUpdates(updates)
	testProtocol.AssertExpectations(t)
}

func TestHandleUpdatesAgents(t *testing.T) {
	managerTestReset()
	testAgent := makeTestAgent("A")
	updates := store.NewUpdates()
	updates.IncludeAgent(testAgent, store.EventTypeLabel)
	labels := model.MakeLabels()
	agentUpdates := &AgentUpdates{
		Labels: &labels,
	}
	testProtocol.
		On("Connected", testAgent.ID).Return(true).
		On("UpdateAgent", mock.Anything, testAgent, agentUpdates).Return(nil)

	testManager.handleUpdates(updates)

	testProtocol.AssertExpectations(t)
}

func TestHandleUpdatesTwoAgents(t *testing.T) {
	managerTestReset()
	testAgentA := makeTestAgentWithLabels("A", "x=y")
	testAgentB := makeTestAgentWithLabels("B", "x=y")

	updates := store.NewUpdates()
	updates.IncludeAgent(testAgentA, store.EventTypeLabel)
	updates.IncludeAgent(testAgentB, store.EventTypeLabel)

	labels, err := model.LabelsFromSelector("x=y")
	require.NoError(t, err)
	agentUpdates := &AgentUpdates{
		Labels: &labels,
	}
	testProtocol.
		On("Connected", testAgentA.ID).Return(true).
		On("Connected", testAgentB.ID).Return(true).
		On("UpdateAgent", mock.Anything, testAgentA, agentUpdates).Return(nil).
		On("UpdateAgent", mock.Anything, testAgentB, agentUpdates).Return(nil)

	testManager.handleUpdates(updates)

	testProtocol.AssertExpectations(t)
}

func TestHandleUpdatesAgentLabels(t *testing.T) {
	managerTestReset()
	testAgentA := makeTestAgentWithLabels("A", "w=x,y=z")
	updates := store.NewUpdates()
	updates.IncludeAgent(testAgentA, store.EventTypeLabel)
	labels, err := model.LabelsFromMap(map[string]string{
		"w": "x",
		"y": "z",
	})
	require.NoError(t, err)

	agentUpdates := &AgentUpdates{
		Labels: &labels,
	}
	testProtocol.
		On("Connected", testAgentA.ID).Return(true).
		On("UpdateAgent", mock.Anything, testAgentA, agentUpdates).Return(nil)

	testManager.handleUpdates(updates)

	testProtocol.AssertExpectations(t)
}

func TestHandleUpdatesNewConfiguration(t *testing.T) {
	managerTestReset()
	testAgentA := makeTestAgentWithLabels("A", "configuration=test")
	makeTestAgentWithLabels("B", "configuration=other")
	configuration := makeTestConfiguration(t, "test", "configuration=test", "raw:")
	_, err := testMapstore.ApplyResources([]model.Resource{configuration})
	require.NoError(t, err)

	updates := store.NewUpdates()
	updates.Configurations.Include(configuration, store.EventTypeUpdate)

	agentUpdates := &AgentUpdates{
		Configuration: configuration,
	}

	testProtocol.
		On("Connected", testAgentA.ID).Return(true).
		On("UpdateAgent", mock.Anything, testAgentA, agentUpdates).Return(nil)

	testManager.handleUpdates(updates)

	testProtocol.AssertExpectations(t)
}

func TestHandleUpdatesNewConfigurationAndLabels(t *testing.T) {
	managerTestReset()
	testAgentA := makeTestAgentWithLabels("A", "configuration=test")
	testAgentB := makeTestAgentWithLabels("B", "configuration=other")
	testAgentC := makeTestAgentWithLabels("C", "configuration=test") // not connected
	configuration := makeTestConfiguration(t, "test", "configuration=test", "raw:")
	_, err := testMapstore.ApplyResources([]model.Resource{configuration})
	require.NoError(t, err)

	testAgentB2, err := testMapstore.UpsertAgent(context.TODO(), "B", func(current *model.Agent) {
		l, err := model.LabelsFromSelector("configuration=test")
		require.NoError(t, err)
		current.Labels = l
	})
	require.NoError(t, err)

	updates := store.NewUpdates()
	updates.Configurations.Include(configuration, store.EventTypeUpdate)
	updates.IncludeAgent(testAgentB2, store.EventTypeLabel)
	updates.IncludeAgent(testAgentC, store.EventTypeLabel)

	labels, err := model.LabelsFromMap(map[string]string{
		"configuration": "test",
	})
	require.NoError(t, err)

	agentAUpdates := &AgentUpdates{
		Configuration: configuration,
	}

	agentBUpdates := &AgentUpdates{
		Labels:        &labels,
		Configuration: configuration,
	}

	testProtocol.
		On("Connected", testAgentA.ID).Return(true).
		On("Connected", testAgentB.ID).Return(true).
		On("Connected", testAgentC.ID).Return(false).
		On("UpdateAgent", mock.Anything, testAgentA, agentAUpdates).Return(nil).
		On("UpdateAgent", mock.Anything, testAgentB2, agentBUpdates).Return(nil)

	testManager.handleUpdates(updates)

	testProtocol.AssertExpectations(t)
}

func TestManagerVerifySecretKey(t *testing.T) {
	tests := []struct {
		name             string
		managerSecretKey string
		agentSecretKey   string
		expect           bool
	}{
		{
			name:             "no manager key, no agent key",
			managerSecretKey: "",
			agentSecretKey:   "",
			expect:           true,
		},
		{
			name:             "no manager key, any agent key",
			managerSecretKey: "",
			agentSecretKey:   "any",
			expect:           true,
		},
		{
			name:             "manager key, no agent key",
			managerSecretKey: "test",
			agentSecretKey:   "",
			expect:           false,
		},
		{
			name:             "manager key matches agent key",
			managerSecretKey: "test",
			agentSecretKey:   "test",
			expect:           true,
		},
		{
			name:             "manager key doesn't match agent key",
			managerSecretKey: "test",
			agentSecretKey:   "something else",
			expect:           false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testManager := &manager{
				secretKey: test.managerSecretKey,
			}
			require.Equal(t, test.expect, testManager.VerifySecretKey(context.TODO(), test.agentSecretKey))
		})
	}
}

// -------------------------
// Protocol is an autogenerated mock type for the Protocol type
type mockProtocol struct {
	mock.Mock
}

// Connected provides a mock function with given fields: agentID
func (_m *mockProtocol) Disconnect(agentID string) bool {
	ret := _m.Called(agentID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(agentID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

func (_m *mockProtocol) Connected(agentID string) bool {
	ret := _m.Called(agentID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(agentID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ConnectedAgentIDs provides a mock function with given fields: _a0
func (_m *mockProtocol) ConnectedAgentIDs(_a0 context.Context) ([]string, error) {
	ret := _m.Called(_a0)

	var r0 []string
	if rf, ok := ret.Get(0).(func(context.Context) []string); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Name provides a mock function with given fields:
func (_m *mockProtocol) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// SendHeartbeat provides a mock function with given fields: agentID
func (_m *mockProtocol) SendHeartbeat(agentID string) error {
	ret := _m.Called(agentID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(agentID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateAgent provides a mock function with given fields: _a0, _a1, _a2
func (_m *mockProtocol) UpdateAgent(_a0 context.Context, _a1 *model.Agent, _a2 *AgentUpdates) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Agent, *AgentUpdates) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
