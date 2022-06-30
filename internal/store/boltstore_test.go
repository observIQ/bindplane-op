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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"github.com/observiq/bindplane/internal/store/search"
	"github.com/observiq/bindplane/model"
)

func TestClear(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err, "error while initializing test database", err)
	defer cleanupTestDB(t)

	s := NewBoltStore(db, "super-secret-key", zap.NewNop())
	s.ApplyResources([]model.Resource{
		macosSourceType,
		macosSource,
		cabinDestinationType,
		cabinDestination1,
		testRawConfiguration1,
	})

	s.Clear()

	sources, err := s.Sources()
	require.NoError(t, err)
	destinations, err := s.Destinations()
	require.NoError(t, err)
	configurations, err := s.Configurations()
	require.NoError(t, err)

	assert.Empty(t, sources)
	assert.Empty(t, destinations)
	assert.Empty(t, configurations)
}

func TestAddAgent(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err, "error while initializing test database", err)
	defer cleanupTestDB(t)

	s := NewBoltStore(db, "super-secret-key", zap.NewNop())
	a1 := &model.Agent{ID: "1", Name: "Fake Agent 1", Labels: model.Labels{Set: model.MakeLabels().Set}}
	a2 := &model.Agent{ID: "2", Name: "Fake Agent 2", Labels: model.Labels{Set: model.MakeLabels().Set}}

	err = addAgent(s, a1)
	require.NoError(t, err)
	err = addAgent(s, a2)
	require.NoError(t, err)

	var agents []*model.Agent

	db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("Agents")).Cursor()

		prefix := []byte("Agent")
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			agent := &model.Agent{}
			json.Unmarshal(v, agent)
			agents = append(agents, agent)
		}
		return nil
	})

	assert.Len(t, agents, 2)
	assert.ElementsMatch(t, agents, []interface{}{a1, a2})
}

type mockUnknownResource struct{}

func (x mockUnknownResource) ID() string                                  { return "" }
func (x mockUnknownResource) SetID(string)                                {}
func (x mockUnknownResource) EnsureID()                                   {}
func (x mockUnknownResource) GetKind() model.Kind                         { return model.KindUnknown }
func (x mockUnknownResource) Name() string                                { return "" }
func (x mockUnknownResource) Description() string                         { return "" }
func (x mockUnknownResource) Validate() error                             { return nil }
func (x mockUnknownResource) ValidateWithStore(model.ResourceStore) error { return nil }
func (x mockUnknownResource) GetLabels() model.Labels                     { return model.MakeLabels() }
func (x mockUnknownResource) UniqueKey() string                           { return x.ID() }

func (x mockUnknownResource) IndexID() string                  { return "" }
func (x mockUnknownResource) IndexFields(index search.Indexer) {}
func (x mockUnknownResource) IndexLabels(index search.Indexer) {}

var _ model.Resource = (*mockUnknownResource)(nil)

func TestKeyFromResource(t *testing.T) {
	cases := []struct {
		name     string
		resource model.Resource
		expect   string
	}{
		{
			"source",
			model.NewSourceType("test", []model.ParameterDefinition{}),
			"SourceType|test",
		},
		{
			"destination",
			model.NewDestinationType("test", []model.ParameterDefinition{}),
			"DestinationType|test",
		},
		{
			"nil",
			nil,
			"",
		},
		{
			"unknown",
			mockUnknownResource{},
			"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := string(keyFromResource(tc.resource))
			require.Equal(t, tc.expect, output)
		})
	}
}

func TestBoltStoreConfigurations(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err, "error while initializing test database", err)
	defer cleanupTestDB(t)
	s := NewBoltStore(db, "super-secret-key", zap.NewNop())

	runConfigurationsTests(t, s)
}

func TestBoltstoreConfiguration(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err, "error while initializing test database", err)
	defer cleanupTestDB(t)
	s := NewBoltStore(db, "super-secret-key", zap.NewNop())

	runConfigurationTests(t, s)
}
func TestAgents(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err, "error while initializing test database", err)
	defer cleanupTestDB(t)

	s := NewBoltStore(db, "super-secret-key", zap.NewNop())
	a1 := &model.Agent{ID: "1", Name: "Fake Agent 1", Labels: model.Labels{Set: model.MakeLabels().Set}}
	a2 := &model.Agent{ID: "2", Name: "Fake Agent 2", Labels: model.Labels{Set: model.MakeLabels().Set}}

	addAgent(s, a1)
	addAgent(s, a2)

	agents, err := s.Agents(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, agents, 2)
	assert.ElementsMatch(t, agents, []interface{}{a1, a2})
}

func TestAgent(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err, "error while initializing test database", err)
	defer cleanupTestDB(t)

	s := NewBoltStore(db, "super-secret-key", zap.NewNop())
	a1 := &model.Agent{ID: "1", Name: "Fake Agent 1", Labels: model.Labels{Set: model.MakeLabels().Set}}
	a2 := &model.Agent{ID: "2", Name: "Fake Agent 2", Labels: model.Labels{Set: model.MakeLabels().Set}}

	addAgent(s, a1)
	addAgent(s, a2)

	agent, err := s.Agent(a1.ID)
	assert.NoError(t, err)
	assert.Equal(t, a1, agent)
}

var updaterCalled bool

func testUpdater(agent *model.Agent) {
	updaterCalled = true
	agent.Name = "updated"
}

func TestUpsertAgent(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err, "error while initializing test database", err)
	defer cleanupTestDB(t)

	// Seed with one agent
	s := NewBoltStore(db, "super-secret-key", zap.NewNop())
	a1 := &model.Agent{ID: "1", Name: "Fake Agent 1", Labels: model.Labels{Set: model.MakeLabels().Set}}
	addAgent(s, a1)

	t.Run("creates a new agent if not found", func(t *testing.T) {
		newAgentID := "3"
		s.UpsertAgent(context.TODO(), newAgentID, testUpdater)

		got, err := s.Agent(newAgentID)
		require.NoError(t, err)

		assert.NotNil(t, got)
		assert.Equal(t, got.ID, newAgentID)
	})
	t.Run("calls updater and updates an agent if exists", func(t *testing.T) {
		updaterCalled = false
		s.UpsertAgent(context.TODO(), a1.ID, testUpdater)

		assert.True(t, updaterCalled)

		got, err := s.Agent(a1.ID)
		require.NoError(t, err)

		assert.Equal(t, got.Name, "updated")
	})
}

func TestBoltStoreNotifyUpdates(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err)
	defer cleanupTestDB(t)

	store := NewBoltStore(db, "super-secret-key", zap.NewNop())
	done := make(chan bool, 1)

	runNotifyUpdatesTests(t, store, done)
}

func TestBoltStoreDeleteChannel(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err)
	defer cleanupTestDB(t)

	store := NewBoltStore(db, "super-secret-key", zap.NewNop())
	done := make(chan bool, 1)

	runDeleteChannelTests(t, store, done)
}

func TestBoltStoreAgentSubscriptionChannel(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err)
	defer cleanupTestDB(t)

	store := NewBoltStore(db, "super-secret-key", zap.NewNop())
	runAgentSubscriptionsTest(t, store)
}

func TestBoltStoreAgentUpdatesChannel(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err)
	defer cleanupTestDB(t)
	store := NewBoltStore(db, "super-secret-key", zap.NewNop())

	runUpdateAgentsTests(t, store)
}

func TestBoltstoreApplyResourceReturn(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err)
	defer cleanupTestDB(t)

	store := NewBoltStore(db, "super-secret-key", zap.NewNop())
	runApplyResourceReturnTests(t, store)
}

func TestBoltstoreDeleteResourcesReturn(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err)
	defer cleanupTestDB(t)

	store := NewBoltStore(db, "super-secret-key", zap.NewNop())
	runDeleteResourcesReturnTests(t, store)
}

func TestBoltstoreValidateApplyResourcesTests(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err)
	defer cleanupTestDB(t)

	store := NewBoltStore(db, "super-secret-key", zap.NewNop())
	runValidateApplyResourcesTests(t, store)
}

func TestInitDB(t *testing.T) {
	cases := []struct {
		name      string
		setupFunc func() (string, error)
		errStr    string
	}{
		{
			"valid_path",
			func() (string, error) {
				return ioutil.TempDir("./", "tmp_store_test")
			},
			"",
		},
		{
			"invalid_path",
			func() (string, error) {
				return "not/valid/path", nil
			},
			"error while opening bbolt storage file: not/valid/path/bindplane.db, open not/valid/path/bindplane.db: no such file or directory",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup the test directory
			dir, err := tc.setupFunc()
			require.NoError(t, err, "failed to initialize test case")
			defer os.RemoveAll(dir)

			// Begin test
			path := path.Join(dir, "bindplane.db")
			db, err := InitDB(path)
			if tc.errStr != "" {
				require.Error(t, err)
				require.Equal(t, tc.errStr, err.Error())
				return
			}
			require.NoError(t, err, "did not expect an error while creating database at path %s", path)
			require.NotNil(t, db)
			require.Equal(t, path, db.Path())
			require.Equal(t, fmt.Sprintf("DB<\"%s\">", path), db.String())
			require.False(t, db.IsReadOnly(), "expected the boltstore to be read write")
			require.NoError(t, db.Close())

			// cursor count increases by 2 for every empty bucket created
			// a count of 6 means we have three buckets.
			bucketCount := 3
			require.Equal(t, bucketCount*2, db.Stats().TxStats.CursorCount)

			// InitDB creates three buckets: Resources, Tasks, Agents
			_ = db.Update(func(tx *bbolt.Tx) error {
				for _, bucket := range []string{bucketResources, bucketTasks, bucketAgents} {
					// Deleting the bucket
					err := tx.DeleteBucket([]byte(bucket))
					require.NoError(t, err, "expected bucket %s to exist", bucket)
				}
				return nil
			})
		})
	}
}

func TestNewBoltStore(t *testing.T) {
	cases := []struct {
		name string
	}{
		{
			"valid",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup bbolt
			dir, err := ioutil.TempDir("./", "tmp_store_test")
			require.NoError(t, err, "failed to initialize test directory")
			defer os.RemoveAll(dir)
			db, err := InitDB(path.Join(dir, "bindplane.db"))
			require.NoError(t, err, "failed to initialize test bbolt, got error")
			require.NotNil(t, db, "failed to initialize test bbolt, is nil")

			// Test
			output := NewBoltStore(db, "super-secret-key", nil)
			require.NotNil(t, output)
			require.IsType(t, &boltstore{}, output)
			require.Equal(t, db, output.(*boltstore).db)
			require.Equal(t, 0, output.Updates().Subscribers())
			require.Nil(t, output.(*boltstore).logger)
		})
	}
}
func TestBucketNames(t *testing.T) {
	require.Equal(t, "Resources", bucketResources)
	require.Equal(t, "Tasks", bucketTasks)
	require.Equal(t, "Agents", bucketAgents)
}

func TestBoltstoreDependentResources(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err)
	defer cleanupTestDB(t)

	store := NewBoltStore(db, "super-secret-key", zap.NewNop())
	runDependentResourcesTests(t, store)
}

func TestBoltstoreIndividualDelete(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err)
	defer cleanupTestDB(t)

	store := NewBoltStore(db, "super-secret-key", zap.NewNop())
	runIndividualDeleteTests(t, store)
}

func TestBoltstorePaging(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err)
	defer cleanupTestDB(t)

	store := NewBoltStore(db, "super-secret-key", zap.NewNop())
	runPagingTests(t, store)
}

func TestBoltStoreDeleteAgents(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err)
	defer cleanupTestDB(t)

	store := NewBoltStore(db, "super-secret-key", zap.NewNop())
	runDeleteAgentsTests(t, store)
}

func TestBoltstoreUpsertAgents(t *testing.T) {
	db, err := initTestDB(t)
	require.NoError(t, err)
	defer cleanupTestDB(t)

	store := NewBoltStore(db, "super-secret-key", zap.NewNop())
	runTestUpsertAgents(t, store)
}

/* ------------------------ SETUP + HELPER FUNCTIONS ------------------------ */

func initTestDB(t *testing.T) (*bbolt.DB, error) {
	db, err := bbolt.Open(testStorageFile(t), 0666, nil)
	require.NoError(t, err, "error while opening test database", err)

	// make sure buckets exists
	return db, db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketResources))
		require.NoError(t, err, "error while initializing test database, %w", err)
		_, err = tx.CreateBucketIfNotExists([]byte(bucketTasks))
		require.NoError(t, err, "error while initializing test database, %w", err)
		_, err = tx.CreateBucketIfNotExists([]byte(bucketAgents))
		require.NoError(t, err, "error while initializing test database, %w", err)

		return nil
	})
}

func cleanupTestDB(t *testing.T) {
	err := os.Remove(testStorageFile(t))
	require.NoError(t, err, "error while cleaning up test database, %w", err)
}

func testStorageFile(t *testing.T) string {
	exPath, err := os.Getwd()
	require.NoError(t, err, "error while finding the current directory for the test storage")

	dir := filepath.Dir(exPath)
	return path.Join(dir, "test-storage")
}

func storedAgentIDs(t *testing.T, store Store) []string {
	result := []string{}

	agents, err := store.Agents(context.TODO())
	require.NoError(t, err)

	for _, agent := range agents {
		result = append(result, agent.ID)
	}
	return result
}
