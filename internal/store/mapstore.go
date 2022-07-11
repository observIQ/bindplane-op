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

package store

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/observiq/bindplane-op/internal/eventbus"
	"github.com/observiq/bindplane-op/internal/store/search"
	"github.com/observiq/bindplane-op/model"
)

type mapStore struct {
	agents map[string]*model.Agent

	configurations   resourceStore[*model.Configuration]
	sources          resourceStore[*model.Source]
	sourceTypes      resourceStore[*model.SourceType]
	destinations     resourceStore[*model.Destination]
	destinationTypes resourceStore[*model.DestinationType]

	updates            *storeUpdates
	agentIndex         search.Index
	configurationIndex search.Index
	logger             *zap.Logger
	sync.RWMutex

	sessionStore sessions.Store
}

var _ Store = (*mapStore)(nil)

// NewMapStore returns an in memory Store
func NewMapStore(ctx context.Context, options Options, logger *zap.Logger) Store {
	return &mapStore{
		agents:             make(map[string]*model.Agent),
		configurations:     newResourceStore[*model.Configuration](),
		sources:            newResourceStore[*model.Source](),
		sourceTypes:        newResourceStore[*model.SourceType](),
		destinations:       newResourceStore[*model.Destination](),
		destinationTypes:   newResourceStore[*model.DestinationType](),
		updates:            newStoreUpdates(ctx, options.MaxEventsToMerge),
		agentIndex:         search.NewInMemoryIndex("agent"),
		configurationIndex: search.NewInMemoryIndex("configuration"),
		logger:             logger,
		sessionStore:       newBPCookieStore(options.SessionsSecret),
	}
}

// ----------------------------------------------------------------------
//
// resourceStore stores a single type of resource and has its own lock
type resourceStore[T model.Resource] struct {
	store map[string]T
	mtx   sync.RWMutex
}

func newResourceStore[T model.Resource]() resourceStore[T] {
	return resourceStore[T]{
		store: map[string]T{},
	}
}

func (r *resourceStore[T]) get(name string) T {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	return r.store[name]
}

func (r *resourceStore[T]) add(resource T) *model.ResourceStatus {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// generate a uuid if none supplied
	if resource.ID() == "" {
		resource.SetID(uuid.NewString())
	}
	existing, ok := r.store[resource.Name()]
	if ok && existing.ID() != "" {
		resource.SetID(existing.ID())
	}

	r.store[resource.Name()] = resource

	var status model.UpdateStatus
	switch {
	case !ok:
		status = model.StatusCreated
	case !resourcesEqual(existing, resource):
		status = model.StatusConfigured
	default:
		status = model.StatusUnchanged
	}

	return model.NewResourceStatus(resource, status)
}

func (r *resourceStore[T]) remove(name string) (item T, exists bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	existing, ok := r.store[name]
	if ok {
		delete(r.store, name)
	}
	return existing, ok
}

func (r *resourceStore[T]) removeAndNotify(name string, store *mapStore) (item T, exists bool, err error) {
	r.mtx.Lock()
	existing, ok := r.store[name]

	if ok {
		dependencies, err := FindDependentResources(context.TODO(), store, existing)
		if err != nil {
			r.mtx.Unlock()
			return existing, ok, err
		}

		if !dependencies.empty() {
			r.mtx.Unlock()
			return existing, ok, newDependencyError(dependencies)
		}
		delete(r.store, name)
	}
	r.mtx.Unlock()

	if ok {
		updates := NewUpdates()
		updates.IncludeResource(existing, EventTypeRemove)
		store.notify(updates)
	}

	return existing, ok, nil
}

func (r *resourceStore[T]) list() []T {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	return maps.Values(r.store)
}

func (r *resourceStore[T]) clear() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.store = map[string]T{}
}

// ----------------------------------------------------------------------

func (mapstore *mapStore) Clear() {
	mapstore.Lock()
	defer mapstore.Unlock()

	mapstore.agents = make(map[string]*model.Agent)

	mapstore.configurations.clear()
	mapstore.sources.clear()
	mapstore.sourceTypes.clear()
	mapstore.destinations.clear()
	mapstore.destinationTypes.clear()
}

func (mapstore *mapStore) UpsertAgents(ctx context.Context, agentIDs []string, updater AgentUpdater) ([]*model.Agent, error) {
	mapstore.Lock()
	defer mapstore.Unlock()

	agents := make([]*model.Agent, 0, len(agentIDs))
	u := NewUpdates()

	for _, id := range agentIDs {
		agents = append(agents, mapstore.upsertAgent(id, updater, u))
	}

	mapstore.notify(u)

	return agents, nil
}

func (mapstore *mapStore) UpsertAgent(ctx context.Context, agentID string, updater AgentUpdater) (*model.Agent, error) {
	mapstore.Lock()
	defer mapstore.Unlock()

	u := NewUpdates()
	agent := mapstore.upsertAgent(agentID, updater, u)

	mapstore.notify(u)

	return agent, nil
}

func (mapstore *mapStore) Agents(ctx context.Context, options ...QueryOption) ([]*model.Agent, error) {
	mapstore.RLock()
	defer mapstore.RUnlock()

	opts := makeQueryOptions(options)
	result := make([]*model.Agent, 0, len(mapstore.agents))

	for _, value := range mapstore.agents {
		if opts.selector.Matches(value.Labels) {
			result = append(result, value)
		}
	}

	if opts.sort == "" {
		opts.sort = "name"
	}
	return applySortOffsetAndLimit(result, opts, func(field string, item *model.Agent) string {
		switch field {
		case "id":
			return item.ID
		default:
			return item.Name
		}
	}), nil
}

func (mapstore *mapStore) AgentsCount(ctx context.Context, options ...QueryOption) (int, error) {
	agents, err := mapstore.Agents(ctx, options...)
	if err != nil {
		return -1, err
	}
	return len(agents), nil
}

func (mapstore *mapStore) Agent(id string) (*model.Agent, error) {
	mapstore.RLock()
	defer mapstore.RUnlock()
	return mapstore.agents[id], nil
}

func (mapstore *mapStore) DeleteAgents(ctx context.Context, agentIDs []string) ([]*model.Agent, error) {
	deleted := make([]*model.Agent, 0, len(agentIDs))
	updates := NewUpdates()

	mapstore.Lock()
	defer mapstore.Unlock()

	for _, id := range agentIDs {
		if agent, ok := mapstore.agents[id]; ok {
			// set status deleted
			agent.Status = 5

			// save the agent to return
			deleted = append(deleted, agent)

			// delete the agent
			delete(mapstore.agents, id)

			// include in the agent updates
			updates.Agents.Include(agent, EventTypeRemove)

			// remove from the search index
			if err := mapstore.agentIndex.Remove(agent); err != nil {
				mapstore.logger.Error("failed to remove agent from the search index", zap.String("agentID", agent.ID))
			}
		}
	}

	mapstore.notify(updates)

	return deleted, nil
}

func (mapstore *mapStore) Configurations(options ...QueryOption) ([]*model.Configuration, error) {
	opts := makeQueryOptions(options)
	config := mapstore.configurations.list()
	if opts.sort == "" {
		opts.sort = "name"
	}
	return applySortOffsetAndLimit(config, opts, func(field string, item *model.Configuration) string {
		// we currently only support sorting by name
		return item.Name()
	}), nil
}
func (mapstore *mapStore) Configuration(name string) (*model.Configuration, error) {
	return mapstore.configurations.get(name), nil
}
func (mapstore *mapStore) DeleteConfiguration(name string) (*model.Configuration, error) {
	item, exists, err := mapstore.configurations.removeAndNotify(name, mapstore)
	if err != nil {
		return item, err
	}

	if !exists {
		return nil, nil
	}
	return item, nil
}

func (mapstore *mapStore) Source(name string) (*model.Source, error) {
	return mapstore.sources.get(name), nil
}
func (mapstore *mapStore) Sources() ([]*model.Source, error) {
	return mapstore.sources.list(), nil
}
func (mapstore *mapStore) DeleteSource(name string) (*model.Source, error) {
	item, exists, err := mapstore.sources.removeAndNotify(name, mapstore)
	if err != nil {
		return item, err
	}

	if !exists {
		return nil, nil
	}
	return item, nil
}

func (mapstore *mapStore) SourceType(name string) (*model.SourceType, error) {
	return mapstore.sourceTypes.get(name), nil
}
func (mapstore *mapStore) SourceTypes() ([]*model.SourceType, error) {
	return mapstore.sourceTypes.list(), nil
}
func (mapstore *mapStore) DeleteSourceType(name string) (*model.SourceType, error) {
	item, exists, err := mapstore.sourceTypes.removeAndNotify(name, mapstore)
	if err != nil {
		return item, err
	}

	if !exists {
		return nil, nil
	}
	return item, nil
}

func (mapstore *mapStore) Destination(name string) (*model.Destination, error) {
	return mapstore.destinations.get(name), nil
}
func (mapstore *mapStore) Destinations() ([]*model.Destination, error) {
	return mapstore.destinations.list(), nil
}
func (mapstore *mapStore) DeleteDestination(name string) (*model.Destination, error) {
	item, exists, err := mapstore.destinations.removeAndNotify(name, mapstore)
	if err != nil {
		return item, err
	}

	if !exists {
		return nil, nil
	}
	return item, nil
}

func (mapstore *mapStore) DestinationType(name string) (*model.DestinationType, error) {
	return mapstore.destinationTypes.get(name), nil
}
func (mapstore *mapStore) DestinationTypes() ([]*model.DestinationType, error) {
	return mapstore.destinationTypes.list(), nil
}
func (mapstore *mapStore) DeleteDestinationType(name string) (*model.DestinationType, error) {
	item, exists, err := mapstore.destinationTypes.removeAndNotify(name, mapstore)
	if err != nil {
		return item, err
	}

	if !exists {
		return nil, nil
	}
	return item, nil
}

func (mapstore *mapStore) ApplyResources(resources []model.Resource) ([]model.ResourceStatus, error) {
	mapstore.Lock()
	defer mapstore.Unlock()
	var result error

	updates := NewUpdates()
	resourceStatuses := make([]model.ResourceStatus, 0)

	for _, resource := range resources {
		err := resource.ValidateWithStore(mapstore)
		if err != nil {
			resourceStatuses = append(resourceStatuses, *model.NewResourceStatusWithReason(resource, model.StatusInvalid, err.Error()))
			continue
		}

		var resourceStatus *model.ResourceStatus
		switch r := resource.(type) {
		case *model.Configuration:
			resourceStatus = mapstore.configurations.add(r)
			if err := mapstore.configurationIndex.Upsert(resourceStatus.Resource); err != nil {
				mapstore.logger.Error("error updating configuration in the search index", zap.Error(err))
			}
		case *model.Source:
			resourceStatus = mapstore.sources.add(r)
		case *model.SourceType:
			resourceStatus = mapstore.sourceTypes.add(r)
		case *model.Destination:
			resourceStatus = mapstore.destinations.add(r)
		case *model.DestinationType:
			resourceStatus = mapstore.destinationTypes.add(r)
		default:
			resourceStatus = model.NewResourceStatusWithReason(resource, model.StatusInvalid, fmt.Sprintf("unknown resource type in apply: %s", r.Name()))
		}

		if resourceStatus != nil {
			resourceStatuses = append(resourceStatuses, *resourceStatus)

			switch resourceStatus.Status {
			case model.StatusCreated:
				updates.IncludeResource(resource, EventTypeInsert)
			case model.StatusConfigured:
				updates.IncludeResource(resource, EventTypeUpdate)
			}
		}
	}

	mapstore.notify(updates)
	return resourceStatuses, result
}

func (mapstore *mapStore) DeleteResources(resources []model.Resource) ([]model.ResourceStatus, error) {
	mapstore.Lock()
	defer mapstore.Unlock()

	// Send the pipeline deletes, even if its an empty map to satisfy tests
	updates := NewUpdates()

	resourceStatuses := make([]model.ResourceStatus, 0)

	for _, r := range resources {
		dependencies, err := FindDependentResources(context.TODO(), mapstore, r)
		if err != nil {
			mapstore.logger.Error("failed to get dependent resources", zap.Error(err))
			continue
		}

		if !dependencies.empty() {
			resourceStatuses = append(resourceStatuses,
				*model.NewResourceStatusWithReason(
					r,
					model.StatusInUse,
					dependencies.message(),
				),
			)
			continue
		}

		var exists bool
		switch r := r.(type) {
		case *model.Configuration:
			c, e := mapstore.configurations.remove(r.Name())
			if err := mapstore.configurationIndex.Remove(c); err != nil {
				mapstore.logger.Error("error removing configuration from the search index", zap.Error(err))
			}
			exists = e

		case *model.Source:
			_, exists = mapstore.sources.remove(r.Name())

		case *model.SourceType:
			_, exists = mapstore.sourceTypes.remove(r.Name())

		case *model.Destination:
			_, exists = mapstore.destinations.remove(r.Name())

		case *model.DestinationType:
			_, exists = mapstore.destinationTypes.remove(r.Name())

		default:
			continue
		}
		if exists {
			resourceStatuses = append(resourceStatuses, *model.NewResourceStatus(r, model.StatusDeleted))
			updates.IncludeResource(r, EventTypeRemove)
		}
	}

	mapstore.notify(updates)
	return resourceStatuses, nil
}

// AgentConfiguration returns the configuration that should be applied to an agent.
func (mapstore *mapStore) AgentConfiguration(agentID string) (*model.Configuration, error) {
	mapstore.RLock()
	defer mapstore.RUnlock()

	agent, err := mapstore.Agent(agentID)
	if err != nil {
		return nil, fmt.Errorf("cannot return configuration for unknown agent: %w", err)
	}

	labels := agent.Labels

	// look through all of the configurations and check their selector to see if they match this agent. there are more
	// efficient implementations, but this is fine for mapstore.
	for _, c := range mapstore.configurations.store {
		if c.AgentSelector().Matches(labels) {
			return c, nil
		}
	}

	return nil, nil
}

// AgentsIDsMatchingConfiguration returns the list of agent IDs that are using the specified configuration
func (mapstore *mapStore) AgentsIDsMatchingConfiguration(configuration *model.Configuration) ([]string, error) {
	ids := mapstore.agentIndex.Select(configuration.Spec.Selector.MatchLabels)
	return ids, nil
}

func (mapstore *mapStore) Updates() eventbus.Source[*Updates] {
	return mapstore.updates.Updates()
}

// CleanupDisconnectedAgents removes agents that have disconnected before the specified time
func (mapstore *mapStore) CleanupDisconnectedAgents(since time.Time) error {
	mapstore.Lock()
	defer mapstore.Unlock()

	updates := NewUpdates()

	for _, agent := range mapstore.agents {
		if agent.DisconnectedSince(since) {
			delete(mapstore.agents, agent.ID)
			updates.IncludeAgent(agent, EventTypeRemove)
		}
	}

	mapstore.notify(updates)

	return nil
}

// Index provides access to the search Index implementation managed by the Store
func (mapstore *mapStore) AgentIndex() search.Index {
	return mapstore.agentIndex
}

// ConfigurationIndex provides access to the search Index for Configurations
func (mapstore *mapStore) ConfigurationIndex() search.Index {
	return mapstore.configurationIndex
}

func (mapstore *mapStore) UserSessions() sessions.Store {
	return mapstore.sessionStore
}

// ----------------------------------------------------------------------
// these functions require that the mapstore is already locked

// ----------------------------------------------------------------------

// upsertAgent updates the agent with given id while the mapstore is locked.
// it updates the passed *Updates to include the change.
func (mapstore *mapStore) upsertAgent(agentID string, updater AgentUpdater, updates *Updates) *model.Agent {
	agentEventType := EventTypeInsert

	agent := &model.Agent{ID: agentID}
	if curAgent, ok := mapstore.agents[agentID]; ok {
		agent = curAgent
		agentEventType = EventTypeUpdate
	}

	// compare labels before/after and notify if they change
	labelsBefore := agent.Labels.String()

	updater(agent)
	mapstore.agents[agentID] = agent

	// update the index
	err := mapstore.agentIndex.Upsert(agent)
	if err != nil {
		mapstore.logger.Error("failed to update the search index", zap.String("agentID", agent.ID))
	}

	labelsAfter := agent.Labels.String()
	if labelsBefore != "" && labelsAfter != labelsBefore {
		agentEventType = EventTypeLabel
	}

	updates.IncludeAgent(agent, agentEventType)
	return agent
}

func (mapstore *mapStore) notify(updates *Updates) {
	err := updates.addTransitiveUpdates(mapstore)
	if err != nil {
		// TODO: if we can't notify about all updates, what do we do?
		mapstore.logger.Error("unable to add transitive updates", zap.Any("updates", updates), zap.Error(err))
	}
	if !updates.Empty() {
		mapstore.updates.Send(updates)
	}
}

func resourcesEqual(r1 model.Resource, r2 model.Resource) bool {
	r1Any := &model.AnyResource{}
	r2Any := &model.AnyResource{}
	err := mapstructure.Decode(r1, r1Any)
	if err != nil {
		return false
	}
	err = mapstructure.Decode(r2, r2Any)
	if err != nil {
		return false
	}

	r1Any.Metadata.ID = ""
	r2Any.Metadata.ID = ""
	return reflect.DeepEqual(r1Any, r2Any)
}
