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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/sessions"
	"github.com/hashicorp/go-multierror"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/internal/eventbus"
	"github.com/observiq/bindplane-op/internal/store/search"
	"github.com/observiq/bindplane-op/model"
)

// bucket names
const (
	bucketResources = "Resources"
	bucketTasks     = "Tasks"
	bucketAgents    = "Agents"
)

type boltstore struct {
	db                 *bbolt.DB
	updates            *storeUpdates
	agentIndex         search.Index
	configurationIndex search.Index
	logger             *zap.Logger
	sync.RWMutex
	sessionStorage sessions.Store
}

var _ Store = (*boltstore)(nil)

// NewBoltStore returns a new store boltstore struct that implements the store.Store interface.
func NewBoltStore(ctx context.Context, db *bbolt.DB, options Options, logger *zap.Logger) Store {
	store := &boltstore{
		db:                 db,
		updates:            newStoreUpdates(ctx, options.MaxEventsToMerge),
		agentIndex:         search.NewInMemoryIndex("agent"),
		configurationIndex: search.NewInMemoryIndex("configuration"),
		logger:             logger,

		sessionStorage: newBPCookieStore(options.SessionsSecret),
	}

	// boltstore is not used for clusters, disconnect all agents
	store.disconnectAllAgents(context.Background())

	return store
}

// InitDB takes in the full path to a storage file and returns an opened bbolt database.
// It will return an error if the file cannot be opened.
func InitDB(storageFilePath string) (*bbolt.DB, error) {
	var db, err = bbolt.Open(storageFilePath, 0640, nil)
	if err != nil {
		return nil, fmt.Errorf("error while opening bbolt storage file: %s, %w", storageFilePath, err)
	}

	buckets := []string{
		bucketResources,
		bucketTasks,
		bucketAgents,
	}

	// make sure buckets exists, errors are ignored here because bucket names are
	// all valid, non-empty strings.
	_ = db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range buckets {
			_, _ = tx.CreateBucketIfNotExists([]byte(bucket))
		}
		return nil
	})

	return db, nil
}

// AgentConfiguration returns the configuration that should be applied to an agent.
func (s *boltstore) AgentConfiguration(agentID string) (*model.Configuration, error) {
	if agentID == "" {
		return nil, fmt.Errorf("cannot return configuration for empty agentID")
	}

	agent, err := s.Agent(agentID)
	if agent == nil {
		return nil, nil
	}

	// check for configuration= label and use that
	if configurationName, ok := agent.Labels.Set["configuration"]; ok {
		// if there is a configuration label, this takes precedence and we don't need to look any further
		return s.Configuration(configurationName)
	}

	var result *model.Configuration

	err = s.db.View(func(tx *bbolt.Tx) error {
		// iterate over the configurations looking for one that applies
		prefix := []byte(model.KindConfiguration)
		cursor := resourcesBucket(tx).Cursor()

		for k, v := cursor.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = cursor.Next() {
			configuration := &model.Configuration{}
			if err := json.Unmarshal(v, configuration); err != nil {
				s.logger.Error("unable to unmarshal configuration, ignoring", zap.Error(err))
				continue
			}
			if configuration.IsForAgent(agent) {
				result = configuration
				break
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve agent configuration: %w", err)
	}

	return result, nil
}

// AgentsIDsMatchingConfiguration returns the list of agent IDs that are using the specified configuration
func (s *boltstore) AgentsIDsMatchingConfiguration(configuration *model.Configuration) ([]string, error) {
	ids := s.AgentIndex().Select(configuration.Spec.Selector.MatchLabels)
	return ids, nil
}

func (s *boltstore) Updates() eventbus.Source[*Updates] {
	return s.updates.Updates()
}

// DeleteResources iterates threw a slice of resources, and removes them from storage by name.
// Sends any successful pipeline deletes to the pipelineDeletes channel, to be handled by the manager.
// Exporter and receiver deletes are sent to the manager via notifyUpdates.
func (s *boltstore) DeleteResources(resources []model.Resource) ([]model.ResourceStatus, error) {
	updates := NewUpdates()

	// track deleteStatuses to return
	deleteStatuses := make([]model.ResourceStatus, 0)

	for _, r := range resources {
		empty, err := model.NewEmptyResource(r.GetKind())
		if err != nil {
			deleteStatuses = append(deleteStatuses, *model.NewResourceStatusWithReason(r, model.StatusError, err.Error()))
			continue
		}

		deleted, exists, err := deleteResource(s, r.GetKind(), r.Name(), empty)

		switch err.(type) {
		case *DependencyError:
			deleteStatuses = append(
				deleteStatuses,
				*model.NewResourceStatusWithReason(r, model.StatusInUse, err.Error()))
			continue

		case nil:
			break

		default:
			deleteStatuses = append(deleteStatuses, *model.NewResourceStatusWithReason(r, model.StatusError, err.Error()))
			continue
		}

		if !exists {
			continue
		}

		deleteStatuses = append(deleteStatuses, *model.NewResourceStatus(r, model.StatusDeleted))
		updates.IncludeResource(deleted, EventTypeRemove)
	}

	s.notify(updates)
	return deleteStatuses, nil
}

// Apply resources iterates through a slice of resources, then adds them to storage,
// and calls notify updates on the updated resources.
func (s *boltstore) ApplyResources(resources []model.Resource) ([]model.ResourceStatus, error) {
	updates := NewUpdates()

	// resourceStatuses to return for the applied resources
	resourceStatuses := make([]model.ResourceStatus, 0)

	var errs error
	for _, resource := range resources {
		// Set the resource's initial ID, which wil be overwritten if
		// the resource already exists (using the existing resource ID)
		resource.EnsureID()

		err := resource.ValidateWithStore(s)
		if err != nil {
			resourceStatuses = append(resourceStatuses, *model.NewResourceStatusWithReason(resource, model.StatusInvalid, err.Error()))
			continue
		}

		err = s.db.Update(func(tx *bbolt.Tx) error {
			// update the resource in the database
			status, err := upsertResource(tx, resource, resource.GetKind())
			if err != nil {
				resourceStatuses = append(resourceStatuses, *model.NewResourceStatusWithReason(resource, model.StatusError, err.Error()))
				return err
			}
			resourceStatuses = append(resourceStatuses, *model.NewResourceStatus(resource, status))

			switch status {
			case model.StatusCreated:
				updates.IncludeResource(resource, EventTypeInsert)
			case model.StatusConfigured:
				updates.IncludeResource(resource, EventTypeUpdate)
			}

			// some resources need special treatment
			switch r := resource.(type) {
			case *model.Configuration:
				// update the index
				err = s.configurationIndex.Upsert(r)
				if err != nil {
					s.logger.Error("failed to update the search index", zap.String("configuration", r.Name()))
				}
			}
			return nil
		})
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	s.notify(updates)

	return resourceStatuses, errs
}

// ----------------------------------------------------------------------

func (s *boltstore) notify(updates *Updates) {
	err := updates.addTransitiveUpdates(s)
	if err != nil {
		// TODO: if we can't notify about all updates, what do we do?
		s.logger.Error("unable to add transitive updates", zap.Any("updates", updates), zap.Error(err))
	}
	if !updates.Empty() {
		s.updates.Send(updates)
	}
}

// ----------------------------------------------------------------------

// Clear clears the db store of resources, agents, and tasks.  Mostly used for testing.
func (s *boltstore) Clear() {
	// Disregarding error from update because these actions errors are known and prevented
	_ = s.db.Update(func(tx *bbolt.Tx) error {
		// Delete all the buckets.
		// Disregarding errors because it will only error if the bucket doesn't exist
		// or isn't a bucket key - which we're confident its not.
		_ = tx.DeleteBucket([]byte(bucketResources))
		_ = tx.DeleteBucket([]byte(bucketTasks))
		_ = tx.DeleteBucket([]byte(bucketAgents))

		// create them again
		// Disregarding errors because bucket names are valid.
		_, _ = tx.CreateBucketIfNotExists([]byte(bucketResources))
		_, _ = tx.CreateBucketIfNotExists([]byte(bucketTasks))
		_, _ = tx.CreateBucketIfNotExists([]byte(bucketAgents))
		return nil
	})
}

func (s *boltstore) UpsertAgents(ctx context.Context, agentIDs []string, updater AgentUpdater) ([]*model.Agent, error) {
	ctx, span := tracer.Start(ctx, "store/UpsertAgents")
	defer span.End()

	agents := make([]*model.Agent, 0, len(agentIDs))
	updates := NewUpdates()

	err := s.db.Update(func(tx *bbolt.Tx) error {
		for _, agentID := range agentIDs {
			agent, err := upsertAgentTx(tx, agentID, updater, updates)
			if err != nil {
				return err
			}

			agents = append(agents, agent)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// update the search index with changes
	for _, a := range agents {
		if err := s.agentIndex.Upsert(a); err != nil {
			s.logger.Error("failed to update the search index", zap.String("agentID", a.ID))
		}
	}

	// notify results
	s.notify(updates)
	return agents, nil
}

// UpsertAgent creates or updates the given agent and calls the updater method on it.
func (s *boltstore) UpsertAgent(ctx context.Context, id string, updater AgentUpdater) (*model.Agent, error) {
	ctx, span := tracer.Start(ctx, "store/UpsertAgent")
	defer span.End()

	var updatedAgent *model.Agent
	updates := NewUpdates()

	err := s.db.Update(func(tx *bbolt.Tx) error {
		agent, err := upsertAgentTx(tx, id, updater, updates)
		updatedAgent = agent
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// update the index
	err = s.agentIndex.Upsert(updatedAgent)
	if err != nil {
		s.logger.Error("failed to update the search index", zap.String("agentID", updatedAgent.ID))
	}

	s.notify(updates)

	return updatedAgent, nil
}

func (s *boltstore) Agents(ctx context.Context, options ...QueryOption) ([]*model.Agent, error) {
	opts := makeQueryOptions(options)

	// search is implemented using the search index
	if opts.query != nil {
		ids, err := s.agentIndex.Search(ctx, opts.query)
		if err != nil {
			return nil, err
		}
		return s.agentsByID(ids, opts)
	}

	agents := []*model.Agent{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		cursor := agentBucket(tx).Cursor()
		prefix := []byte("Agent")

		for k, v := cursor.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = cursor.Next() {
			agent := &model.Agent{}
			if err := json.Unmarshal(v, agent); err != nil {
				return fmt.Errorf("agents: %w", err)
			}

			if opts.selector.Matches(agent.Labels) {
				agents = append(agents, agent)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if opts.sort == "" {
		opts.sort = "name"
	}
	return applySortOffsetAndLimit(agents, opts, func(field string, item *model.Agent) string {
		switch field {
		case "id":
			return item.ID
		default:
			return item.Name
		}
	}), nil
}

func (s *boltstore) DeleteAgents(ctx context.Context, agentIDs []string) ([]*model.Agent, error) {
	updates := NewUpdates()
	deleted := make([]*model.Agent, 0, len(agentIDs))

	err := s.db.Update(func(tx *bbolt.Tx) error {
		c := agentBucket(tx).Cursor()

		for _, id := range agentIDs {
			agentKey := agentKey(id)
			k, v := c.Seek(agentKey)

			if k != nil && bytes.Equal(k, agentKey) {

				// Save the agent to return and set its status to deleted.
				agent := &model.Agent{}
				err := json.Unmarshal(v, agent)
				if err != nil {
					return err
				}

				agent.Status = 5
				deleted = append(deleted, agent)

				// delete it
				err = c.Delete()
				if err != nil {
					return err
				}

				// include it in updates
				updates.IncludeAgent(agent, EventTypeRemove)
			}
		}

		return nil
	})

	if err != nil {
		return deleted, err
	}

	// remove deleted agents from the index
	for _, agent := range deleted {
		if err := s.agentIndex.Remove(agent); err != nil {
			s.logger.Error("failed to remove from the search index", zap.String("agentID", agent.ID))
		}
	}

	// notify updates
	s.notify(updates)

	return deleted, nil
}

func (s *boltstore) agentsByID(ids []string, opts queryOptions) ([]*model.Agent, error) {
	var agents []*model.Agent

	err := s.db.View(func(tx *bbolt.Tx) error {
		for _, id := range ids {
			data := agentBucket(tx).Get(agentKey(id))
			if data == nil {
				return nil
			}
			agent := &model.Agent{}
			if err := json.Unmarshal(data, agent); err != nil {
				return fmt.Errorf("agents: %w", err)
			}

			if opts.selector.Matches(agent.Labels) {
				agents = append(agents, agent)
			}
		}
		return nil
	})

	return agents, err
}

func (s *boltstore) AgentsCount(ctx context.Context, options ...QueryOption) (int, error) {
	agents, err := s.Agents(ctx, options...)
	if err != nil {
		return -1, err
	}
	return len(agents), nil
}

func (s *boltstore) Agent(id string) (*model.Agent, error) {
	var agent *model.Agent

	err := s.db.View(func(tx *bbolt.Tx) error {
		data := agentBucket(tx).Get(agentKey(id))
		if data == nil {
			return nil
		}
		agent = &model.Agent{}
		return json.Unmarshal(data, agent)
	})

	return agent, err
}

func (s *boltstore) Configurations(options ...QueryOption) ([]*model.Configuration, error) {
	opts := makeQueryOptions(options)
	// search is implemented using the search index
	if opts.query != nil {
		names, err := s.configurationIndex.Search(context.TODO(), opts.query)
		if err != nil {
			return nil, err
		}
		return resourcesByName[*model.Configuration](s, model.KindConfiguration, names, opts)
	}

	return resourcesWithFilter(s, model.KindConfiguration, func(c *model.Configuration) bool {
		return opts.selector.Matches(c.GetLabels())
	})
}
func (s *boltstore) Configuration(name string) (*model.Configuration, error) {
	item, exists, err := resource[*model.Configuration](s, model.KindConfiguration, name)
	if !exists {
		item = nil
	}
	return item, err
}
func (s *boltstore) DeleteConfiguration(name string) (*model.Configuration, error) {
	item, exists, err := deleteResourceAndNotify(s, model.KindConfiguration, name, &model.Configuration{})
	if !exists {
		return nil, err
	}
	return item, err
}

func (s *boltstore) Source(name string) (*model.Source, error) {
	item, exists, err := resource[*model.Source](s, model.KindSource, name)
	if !exists {
		item = nil
	}
	return item, err
}
func (s *boltstore) Sources() ([]*model.Source, error) {
	return resources[*model.Source](s, model.KindSource)
}
func (s *boltstore) DeleteSource(name string) (*model.Source, error) {
	item, exists, err := deleteResourceAndNotify(s, model.KindSource, name, &model.Source{})
	if !exists {
		return nil, err
	}
	return item, err
}

func (s *boltstore) SourceType(name string) (*model.SourceType, error) {
	item, exists, err := resource[*model.SourceType](s, model.KindSourceType, name)
	if !exists {
		item = nil
	}
	return item, err
}
func (s *boltstore) SourceTypes() ([]*model.SourceType, error) {
	return resources[*model.SourceType](s, model.KindSourceType)
}
func (s *boltstore) DeleteSourceType(name string) (*model.SourceType, error) {
	item, exists, err := deleteResourceAndNotify(s, model.KindSourceType, name, &model.SourceType{})
	if !exists {
		return nil, err
	}
	return item, err
}

func (s *boltstore) Destination(name string) (*model.Destination, error) {
	item, exists, err := resource[*model.Destination](s, model.KindDestination, name)
	if !exists {
		item = nil
	}
	return item, err
}
func (s *boltstore) Destinations() ([]*model.Destination, error) {
	return resources[*model.Destination](s, model.KindDestination)
}
func (s *boltstore) DeleteDestination(name string) (*model.Destination, error) {
	item, exists, err := deleteResourceAndNotify(s, model.KindDestination, name, &model.Destination{})
	if !exists {
		return nil, err
	}
	return item, err
}

func (s *boltstore) DestinationType(name string) (*model.DestinationType, error) {
	item, exists, err := resource[*model.DestinationType](s, model.KindDestinationType, name)
	if !exists {
		item = nil
	}
	return item, err
}
func (s *boltstore) DestinationTypes() ([]*model.DestinationType, error) {
	return resources[*model.DestinationType](s, model.KindDestinationType)
}
func (s *boltstore) DeleteDestinationType(name string) (*model.DestinationType, error) {
	item, exists, err := deleteResourceAndNotify(s, model.KindDestinationType, name, &model.DestinationType{})
	if !exists {
		return nil, err
	}
	return item, err
}

// CleanupDisconnectedAgents removes agents that have disconnected before the specified time
func (s *boltstore) CleanupDisconnectedAgents(since time.Time) error {
	agents, err := s.Agents(context.TODO())
	if err != nil {
		return err
	}
	changes := NewUpdates()

	for _, agent := range agents {
		if agent.DisconnectedSince(since) {
			err := s.db.Update(func(tx *bbolt.Tx) error {
				return agentBucket(tx).Delete(agentKey(agent.ID))
			})
			if err != nil {
				return err
			}
			changes.IncludeAgent(agent, EventTypeRemove)

			// update the index
			if err := s.agentIndex.Remove(agent); err != nil {
				s.logger.Error("failed to remove from the search index", zap.String("agentID", agent.ID))
			}
		}
	}

	s.notify(changes)
	return nil
}

// Index provides access to the search Index implementation managed by the Store
func (s *boltstore) AgentIndex() search.Index {
	return s.agentIndex
}

// ConfigurationIndex provides access to the search Index for Configurations
func (s *boltstore) ConfigurationIndex() search.Index {
	return s.configurationIndex
}

func (s *boltstore) UserSessions() sessions.Store {
	return s.sessionStorage
}

// ----------------------------------------------------------------------

func (s *boltstore) disconnectAllAgents(ctx context.Context) {
	if agents, err := s.Agents(ctx); err != nil {
		s.logger.Error("error while disconnecting all agents on startup", zap.Error(err))
	} else {
		s.logger.Info("disconnecting all agents on startup", zap.Int("count", len(agents)))
		for _, agent := range agents {
			_, err := s.UpsertAgent(ctx, agent.ID, func(a *model.Agent) {
				a.Disconnect()
			})
			if err != nil {
				s.logger.Error("error while disconnecting agent on startup", zap.Error(err))
			}
		}
	}
}

/* ---------------------------- helper functions ---------------------------- */
func resourcesPrefix(kind model.Kind) []byte {
	return []byte(fmt.Sprintf("%s|", kind))
}

func resourceKey(kind model.Kind, name string) []byte {
	return []byte(fmt.Sprintf("%s|%s", kind, name))
}

func agentKey(id string) []byte {
	return []byte(fmt.Sprintf("%s|%s", "Agent", id))
}

func agentBucket(tx *bbolt.Tx) *bbolt.Bucket {
	return tx.Bucket([]byte(bucketAgents))
}

func keyFromResource(r model.Resource) []byte {
	if r == nil || r.GetKind() == model.KindUnknown {
		return make([]byte, 0)
	}
	return resourceKey(r.GetKind(), r.Name())
}

/* --------------------------- transaction helpers -------------------------- */
/* ------- These helper functions happen inside of a bbolt transaction ------ */

func resourcesBucket(tx *bbolt.Tx) *bbolt.Bucket {
	return tx.Bucket([]byte(bucketResources))
}

func upsertResource(tx *bbolt.Tx, r model.Resource, kind model.Kind) (model.UpdateStatus, error) {
	key := resourceKey(kind, r.Name())
	bucket := resourcesBucket(tx)
	existing := bucket.Get(key)

	// preserve the id (if possible)
	if len(existing) > 0 {
		var cur model.AnyResource
		if err := json.Unmarshal(existing, &cur); err == nil {
			r.SetID(cur.ID())
		}
	}

	data, err := json.Marshal(r)
	if err != nil {
		// error, status unchanged
		return model.StatusUnchanged, fmt.Errorf("upsert resource: %w", err)
	}
	if bytes.Equal(existing, data) {
		return model.StatusUnchanged, nil
	}

	if err = bucket.Put(key, data); err != nil {
		// error, status unchanged
		return model.StatusUnchanged, fmt.Errorf("upsert resource: %w", err)
	}

	if len(existing) == 0 {
		return model.StatusCreated, nil
	}
	return model.StatusConfigured, nil
}

// upsertAgentTx is a transaction helper that updates the given agent,
// puts it into the agent bucket  and includes it in the passed updates.
// it does *not* update the search index or notify any subscribers of updates.
func upsertAgentTx(tx *bbolt.Tx, agentID string, updater AgentUpdater, updates *Updates) (*model.Agent, error) {
	bucket := agentBucket(tx)
	key := agentKey(agentID)

	agentEventType := EventTypeInsert
	agent := &model.Agent{ID: agentID}

	// load the existing agent or create it
	if data := bucket.Get(key); data != nil {
		// existing agent, unmarshal
		if err := json.Unmarshal(data, agent); err != nil {
			return agent, err
		}
		agentEventType = EventTypeUpdate
	}

	// compare labels before/after and notify if they change
	labelsBefore := agent.Labels.String()

	// update the agent
	updater(agent)

	labelsAfter := agent.Labels.String()

	// if the labels changes is this is just an update, use EventTypeLabel
	if labelsAfter != labelsBefore && agentEventType == EventTypeUpdate {
		agentEventType = EventTypeLabel
	}

	// marshal it back to to json
	data, err := json.Marshal(agent)
	if err != nil {
		return agent, err
	}

	err = bucket.Put(key, data)
	if err != nil {
		return agent, err
	}

	updates.IncludeAgent(agent, agentEventType)
	return agent, nil
}

// ----------------------------------------------------------------------
// generic resource accessors

func resource[R model.Resource](s *boltstore, kind model.Kind, name string) (resource R, exists bool, err error) {
	err = s.db.View(func(tx *bbolt.Tx) error {
		key := resourceKey(kind, name)
		data := resourcesBucket(tx).Get(key)
		if data == nil {
			return nil
		}
		exists = true
		return json.Unmarshal(data, &resource)
	})
	return resource, exists, err
}

func resources[R model.Resource](s *boltstore, kind model.Kind) ([]R, error) {
	return resourcesWithFilter[R](s, kind, nil)
}
func resourcesWithFilter[R model.Resource](s *boltstore, kind model.Kind, include func(R) bool) ([]R, error) {
	var resources []R

	err := s.db.View(func(tx *bbolt.Tx) error {
		prefix := resourcesPrefix(kind)
		cursor := resourcesBucket(tx).Cursor()

		for k, v := cursor.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = cursor.Next() {
			var resource R
			if err := json.Unmarshal(v, &resource); err != nil {
				// TODO(andy): if it can't be unmarshaled, it should probably be removed from the store. ignore it for now.
				s.logger.Error("failed to unmarshal resource", zap.String("key", string(k)), zap.String("kind", string(kind)), zap.Error(err))
				continue
			}
			if include == nil || include(resource) {
				resources = append(resources, resource)
			}
		}

		return nil
	})

	return resources, err
}

// resourcesByName returns the resources of the specified name with the specified names. If requesting some resources
// results in an error, the errors will be accumulated and return with the list of resources successfully retrieved.
func resourcesByName[R model.Resource](s *boltstore, kind model.Kind, names []string, opts queryOptions) ([]R, error) {
	var errs error
	var results []R

	for _, name := range names {
		if result, exists, err := resource[R](s, kind, name); err != nil {
			errs = multierror.Append(err, err)
		} else {
			if exists && opts.selector.Matches(result.GetLabels()) {
				results = append(results, result)
			}
		}
	}
	return results, errs
}

func deleteResourceAndNotify[R model.Resource](s *boltstore, kind model.Kind, name string, emptyResource R) (resource R, exists bool, err error) {
	deleted, exists, err := deleteResource(s, kind, name, emptyResource)

	if err == nil && exists {
		updates := NewUpdates()
		updates.IncludeResource(deleted, EventTypeRemove)
		s.notify(updates)
	}

	return deleted, exists, err
}

// deleteResource removes the resource with the given kind and name. Returns ResourceMissingError if the resource wasn't
// found. Returns DependencyError if the resource is referenced by another.
// emptyResource will be populated with the deleted resource. For convenience, if the delete is successful, the
// populated resource will also be returned. If there was an error, nil will be returned for the resource.
func deleteResource[R model.Resource](s *boltstore, kind model.Kind, name string, emptyResource R) (resource R, exists bool, err error) {
	var dependencies DependentResources

	err = s.db.Update(func(tx *bbolt.Tx) error {
		key := resourceKey(kind, name)

		c := resourcesBucket(tx).Cursor()
		k, v := c.Seek(key)

		if bytes.Equal(k, key) {
			// populate the emptyResource with the data before deleting
			err := json.Unmarshal(v, emptyResource)
			if err != nil {
				return err
			}

			exists = true

			// Check if the resources is referenced by another
			dependencies, err = FindDependentResources(context.TODO(), s, emptyResource)
			if !dependencies.empty() {
				return ErrResourceInUse
			}

			// Delete the key from the store
			return c.Delete()
		}

		return ErrResourceMissing
	})

	switch {
	case errors.Is(err, ErrResourceMissing):
		return resource, exists, nil
	case errors.Is(err, ErrResourceInUse):
		return emptyResource, exists, newDependencyError(dependencies)
	case err != nil:
		return resource, exists, err
	}

	if emptyResource.GetKind() == model.KindConfiguration {
		if err := s.configurationIndex.Remove(emptyResource); err != nil {
			s.logger.Error("failed to remove configuration from the search index", zap.String("name", emptyResource.Name()))
		}
	}

	return emptyResource, exists, nil
}
