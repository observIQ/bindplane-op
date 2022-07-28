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

package store

// This file contains shared tests for mapstore and boltstore

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/observiq/bindplane-op/internal/eventbus"
	"github.com/observiq/bindplane-op/internal/store/search"
	"github.com/observiq/bindplane-op/model"
)

func addAgent(s Store, agent *model.Agent) error {
	_, err := s.UpsertAgent(context.TODO(), agent.ID, func(a *model.Agent) {
		*a = *agent
	})
	return err
}

func labels(m map[string]string) model.Labels {
	labels, _ := model.LabelsFromMap(m)
	return labels
}

var (
	cabinDestinationType = model.NewDestinationType("cabin", []model.ParameterDefinition{
		{
			Name: "s",
			Type: "string",
		},
	})

	cabinDestination1        = model.NewDestination("cabin-1", "cabin", []model.Parameter{})
	cabinDestination1Changed = model.NewDestination("cabin-1", "cabin", []model.Parameter{
		{
			Name:  "s",
			Value: "1",
		},
	})
	cabinDestination2 = model.NewDestination("cabin-2", "cabin", []model.Parameter{})

	macosSourceType = model.NewSourceType("macos", []model.ParameterDefinition{
		{
			Name: "s",
			Type: "string",
		},
	})
	macosSource        = model.NewSource("macos-1", "macos", []model.Parameter{})
	macosSourceChanged = model.NewSource("macos-1", "macos", []model.Parameter{
		{
			Name:  "s",
			Value: "1",
		},
	})

	nginxSourceType = model.NewSourceType("nginx", []model.ParameterDefinition{
		{
			Name: "s",
			Type: "string",
		},
	})
	nginxSource        = model.NewSource("nginx", "nginx", []model.Parameter{})
	nginxSourceChanged = model.NewSource("nginx", "nginx", []model.Parameter{
		{
			Name:  "s",
			Value: "1",
		},
	})

	invalidSource  = model.NewSource("_production-nginx-ingress_", "macos", []model.Parameter{})
	invalidSource2 = model.NewSource("foo/bar/baz", "macos", []model.Parameter{})

	unknownResource = model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			Kind: model.Kind("not-a-real-resource"),
			Metadata: model.Metadata{
				Name: "unknown",
			},
		},
	}

	testConfiguration = model.NewConfigurationWithSpec("configuration-1", model.ConfigurationSpec{
		Sources: []model.ResourceConfiguration{
			{
				Name: macosSource.Name(),
			},
		},
		Destinations: []model.ResourceConfiguration{
			{
				Name: cabinDestination1.Name(),
			},
		},
	})

	testConfigurationChanged = model.NewConfigurationWithSpec("configuration-1", model.ConfigurationSpec{
		Sources: []model.ResourceConfiguration{
			{
				Name: macosSource.Name(),
				Parameters: []model.Parameter{
					{
						Name:  "s",
						Value: "1",
					},
				},
			},
		},
		Destinations: []model.ResourceConfiguration{
			{
				Name: cabinDestination1.Name(),
			},
		},
	})

	testRawConfiguration1 = model.NewRawConfiguration("test-configuration-1", "raw:")
	testRawConfiguration2 = model.NewRawConfiguration("test-configuration-2", "raw:")
)

func applyTestTypes(t *testing.T, store Store) {
	statuses, err := store.ApplyResources([]model.Resource{
		cabinDestinationType,
		macosSourceType,
		nginxSourceType,
	})
	require.NoError(t, err)
	requireOkStatuses(t, statuses)
}

func applyTestConfiguration(t *testing.T, store Store) {
	statuses, err := store.ApplyResources([]model.Resource{
		cabinDestinationType,
		cabinDestination1,
		cabinDestination2,
		macosSourceType,
		macosSource,
		nginxSourceType,
		nginxSource,
		testConfiguration,
	})
	t.Logf("statuses %v\n", statuses)
	require.NoError(t, err)
	requireOkStatuses(t, statuses)
}

func applyAllTestResources(t *testing.T, store Store) {
	statuses, err := store.ApplyResources([]model.Resource{
		cabinDestinationType,
		cabinDestination1,
		cabinDestination2,
		macosSourceType,
		macosSource,
		nginxSourceType,
		nginxSource,
		testConfiguration,
		testRawConfiguration1,
		testRawConfiguration2,
	})
	require.NoError(t, err)
	requireOkStatuses(t, statuses)
}

type configurationChanges struct {
	configurationsUpdated []string
	configurationsRemoved []string
}

func expectedUpdates(configurations ...string) configurationChanges {
	return configurationChanges{
		configurationsUpdated: configurations,
	}
}

func expectedRemoves(configurations ...string) configurationChanges {
	return configurationChanges{
		configurationsRemoved: configurations,
	}
}

func configurationChangesFromUpdates(updates *Updates) configurationChanges {
	var updated []string
	var removed []string

	for _, event := range updates.Configurations {
		if event.Type == EventTypeRemove {
			removed = append(removed, event.Item.Name())
		} else {
			updated = append(updated, event.Item.Name())
		}
	}
	changes := configurationChanges{
		configurationsUpdated: updated,
		configurationsRemoved: removed,
	}
	return changes
}

func verifyUpdates(t *testing.T, done chan bool, Updates <-chan *Updates, expected []configurationChanges) {
	complete := func(success bool) {
		done <- success
	}
	i := 0
	for {
		select {
		case <-time.After(5 * time.Second):
			complete(false)
			t.Log("Timed out waiting for updates.")
			return
		case updates := <-Updates:
			if !assert.Less(t, i, len(expected), "more changes than expected") {
				complete(false)
				return
			}
			actual := configurationChangesFromUpdates(updates)

			t.Logf("actual %v\nexpected %v", actual, expected)

			if !assert.ElementsMatch(t, expected[i].configurationsRemoved, actual.configurationsRemoved, "configurationsRemoved should match") {
				complete(false)
				return
			}
			if !assert.ElementsMatch(t, expected[i].configurationsUpdated, actual.configurationsUpdated, "configurationsUpdated should match") {
				complete(false)
				return
			}
			i++
			if i == len(expected) {
				complete(true)
				return
			}
		}
	}
}

func runNotifyUpdatesTests(t *testing.T, store Store, done chan bool) {

	update := func(r model.Resource) {
		status, err := store.ApplyResources([]model.Resource{r})
		require.NoError(t, err)
		requireOkStatuses(t, status)
	}

	updates, _ := eventbus.Subscribe(store.Updates())
	applyAllTestResources(t, store)
	go verifyUpdates(t, done, updates, []configurationChanges{
		expectedUpdates(testConfiguration.Name(), testRawConfiguration1.Name(), testRawConfiguration2.Name()),
	})
	ok := <-done
	require.True(t, ok)

	// these tests are dependent on each other and are expected to run in order.

	t.Run("update nginx, expect no configuration changes", func(t *testing.T) {
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(),
		})
		update(nginxSourceChanged)
		ok := <-done
		require.True(t, ok)
	})

	t.Run("update configuration, expect configuration-1 change", func(t *testing.T) {
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
		})
		update(testConfigurationChanged)
		ok := <-done
		require.True(t, ok)
	})

	t.Run("update macos, expect configuration-1 change", func(t *testing.T) {
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
		})
		update(macosSourceChanged)
		ok := <-done
		require.True(t, ok)
	})

	t.Run("update cabin-1, expect configuration-1 change", func(t *testing.T) {
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
		})
		update(cabinDestination1Changed)
		ok := <-done
		require.True(t, ok)
	})

	t.Run("update everything, expect configuration-1 change", func(t *testing.T) {
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
		})
		store.ApplyResources([]model.Resource{
			macosSource,
			macosSourceType,
			nginxSource,
			nginxSourceType,
			cabinDestination1,
			cabinDestination2,
			testConfiguration,
		})
		ok := <-done
		require.True(t, ok)
	})

	t.Run("delete configuration, expect configuration-1 remove", func(t *testing.T) {
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedRemoves(testConfiguration.Name()),
		})

		// setup
		applyTestConfiguration(t, store)
		// Test batch delete here
		_, err := store.DeleteConfiguration(testConfiguration.Name())
		require.NoError(t, err)

		ok := <-done
		require.True(t, ok)
	})
}

func runDeleteChannelTests(t *testing.T, store Store, done chan bool) {
	t.Run("delete configuration, expect configuration-1 in deleteconfigurations channel", func(t *testing.T) {
		updates, unsubscribe := eventbus.Subscribe(store.Updates())
		defer unsubscribe()
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
			expectedRemoves(testConfiguration.Name()),
		})

		// seed
		store.Clear()
		applyTestConfiguration(t, store)
		// delete the configuration
		_, err := store.DeleteResources([]model.Resource{
			testConfiguration,
		})
		require.NoError(t, err)

		ok := <-done
		require.True(t, ok)
	})

	t.Run("batch delete a single configuration, expect configuration-1 in deleteconfigurations channel", func(t *testing.T) {
		updates, unsubscribe := eventbus.Subscribe(store.Updates())
		defer unsubscribe()
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
			expectedRemoves(testConfiguration.Name()),
		})

		// seed
		store.Clear()
		applyTestConfiguration(t, store)
		_, err := store.DeleteResources([]model.Resource{
			testConfiguration,
		})

		require.NoError(t, err)

		ok := <-done
		require.True(t, ok)
	})

	t.Run("batch delete a source attached to a configuration expect source in-use status", func(t *testing.T) {
		updates, unsubscribe := eventbus.Subscribe(store.Updates())
		defer unsubscribe()
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
		})

		// seed
		store.Clear()
		applyTestConfiguration(t, store)
		statuses, err := store.DeleteResources([]model.Resource{
			macosSourceChanged,
		})
		assert.NoError(t, err, "expect no error on valid delete")
		require.ElementsMatch(t, []model.ResourceStatus{
			{
				Resource: macosSourceChanged,
				Status:   model.StatusInUse,
				Reason:   "Dependent resources:\nConfiguration configuration-1\n",
			},
		}, statuses)

		ok := <-done
		require.True(t, ok)
	})

	t.Run("batch delete source and its configuration, expect configuration-1 in channel", func(t *testing.T) {
		updates, unsubscribe := eventbus.Subscribe(store.Updates())
		defer unsubscribe()
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
			expectedRemoves(testConfiguration.Name()),
		})

		// seed
		store.Clear()
		applyTestConfiguration(t, store)
		_, err := store.DeleteResources([]model.Resource{
			testConfiguration,
			macosSource,
		})
		require.NoError(t, err)

		ok := <-done
		require.True(t, ok)
	})
}

func runAgentSubscriptionsTest(t *testing.T, store Store) {
	agent := &model.Agent{
		ID:   "1",
		Name: "agent-1",
	}
	afterStatus := &model.Agent{
		ID:     "1",
		Name:   "agent-1",
		Status: 1,
	}

	tests := []struct {
		description     string
		updaterFunction AgentUpdater
		expect          []*model.Agent
	}{
		{
			description: "agent in channel after creation",
			updaterFunction: func(current *model.Agent) {
				*current = *agent
			},
			expect: []*model.Agent{agent},
		},
		{
			description: "agent in channel after changing status",
			updaterFunction: func(current *model.Agent) {
				*current = *afterStatus
			},
			expect: []*model.Agent{afterStatus},
		},
	}

	channel, _ := eventbus.Subscribe(store.Updates())

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			done := make(chan bool)
			go verifyAgentChanges(t, done, channel, test.expect)

			_, err := store.UpsertAgent(context.TODO(), agent.ID, test.updaterFunction)
			require.NoError(t, err)

			ok := <-done
			assert.True(t, ok)
		})
	}
}

func verifyAgentChanges(t *testing.T, done chan bool, agentChanges <-chan *Updates, expectedUpdates []*model.Agent) {
	for {
		select {
		case <-time.After(5 * time.Second):
			done <- false
			return
		case changes := <-agentChanges:
			agents := []*model.Agent{}
			for _, change := range changes.Agents {
				agents = append(agents, change.Item)
			}
			if !assert.ElementsMatch(t, expectedUpdates, agents) {
				done <- false
				return
			}

			done <- true
			return
		}
	}
}

func verifyAgentUpdates(t *testing.T, done chan bool, agentChanges <-chan *Updates, expectedUpdates []string) {
	for {
		select {
		case <-time.After(5 * time.Second):
			done <- false
			return
		case changes := <-agentChanges:
			ids := []string{}
			for _, change := range changes.Agents {
				ids = append(ids, change.Item.ID)
			}

			if !assert.ElementsMatch(t, ids, expectedUpdates) {
				done <- false
				return
			}

			done <- true
			return
		}
	}
}

func runUpdateAgentsTests(t *testing.T, store Store) {
	// Tests for UpsertAgent
	upsertAgentTests := []struct {
		description   string
		agent         *model.Agent
		updater       AgentUpdater
		expectUpdates []string
	}{
		{
			description:   "upsertAgent passes along updates",
			agent:         &model.Agent{ID: "1", Status: 0},
			updater:       func(current *model.Agent) { current.Status = 1 },
			expectUpdates: []string{"1"},
		},
	}

	done := make(chan bool)
	channel, _ := eventbus.Subscribe(store.Updates())

	for _, test := range upsertAgentTests {
		t.Run(test.description, func(t *testing.T) {
			go verifyAgentUpdates(t, done, channel, test.expectUpdates)

			_, err := store.UpsertAgent(context.TODO(), test.agent.ID, test.updater)
			require.NoError(t, err)

			ok := <-done
			require.True(t, ok)
		})
	}

	// Tests for UpsertAgents (bulk)
	upsertAgentsTests := []struct {
		description   string
		agents        []*model.Agent
		updater       AgentUpdater
		expectUpdates []string
	}{
		{
			description: "upsertAgents passes along a single update",
			agents: []*model.Agent{
				{ID: "1"},
			},
			updater:       func(current *model.Agent) { current.Status = 1 },
			expectUpdates: []string{"1"},
		},
		{
			description: "upsertAgents passes along multiple updates in single message",
			agents: []*model.Agent{
				{ID: "1"},
				{ID: "2"},
				{ID: "3"},
				{ID: "4"},
				{ID: "5"},
				{ID: "6"},
				{ID: "7"},
				{ID: "8"},
				{ID: "9"},
				{ID: "10"},
			},
			updater:       func(current *model.Agent) { current.Status = 1 },
			expectUpdates: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
		},
	}

	for _, test := range upsertAgentsTests {
		t.Run(test.description, func(t *testing.T) {
			go verifyAgentUpdates(t, done, channel, test.expectUpdates)

			ids := make([]string, len(test.agents))
			for ix, a := range test.agents {
				ids[ix] = a.ID
			}

			_, err := store.UpsertAgents(context.TODO(), ids, test.updater)
			require.NoError(t, err)

			ok := <-done
			require.True(t, ok)
		})
	}
}

// These tests that the ApplyResources methods return the expected resources with statuses
func runApplyResourceReturnTests(t *testing.T, store Store) {
	tests := []struct {
		description string
		// initial resources to seed
		initialResources []model.Resource
		// resources to apply in the test call
		applyResources []model.Resource
		expect         []model.ResourceStatus
	}{
		{
			description:      "applies a single resource, returns created status",
			initialResources: []model.Resource{},
			applyResources:   []model.Resource{macosSource},
			expect:           []model.ResourceStatus{*model.NewResourceStatus(macosSource, model.StatusCreated)},
		},
		{
			description:      "applies a multiple new resources, returns all created statuses",
			initialResources: []model.Resource{},
			applyResources:   []model.Resource{macosSource, nginxSource, cabinDestination1},
			expect: []model.ResourceStatus{
				*model.NewResourceStatus(macosSource, model.StatusCreated),
				*model.NewResourceStatus(nginxSource, model.StatusCreated),
				*model.NewResourceStatus(cabinDestination1, model.StatusCreated),
			},
		},
		{
			description:      "applies resource to existing resource, returns status unchanged",
			initialResources: []model.Resource{macosSource},
			applyResources:   []model.Resource{macosSource},
			expect:           []model.ResourceStatus{*model.NewResourceStatus(macosSource, model.StatusUnchanged)},
		},
		{
			description:      "applies a changed resource to an existsting, returns status configured",
			initialResources: []model.Resource{macosSource},
			applyResources:   []model.Resource{macosSourceChanged},
			expect:           []model.ResourceStatus{*model.NewResourceStatus(macosSourceChanged, model.StatusConfigured)},
		},
		{
			description:      "applies mixed resource updates, returns correct statuses",
			initialResources: []model.Resource{macosSource, nginxSource},
			applyResources:   []model.Resource{macosSourceChanged, nginxSource},
			expect: []model.ResourceStatus{
				*model.NewResourceStatus(macosSourceChanged, model.StatusConfigured),
				*model.NewResourceStatus(nginxSource, model.StatusUnchanged),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// Setup
			store.Clear()
			applyTestTypes(t, store)
			_, err := store.ApplyResources(test.initialResources)
			require.NoError(t, err, "expect no error in setup apply call")

			statuses, err := store.ApplyResources(test.applyResources)
			require.NoError(t, err, "expect no error in valid apply call")

			assert.ElementsMatch(t, test.expect, statuses)
		})
	}
}

func runValidateApplyResourcesTests(t *testing.T, store Store) {
	tests := []struct {
		name      string
		resources []model.Resource
		reasons   []string
		statuses  []model.UpdateStatus
	}{
		{
			name:      "none",
			resources: []model.Resource{},
		},
		{
			name:      "all valid",
			resources: []model.Resource{macosSource, nginxSource},
			statuses:  []model.UpdateStatus{model.StatusCreated, model.StatusCreated},
			reasons:   []string{"", ""},
		},
		{
			name:      "one invalid",
			resources: []model.Resource{invalidSource},
			reasons:   []string{"_production-nginx-ingress_ is not a valid resource name"},
			statuses:  []model.UpdateStatus{model.StatusInvalid},
		},
		{
			name:      "two invalid of four",
			resources: []model.Resource{macosSource, invalidSource, invalidSource2, nginxSource},
			reasons:   []string{"", "_production-nginx-ingress_ is not a valid resource name", "foo/bar/baz is not a valid resource name", ""},
			statuses:  []model.UpdateStatus{model.StatusCreated, model.StatusInvalid, model.StatusInvalid, model.StatusCreated},
		},
		{
			name:      "invalid and unknown",
			resources: []model.Resource{invalidSource, &unknownResource},
			reasons:   []string{"_production-nginx-ingress_ is not a valid resource name", "not-a-real-resource is not a valid resource kind"},
			statuses:  []model.UpdateStatus{model.StatusInvalid, model.StatusInvalid},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store.Clear()
			_, err := store.ApplyResources([]model.Resource{
				macosSourceType,
				nginxSourceType,
				cabinDestinationType,
			})
			require.NoError(t, err)
			result, err := store.ApplyResources(test.resources)
			require.NoError(t, err)
			for i, status := range test.statuses {
				require.Equal(t, status, result[i].Status, result[i].Reason)
				require.Contains(t, result[i].Reason, test.reasons[i])
			}
		})
	}
}

func runDeleteResourcesReturnTests(t *testing.T, store Store) {
	tests := []struct {
		description      string
		initialResources []model.Resource
		deleteResources  []model.Resource
		expect           []model.ResourceStatus
	}{
		{
			description:      "calling delete on a non existent resource returns no resource status",
			initialResources: make([]model.Resource, 0),
			deleteResources:  []model.Resource{nginxSource},
			expect:           make([]model.ResourceStatus, 0),
		},
		{
			description:      "calling delete on an existing resource returns a single resource status",
			initialResources: []model.Resource{macosSource},
			deleteResources:  []model.Resource{macosSource},
			expect: []model.ResourceStatus{
				*model.NewResourceStatus(macosSource, model.StatusDeleted),
			},
		},
		{
			description:      "calling delete on one existing and one non existent resource returns single resource status",
			initialResources: []model.Resource{macosSource},
			deleteResources:  []model.Resource{macosSource, nginxSource},
			expect: []model.ResourceStatus{
				*model.NewResourceStatus(macosSource, model.StatusDeleted),
			},
		},
		{
			description:      "calling delete on multiple resources returns all resources deleted",
			initialResources: []model.Resource{macosSource, nginxSource, cabinDestination1},
			deleteResources:  []model.Resource{macosSource, nginxSource, cabinDestination1},
			expect: []model.ResourceStatus{
				*model.NewResourceStatus(macosSource, model.StatusDeleted),
				*model.NewResourceStatus(nginxSource, model.StatusDeleted),
				*model.NewResourceStatus(cabinDestination1, model.StatusDeleted),
			},
		},
		{
			description:      "calling delete on an in use resources returns update with status In Use",
			initialResources: []model.Resource{macosSource, nginxSource, cabinDestination1, testConfiguration},
			deleteResources:  []model.Resource{macosSource},
			expect: []model.ResourceStatus{
				*model.NewResourceStatusWithReason(macosSource, model.StatusInUse, "Dependent resources:\nConfiguration configuration-1\n"),
			},
		},
		{
			description:      "calling delete on an in use resources and its dependency returns all deleted",
			initialResources: []model.Resource{macosSource, nginxSource, cabinDestination1, testConfiguration},
			deleteResources:  []model.Resource{testConfiguration, cabinDestination1},
			expect: []model.ResourceStatus{
				*model.NewResourceStatus(cabinDestination1, model.StatusDeleted),
				*model.NewResourceStatus(testConfiguration, model.StatusDeleted),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// setup
			store.Clear()
			applyTestTypes(t, store)
			_, err := store.ApplyResources(test.initialResources)
			require.NoError(t, err, "expect no error in seed apply")

			statuses, err := store.DeleteResources(test.deleteResources)
			require.NoError(t, err, "expect no error on valid delete call")

			assert.ElementsMatch(t, test.expect, statuses)
		})
	}
}

func runDependentResourcesTests(t *testing.T, s Store) {
	tests := []struct {
		description      string
		initialResources []model.Resource
		testResource     model.Resource
		expect           DependentResources
	}{
		{
			description: "macos source has configuration dependency",
			initialResources: []model.Resource{
				macosSourceType,
				macosSource,
				cabinDestinationType,
				cabinDestination1,
				testConfiguration,
			},
			testResource: macosSource,
			expect: DependentResources{
				{
					name: testConfiguration.Name(),
					kind: model.KindConfiguration,
				},
			},
		},
		{
			description: "cabin destination has configuration dependency",
			initialResources: []model.Resource{
				macosSourceType,
				macosSource,
				cabinDestinationType,
				cabinDestination1,
				testConfiguration,
			},
			testResource: cabinDestination1,
			expect: DependentResources{
				{
					name: testConfiguration.Name(),
					kind: model.KindConfiguration,
				},
			},
		},
	}

	for _, test := range tests {
		updates, err := s.ApplyResources(test.initialResources)
		fmt.Println("UPDATES: ", updates)

		dependencies, err := FindDependentResources(context.TODO(), s, test.testResource)
		require.NoError(t, err)
		assert.Equal(t, test.expect, dependencies)
	}
}

func runIndividualDeleteTests(t *testing.T, store Store) {
	setup := func() {
		store.Clear()
		_, err := store.ApplyResources([]model.Resource{
			macosSourceType,
			macosSource,
			nginxSourceType,
			nginxSource,
			cabinDestinationType,
			cabinDestination1,
			cabinDestination2,
			testConfiguration})
		require.NoError(t, err)
	}

	t.Run("DeleteSource", func(t *testing.T) {
		tests := []struct {
			description  string
			source       string
			expectError  error
			expectSource *model.Source
		}{
			{
				description:  "delete nginx",
				source:       nginxSource.Name(),
				expectError:  nil,
				expectSource: nginxSource,
			},
			{
				description: "delete macos, get dependency error",
				source:      macosSource.Name(),
				expectError: newDependencyError(DependentResources{
					dependency{name: testConfiguration.Name(),
						kind: model.KindConfiguration},
				}),
				expectSource: macosSource,
			},
			{
				description:  "delete non existent, no resource, no error",
				source:       "foo",
				expectError:  nil,
				expectSource: nil,
			},
		}

		for _, test := range tests {
			setup()

			src, err := store.DeleteSource(test.source)
			assert.Equal(t, test.expectSource, src)
			assert.Equal(t, test.expectError, err)
		}
	})

	t.Run("DeleteDestination", func(t *testing.T) {
		tests := []struct {
			description  string
			destination  string
			expectError  error
			expectSource *model.Destination
		}{
			{
				description:  "delete cabinDestination2",
				destination:  cabinDestination2.Name(),
				expectError:  nil,
				expectSource: cabinDestination2,
			},
			{
				description: "delete cabinDestination1, expect error",
				destination: cabinDestination1.Name(),
				expectError: newDependencyError(DependentResources{
					dependency{
						name: testConfiguration.Name(),
						kind: model.KindConfiguration,
					},
				}),
				expectSource: cabinDestination1,
			},
			{
				description:  "delete non existent, expect nil error and  destination",
				destination:  "foo",
				expectError:  nil,
				expectSource: nil,
			},
		}

		for _, test := range tests {
			setup()

			dest, err := store.DeleteDestination(test.destination)
			assert.Equal(t, test.expectSource, dest)
			assert.Equal(t, test.expectError, err)
		}
	})
}

func verifyAgentsRemove(t *testing.T, done chan bool, Updates <-chan *Updates, expectRemoves []string) {
	var val struct{}
	removesRemaining := map[string]struct{}{}
	for _, r := range expectRemoves {
		removesRemaining[r] = val
	}
	for {
		select {
		case <-time.After(5 * time.Second):
			done <- false
			t.Log("Timed out waiting for updates.")
			return
		case updates, ok := <-Updates:
			if !ok {
				done <- false
				return
			}
			agentUpdates := updates.Agents

			// skip when we're seeding
			var skip = false
			for _, update := range agentUpdates {
				if update.Type != EventTypeRemove {
					skip = true
				}
			}
			if skip {
				continue
			}

			for _, update := range updates.Agents {
				if update.Type == EventTypeRemove {
					delete(removesRemaining, update.Item.ID)
				}
			}

			if len(removesRemaining) == 0 {
				done <- true
				return
			}
		}
	}
}

// runDeleteAgentsTests tests store.DeleteAgents
func runDeleteAgentsTests(t *testing.T, store Store) {
	deleteTests := []struct {
		description    string
		seedAgentsIDs  []string
		deleteAgentIDs []string
		// The agents returned by the delete method
		expectDeleted []*model.Agent
		// The agents returned by the store
		expectAgents []*model.Agent
	}{
		{
			description:    "delete 1 agent",
			seedAgentsIDs:  []string{"1"},
			deleteAgentIDs: []string{"1"},
			expectDeleted: []*model.Agent{
				{ID: "1", Status: 5, Labels: model.MakeLabels()},
			},
			// The agents left in the store after delete
			expectAgents: make([]*model.Agent, 0),
		},
		{
			description:    "delete multiple agents",
			seedAgentsIDs:  []string{"1", "2", "3", "4", "5"},
			deleteAgentIDs: []string{"1", "2", "3"},
			expectDeleted: []*model.Agent{
				{ID: "1", Status: 5, Labels: model.MakeLabels()},
				{ID: "2", Status: 5, Labels: model.MakeLabels()},
				{ID: "3", Status: 5, Labels: model.MakeLabels()},
			},
			expectAgents: []*model.Agent{
				{ID: "4", Labels: model.MakeLabels()},
				{ID: "5", Labels: model.MakeLabels()},
			},
		},
		{
			description:    "delete non existing agent, no error, no delete",
			seedAgentsIDs:  []string{"1"},
			deleteAgentIDs: []string{"42"},
			expectDeleted:  make([]*model.Agent, 0),
			expectAgents:   []*model.Agent{{ID: "1", Labels: model.MakeLabels()}},
		},
	}

	// Test the delete operation
	for _, test := range deleteTests {
		// setup
		ctx := context.Background()
		store.Clear()

		// seed agents
		for _, id := range test.seedAgentsIDs {
			addAgent(store, &model.Agent{ID: id, Labels: model.MakeLabels()})
		}

		t.Run(test.description, func(t *testing.T) {
			deleted, err := store.DeleteAgents(ctx, test.deleteAgentIDs)
			require.NoError(t, err)
			assert.ElementsMatch(t, test.expectDeleted, deleted, "deleted agents do not match")

			rest, err := store.Agents(ctx)
			require.NoError(t, err)
			assert.ElementsMatch(t, test.expectAgents, rest, "remaining agents do not match")
		})
	}

	t.Run("deleting an agent removes it from the index", func(t *testing.T) {
		// setup
		store.Clear()
		ctx := context.Background()

		// seed agent
		addAgent(store, &model.Agent{ID: "1"})

		// verify its in the index
		results, err := search.Field(ctx, store.AgentIndex(), "id", "1")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"1"}, results)

		// delete it
		_, err = store.DeleteAgents(ctx, []string{"1"})
		require.NoError(t, err)

		results, err = search.Field(ctx, store.AgentIndex(), "id", "1")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{}, results)
	})

	deleteUpdatesTests := []struct {
		description    string
		seedAgentIDs   []string
		deleteAgentIDs []string
	}{
		{
			description:    "delete an agent, expect a remove update in Updates.Agents",
			seedAgentIDs:   []string{"1"},
			deleteAgentIDs: []string{"1"},
		},
	}

	for _, test := range deleteUpdatesTests {
		t.Run(test.description, func(t *testing.T) {
			// setup
			store.Clear()
			for _, id := range test.seedAgentIDs {
				addAgent(store, &model.Agent{ID: id})
			}

			channel, unsubscribe := eventbus.Subscribe(store.Updates())
			defer unsubscribe()

			done := make(chan bool, 0)
			go verifyAgentsRemove(t, done, channel, test.deleteAgentIDs)

			ctx := context.Background()
			_, err := store.DeleteAgents(ctx, test.deleteAgentIDs)
			require.NoError(t, err)

			ok := <-done
			assert.True(t, ok)
		})
	}
}

// runConfigurationsTests runs tests on Store.Configuration and Store.Configurations
func runConfigurationsTests(t *testing.T, store Store) {
	t.Run("lists all configurations", func(t *testing.T) {
		// Setup
		status, err := store.ApplyResources([]model.Resource{testRawConfiguration1, testRawConfiguration2})
		require.NoError(t, err)
		requireOkStatuses(t, status)

		configs, err := store.Configurations()
		assert.NoError(t, err)
		assert.ElementsMatch(t, []*model.Configuration{testRawConfiguration1, testRawConfiguration2}, configs)
	})
}

func runConfigurationTests(t *testing.T, store Store) {
	t.Run("gets configuration by name", func(t *testing.T) {
		// Setup
		status, err := store.ApplyResources([]model.Resource{testRawConfiguration1, testRawConfiguration2})
		require.NoError(t, err)
		requireOkStatuses(t, status)

		config, err := store.Configuration(testRawConfiguration1.Name())
		assert.NoError(t, err)
		assert.Equal(t, testRawConfiguration1, config)
	})
}

func runPagingTests(t *testing.T, store Store) {
	for i := 0; i < 100; i++ {
		store.UpsertAgent(context.TODO(), fmt.Sprintf("%03d", i), func(current *model.Agent) {
			current.Name = "agent-" + current.ID
		})
	}
	tests := []struct {
		name      string
		offset    int
		limit     int
		expectIDs []string
	}{
		{
			name:   "first page",
			offset: 0,
			limit:  10,
			expectIDs: []string{
				"agent-000",
				"agent-001",
				"agent-002",
				"agent-003",
				"agent-004",
				"agent-005",
				"agent-006",
				"agent-007",
				"agent-008",
				"agent-009",
			},
		},
		{
			name:   "second page",
			offset: 10,
			limit:  10,
			expectIDs: []string{
				"agent-010",
				"agent-011",
				"agent-012",
				"agent-013",
				"agent-014",
				"agent-015",
				"agent-016",
				"agent-017",
				"agent-018",
				"agent-019",
			},
		},
		{
			name:   "last few",
			offset: 95,
			limit:  10,
			expectIDs: []string{
				"agent-095",
				"agent-096",
				"agent-097",
				"agent-098",
				"agent-099",
			},
		},
		{
			name:      "page too large",
			offset:    200,
			limit:     10,
			expectIDs: []string{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agents, err := store.Agents(context.TODO(), WithOffset(test.offset), WithLimit(test.limit))
			require.NoError(t, err)
			ids := []string{}
			for _, agent := range agents {
				ids = append(ids, agent.Name)
			}
			require.ElementsMatch(t, test.expectIDs, ids)
		})
	}
	t.Run("agents count", func(t *testing.T) {
		count, err := store.AgentsCount(context.TODO())
		require.NoError(t, err)
		require.Equal(t, 100, count)
	})
}

func runTestUpsertAgents(t *testing.T, store Store) {
	t.Run("can insert new agents", func(t *testing.T) {
		store.Clear()
		count, err := store.AgentsCount(context.TODO())
		require.NoError(t, err)
		require.Zero(t, count)

		returnedAgents, err := store.UpsertAgents(
			context.TODO(),
			[]string{"1", "2", "3"},
			func(current *model.Agent) {
				current.Labels = model.MakeLabels()
			},
		)
		require.NoError(t, err)

		expectAgents := []*model.Agent{
			{ID: "1", Labels: model.MakeLabels()},
			{ID: "2", Labels: model.MakeLabels()},
			{ID: "3", Labels: model.MakeLabels()},
		}

		require.ElementsMatch(t, expectAgents, returnedAgents)

		gotAgents, err := store.Agents(context.TODO())
		require.NoError(t, err)
		require.ElementsMatch(t, expectAgents, gotAgents)
	})

	t.Run("upserts and updates agents correctly", func(t *testing.T) {
		tests := []struct {
			description    string
			initAgentsIDs  []string
			upsertAgentIDs []string
			updater        AgentUpdater
			expectAgents   []*model.Agent
		}{
			{
				description:    "updates existing agents and inserts",
				initAgentsIDs:  []string{"1"},
				upsertAgentIDs: []string{"1", "2"},
				updater:        func(current *model.Agent) { current.Status = 1; current.Labels = model.MakeLabels() },
				expectAgents: []*model.Agent{
					{ID: "1", Status: 1, Labels: model.MakeLabels()},
					{ID: "2", Status: 1, Labels: model.MakeLabels()},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.description, func(t *testing.T) {
				// setup
				store.Clear()

				// seed agents
				for _, id := range test.initAgentsIDs {
					addAgent(store, &model.Agent{ID: id, Labels: model.MakeLabels()})
				}

				// upsert
				returnedAgents, err := store.UpsertAgents(context.TODO(), test.upsertAgentIDs, test.updater)
				require.NoError(t, err)
				require.ElementsMatch(t, test.expectAgents, returnedAgents)

				// verify
				gotAgents, err := store.Agents(context.TODO())
				require.NoError(t, err)
				require.ElementsMatch(t, test.expectAgents, gotAgents)

			})
		}
	})

}

// ----------------------------------------------------------------------

func requireOkStatuses(t *testing.T, statuses []model.ResourceStatus) {
	for _, status := range statuses {
		require.Contains(t, []model.UpdateStatus{
			model.StatusUnchanged,
			model.StatusConfigured,
			model.StatusCreated,
			model.StatusDeleted,
		}, status.Status)
	}
}
