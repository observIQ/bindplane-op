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

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/observiq/bindplane/model"
)

func TestMapstoreNotifyUpdates(t *testing.T) {
	store := NewMapStore(zap.NewNop(), "super-secret-key")
	done := make(chan bool, 1)

	runNotifyUpdatesTests(t, store, done)
}

func TestMapstoreDeleteChannel(t *testing.T) {
	store := NewMapStore(zap.NewNop(), "super-secret-key")
	done := make(chan bool, 1)
	runDeleteChannelTests(t, store, done)
}

func TestMapstoreUpdateAgentsChannel(t *testing.T) {
	store := NewMapStore(zap.NewNop(), "super-secret-key")
	runUpdateAgentsTests(t, store)
}

func TestMapstoreApplyResourcesReturn(t *testing.T) {
	store := NewMapStore(zap.NewNop(), "super-secret-key")

	runApplyResourceReturnTests(t, store)
}

func TestMapstoreDeleteResourcesReturn(t *testing.T) {
	store := NewMapStore(zap.NewNop(), "super-secret-key")

	runDeleteResourcesReturnTests(t, store)
}

func TestMapstoreAgentSubscriptionsChannel(t *testing.T) {
	store := NewMapStore(zap.NewNop(), "super-secret-key")

	runAgentSubscriptionsTest(t, store)
}

func TestMapstoreResourcesEqual(t *testing.T) {
	resource1 := model.NewSourceType("resource1", []model.ParameterDefinition{})
	resource2 := model.NewSourceType("resource2", []model.ParameterDefinition{})
	resource3 := model.NewSourceType("resource1", []model.ParameterDefinition{})

	resource1.SetID("1")
	resource2.SetID("2")
	resource3.SetID("3")

	tests := []struct {
		description string
		r1          model.Resource
		r2          model.Resource
		expect      bool
	}{
		{
			description: "resources with different names and specs returns false",
			r1:          resource1,
			r2:          resource2,
			expect:      false,
		},
		{
			description: "resources with same name and spec but different ID returns true",
			r1:          resource1,
			r2:          resource3,
			expect:      true,
		},
		{
			description: "resource with same name spec and ID returns true",
			r1:          resource1,
			r2:          resource1,
			expect:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got := resourcesEqual(test.r1, test.r2)
			assert.Equal(t, test.expect, got)
		})
	}
}

func TestMapstoreConfigurations(t *testing.T) {
	store := NewMapStore(zap.NewNop(), "super-secret-key")

	runConfigurationsTests(t, store)
}

func TestMapstoreConfiguration(t *testing.T) {
	store := NewMapStore(zap.NewNop(), "super-secret-key")

	runConfigurationTests(t, store)
}

func TestMapstoreValidateApplyResourcesTests(t *testing.T) {
	store := NewMapStore(zap.NewNop(), "super-secret-key")
	runValidateApplyResourcesTests(t, store)
}

func TestMapstoreDependentResources(t *testing.T) {
	store := NewMapStore(zap.NewNop(), "super-secret-key")
	runDependentResourcesTests(t, store)
}

func TestMapstoreIndividualDelete(t *testing.T) {
	store := NewMapStore(zap.NewNop(), "super-secret-key")
	runIndividualDeleteTests(t, store)
}

func TestMapstorePaging(t *testing.T) {
	store := NewMapStore(zap.NewNop(), "super-secret-key")
	runPagingTests(t, store)
}

func TestMapstoreDeleteAgents(t *testing.T) {
	store := NewMapStore(zap.NewNop(), "super-secret-key")
	runDeleteAgentsTests(t, store)
}

func TestMapstoreUpsertAgents(t *testing.T) {
	store := NewMapStore(zap.NewNop(), "super-secret-key")
	runTestUpsertAgents(t, store)
}
