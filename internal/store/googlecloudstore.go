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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/hashicorp/go-multierror"
	"github.com/observiq/bindplane/common"
	"github.com/observiq/bindplane/internal/eventbus"
	"github.com/observiq/bindplane/internal/store/search"
	"github.com/observiq/bindplane/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

var tracer = otel.Tracer("googlecloudstore")

type googleCloudStore struct {
	client             *datastore.Client
	pubsub             *pubsubClient
	updates            eventbus.Source[*Updates]
	agentIndex         search.Index
	configurationIndex search.Index
	logger             *zap.Logger

	sessionStore sessions.Store
}

var _ Store = (*googleCloudStore)(nil)

// NewGoogleCloudStore creates a new Google Cloud store that uses Cloud Datastore for storage and Pub/sub for events.
func NewGoogleCloudStore(ctx context.Context, cfg *common.Server, logger *zap.Logger) (Store, error) {
	datastoreClient, err := createDatastore(ctx, cfg.GoogleCloudDatastore)
	if err != nil {
		return nil, err
	}

	pubsubClient, err := createPubSub(ctx, cfg.GoogleCloudPubSub)
	if err != nil {
		return nil, err
	}

	s := &googleCloudStore{
		client:             datastoreClient,
		pubsub:             pubsubClient,
		updates:            eventbus.NewSource[*Updates](),
		agentIndex:         search.NewInMemoryIndex("agent"),
		configurationIndex: search.NewInMemoryIndex("configuration"),
		logger:             logger,

		sessionStore: newBPCookieStore(cfg.SessionsSecret),
	}

	// start listening for events
	go func() {
		err = pubsubClient.subscriber.Receive(ctx, s.receivePubsubMessage)
		if err != nil {
			logger.Fatal("subscriber failed", zap.Error(err))
		}
	}()

	return s, nil
}

func (s *googleCloudStore) Clear() {
	// TODO no easy way to clear everything, maybe GetAll (KeysOnly) and DeleteMulti (by page)
}

func (s *googleCloudStore) Agent(id string) (*model.Agent, error) {
	item, exists, err := getDatastoreResource[*model.Agent](s, model.KindAgent, id)
	if !exists {
		item = nil
	}
	return item, err
}
func (s *googleCloudStore) Agents(ctx context.Context, options ...QueryOption) ([]*model.Agent, error) {
	ctx, span := tracer.Start(ctx, "store/Agents")
	defer span.End()

	opts := makeQueryOptions(options)

	span.SetAttributes(
		attribute.Int("bindplane.list.offset", opts.offset),
		attribute.Int("bindplane.list.limit", opts.limit),
		attribute.String("bindplane.list.sort", opts.sort),
		attribute.String("bindplane.list.selector", opts.selector.String()),
	)
	if opts.query != nil {
		span.SetAttributes(attribute.String("bindplane.list.query", opts.query.Original))
	}

	return getDatastoreResourcesWithQuery[*model.Agent](ctx, s, s.agentIndex, model.KindAgent, &opts)
}

func (s *googleCloudStore) AgentsCount(ctx context.Context, options ...QueryOption) (int, error) {
	return s.client.Count(ctx, datastoreQuery(model.KindAgent, nil))
}

// getAndUpdateAgent gets the agent from the data store and calls the updater on it.
// It appends the passed updates with the appropriate agent and status.
// It does *not* PUT to update the agents in the store, notify subscribers of updates,
// or update the search index.
func (s *googleCloudStore) getAndUpdateAgent(ctx context.Context, agentID string, updater AgentUpdater, updates *Updates) (agent *model.Agent, err error) {
	agentEventType := EventTypeUpdate

	agent, exists, err := getDatastoreResource[*model.Agent](s, model.KindAgent, agentID)
	if err != nil {
		return nil, err
	}

	if !exists {
		agentEventType = EventTypeUpdate
		agent = &model.Agent{ID: agentID}
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
	return agent, nil
}

func (s *googleCloudStore) UpsertAgents(ctx context.Context, agentIDs []string, updater AgentUpdater) ([]*model.Agent, error) {
	ctx, span := tracer.Start(ctx, "store/UpsertAgents")
	defer span.End()

	updates := NewUpdates()
	agents := make([]*model.Agent, 0, len(agentIDs))

	// TODO (dsvanlani) This can be optimized to use GetMulti instead of individual gets.
	for _, id := range agentIDs {
		agent, err := s.getAndUpdateAgent(ctx, id, updater, updates)
		if err != nil {
			return nil, err
		}

		agents = append(agents, agent)
	}

	// make data store resources and update the store
	keys := make([]*datastore.Key, 0, len(agents))
	dsrs := make([]*datastoreResource, 0, len(agents))
	for _, agent := range agents {
		dsr, err := newDatastoreAgent(agent)
		if err != nil {
			return nil, err
		}

		dsrs = append(dsrs, dsr)
		keys = append(keys, dsr.Key)
	}

	_, err := s.client.PutMulti(ctx, keys, dsrs)
	if err != nil {
		return nil, err
	}

	// Update the search index
	for _, agent := range agents {
		err = s.agentIndex.Upsert(agent)
		if err != nil {
			s.logger.Error("failed to update the search index", zap.String("agentID", agent.ID))
		}
	}

	// notify updates
	s.notify(updates)

	return nil, nil
}

// UpsertAgent adds a new Agent to the Store or updates an existing one
func (s *googleCloudStore) UpsertAgent(ctx context.Context, agentID string, updater AgentUpdater) (*model.Agent, error) {
	ctx, span := tracer.Start(ctx, "store/UpsertAgent")
	defer span.End()

	updates := NewUpdates()

	agent, err := s.getAndUpdateAgent(ctx, agentID, updater, updates)
	if err != nil {
		return nil, err
	}

	// store the changes
	err = upsertDatastoreAgent(s, agent)
	if err != nil {
		return nil, err
	}

	// update the index
	err = s.agentIndex.Upsert(agent)
	if err != nil {
		s.logger.Error("failed to update the search index", zap.String("agentID", agent.ID))
	}

	// notify agent updates
	s.notify(updates)

	return agent, nil

}

func (s *googleCloudStore) DeleteAgents(ctx context.Context, agentIDs []string) ([]*model.Agent, error) {
	return deleteDatastoreAgents(ctx, s, agentIDs)
}

func (s *googleCloudStore) Configurations(options ...QueryOption) ([]*model.Configuration, error) {
	opts := makeQueryOptions(options)
	return getDatastoreResourcesWithQuery[*model.Configuration](context.TODO(), s, s.configurationIndex, model.KindConfiguration, &opts)
}

func (s *googleCloudStore) Configuration(name string) (*model.Configuration, error) {
	item, exists, err := getDatastoreResource[*model.Configuration](s, model.KindConfiguration, name)
	if !exists {
		item = nil
	}
	return item, err
}
func (s *googleCloudStore) DeleteConfiguration(name string) (*model.Configuration, error) {
	item, exists, err := deleteDatastoreResourceAndNotify[*model.Configuration](s, model.KindConfiguration, name)
	if !exists {
		return nil, err
	}
	return item, err
}

func (s *googleCloudStore) Source(name string) (*model.Source, error) {
	item, exists, err := getDatastoreResource[*model.Source](s, model.KindSource, name)
	if !exists {
		item = nil
	}
	return item, err
}
func (s *googleCloudStore) Sources() ([]*model.Source, error) {
	return getDatastoreResources[*model.Source](s, model.KindSource, nil)
}
func (s *googleCloudStore) DeleteSource(name string) (*model.Source, error) {
	item, exists, err := deleteDatastoreResourceAndNotify[*model.Source](s, model.KindSource, name)
	if !exists {
		return nil, err
	}
	return item, err
}

func (s *googleCloudStore) SourceType(name string) (*model.SourceType, error) {
	item, exists, err := getDatastoreResource[*model.SourceType](s, model.KindSourceType, name)
	if !exists {
		item = nil
	}
	return item, err
}
func (s *googleCloudStore) SourceTypes() ([]*model.SourceType, error) {
	return getDatastoreResources[*model.SourceType](s, model.KindSourceType, nil)
}
func (s *googleCloudStore) DeleteSourceType(name string) (*model.SourceType, error) {
	item, exists, err := deleteDatastoreResourceAndNotify[*model.SourceType](s, model.KindSourceType, name)
	if !exists {
		return nil, err
	}
	return item, err
}

func (s *googleCloudStore) Destination(name string) (*model.Destination, error) {
	item, exists, err := getDatastoreResource[*model.Destination](s, model.KindDestination, name)
	if !exists {
		item = nil
	}
	return item, err
}
func (s *googleCloudStore) Destinations() ([]*model.Destination, error) {
	return getDatastoreResources[*model.Destination](s, model.KindDestination, nil)
}
func (s *googleCloudStore) DeleteDestination(name string) (*model.Destination, error) {
	item, exists, err := deleteDatastoreResourceAndNotify[*model.Destination](s, model.KindDestination, name)
	if !exists {
		return nil, err
	}
	return item, err
}

func (s *googleCloudStore) DestinationType(name string) (*model.DestinationType, error) {
	item, exists, err := getDatastoreResource[*model.DestinationType](s, model.KindDestinationType, name)
	if !exists {
		item = nil
	}
	return item, err
}
func (s *googleCloudStore) DestinationTypes() ([]*model.DestinationType, error) {
	return getDatastoreResources[*model.DestinationType](s, model.KindDestinationType, nil)
}
func (s *googleCloudStore) DeleteDestinationType(name string) (*model.DestinationType, error) {
	item, exists, err := deleteDatastoreResourceAndNotify[*model.DestinationType](s, model.KindDestinationType, name)
	if !exists {
		return nil, err
	}
	return item, err
}

// ----------------------------------------------------------------------

func (s *googleCloudStore) ApplyResources(resources []model.Resource) ([]model.ResourceStatus, error) {
	updates := NewUpdates()

	// resourceStatuses to return for the applied resources
	resourceStatuses := make([]model.ResourceStatus, 0, len(resources))

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

		status, err := upsertAnyDatastoreResource(s, resource)
		if err != nil {
			resourceStatuses = append(resourceStatuses, *model.NewResourceStatusWithReason(resource, model.StatusError, err.Error()))
			errs = multierror.Append(errs, err)
			continue
		}
		resourceStatuses = append(resourceStatuses, *model.NewResourceStatus(resource, status))

		switch status {
		case model.StatusCreated:
			updates.IncludeResource(resource, EventTypeInsert)
		case model.StatusConfigured:
			updates.IncludeResource(resource, EventTypeUpdate)
		}

	}
	s.notify(updates)

	s.logger.Info("ApplyResources complete", zap.Any("resourceStatuses", resourceStatuses))

	return resourceStatuses, errs
}

// Batch delete of a slice of resources, returns the successfully deleted resources or an error.
func (s *googleCloudStore) DeleteResources(resources []model.Resource) ([]model.ResourceStatus, error) {
	updates := NewUpdates()

	// track deleteStatuses to return
	deleteStatuses := make([]model.ResourceStatus, 0)

	for _, r := range resources {
		deleted, exists, err := deleteAnyDatastoreResource(s, r)

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

// ----------------------------------------------------------------------

// AgentConfiguration returns the configuration that should be applied to an agent.
func (s *googleCloudStore) AgentConfiguration(agentID string) (*model.Configuration, error) {
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

	// iterate over all configurations and look for matches
	configurations, err := s.Configurations()
	if err != nil {
		return nil, err
	}
	for _, configuration := range configurations {
		if configuration.IsForAgent(agent) {
			return configuration, nil
		}
	}

	return nil, nil
}

// AgentsIDsMatchingConfiguration returns the list of agent IDs that are using the specified configuration
func (s *googleCloudStore) AgentsIDsMatchingConfiguration(configuration *model.Configuration) ([]string, error) {
	ids := s.AgentIndex().Select(configuration.Spec.Selector.MatchLabels)
	return ids, nil
}

// CleanupDisconnectedAgents removes agents that have disconnected before the specified time
func (s *googleCloudStore) CleanupDisconnectedAgents(since time.Time) error {
	// TODO: find agents where status=disconnected and disconnectedAt < since
	return nil
}

// Updates will receive pipelines and configurations that have been updated or deleted, either because the
// configuration changed or a component in them was updated. Agents with labels that change are also sent with
// Updates.
func (s *googleCloudStore) Updates() eventbus.Source[*Updates] {
	return s.updates
}

// Index provides access to the search Index implementation managed by the Store
func (s *googleCloudStore) AgentIndex() search.Index {
	return s.agentIndex
}

// ConfigurationIndex provides access to the search Index for Configurations
func (s *googleCloudStore) ConfigurationIndex() search.Index {
	return s.configurationIndex
}

// TODO (auth) we need to implement this interface in google cloudstore to allow a
// multi-node running of BindPlane
func (s *googleCloudStore) UserSessions() sessions.Store {
	return s.sessionStore
}

// ----------------------------------------------------------------------
// events

func (s *googleCloudStore) notify(updates *Updates) {
	ctx, span := tracer.Start(context.TODO(), "store/notify")
	defer span.End()

	err := updates.addTransitiveUpdates(s)
	if err != nil {
		// TODO: if we can't notify about all updates, what do we do?
		s.logger.Error("unable to add transitive updates", zap.Any("updates", updates), zap.Error(err))
	}
	if !updates.Empty() {
		// send to pub/sub. eventually the messages will return to this node as events.
		s.sendPubsubMessage(ctx, updates)
	}
}

func (s *googleCloudStore) sendPubsubMessage(ctx context.Context, updates *Updates) {
	bytes, err := json.Marshal(updates)
	if err != nil {
		s.logger.Warn(fmt.Sprintf("failed to marshal Updates message: %s", err))
		return
	}
	s.pubsub.publisher.Publish(ctx, &pubsub.Message{Data: bytes})
}

// receivePubsubMessage receives messages from this node or other nodes and forwards them to subscribers in this node
func (s *googleCloudStore) receivePubsubMessage(ctx context.Context, msg *pubsub.Message) {
	ctx, span := tracer.Start(ctx, "store/receivePubsubMessage")
	defer span.End()

	defer msg.Ack()

	var updates Updates
	if err := json.Unmarshal(msg.Data, &updates); err != nil {
		s.logger.Warn(fmt.Sprintf("failed to unmarshal Updates message: %s", err))
		return
	}

	span.SetAttributes(
		attribute.Int("configurations", len(updates.Configurations)),
		attribute.Int("agents", len(updates.Agents)),
	)

	// update indexes which must be in sync across servers
	for _, event := range updates.Configurations {
		updateIndex(s, s.configurationIndex, event)
	}
	for _, event := range updates.Agents {
		updateIndex(s, s.agentIndex, event)
	}

	s.updates.Send(&updates)
}

type indexed interface {
	model.HasUniqueKey
	search.Indexed
}

func updateIndex[T indexed](s *googleCloudStore, index search.Index, event Event[T]) {
	var err error
	switch event.Type {
	case EventTypeRemove:
		err = index.Remove(event.Item)
	default:
		err = index.Upsert(event.Item)
	}
	if err != nil {
		s.logger.Error("failed to update the search index", zap.String("ID", event.Item.IndexID()), zap.Error(err))
	}
}

// ----------------------------------------------------------------------
// datastore interaction

func datastoreKey(kind model.Kind, uniqueKey string) *datastore.Key {
	return datastore.NameKey(string(kind), uniqueKey, nil)
}

func datastoreQuery(kind model.Kind, opts *queryOptions) *datastore.Query {
	query := datastore.NewQuery(string(kind))
	if opts != nil {
		if opts.offset > 0 {
			query = query.Offset(opts.offset)
		}
		if opts.limit > 0 {
			query = query.Limit(opts.limit)
		}
		if opts.sort != "" && kind == model.KindAgent {
			// sort currently only supported for agents for id, name, or status and id is key order
			switch opts.sort {
			case "name":
				query = query.Order("name")
			case "status":
				query = query.Order("status")
			}
		}

		if !opts.selector.Empty() {
			// translate requirements to a query
			labels, _ := opts.selector.MatchLabels()
			for k, v := range labels {
				query = query.Filter("labels=", fmt.Sprintf("%s=%s", k, v))
			}
		}
	}

	return query
}

// datastoreResource is the value stored in the datastore. It is common to all datastore types.
type datastoreResource struct {
	Key    *datastore.Key `datastore:"__key__"`
	Name   string         `datastore:"name"`
	Status int8           `datastore:"status,omitempty"`
	Labels []string       `datastore:"labels"` // stored as name=value
	Body   []byte         `datastore:"body,noindex"`
}

func datastoreLabels(labels model.Labels) []string {
	set := labels.Set
	result := make([]string, 0, len(set))
	for k, v := range set {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

func newDatastoreResource(resource model.Resource) (*datastoreResource, error) {
	// marshal the body to json
	data, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}
	return &datastoreResource{
		Key:    datastoreKey(resource.GetKind(), resource.UniqueKey()),
		Name:   resource.Name(),
		Labels: datastoreLabels(resource.GetLabels()),
		Body:   data,
	}, nil
}

func newDatastoreAgent(agent *model.Agent) (*datastoreResource, error) {
	// marshal the body to json
	data, err := json.Marshal(agent)
	if err != nil {
		return nil, err
	}
	return &datastoreResource{
		Key:    datastoreKey(model.KindAgent, agent.UniqueKey()),
		Name:   agent.Name,
		Status: int8(agent.Status),
		Labels: datastoreLabels(agent.GetLabels()),
		Body:   data,
	}, nil
}

func decodeDatastoreResource[T any](dr *datastoreResource, resource *T) error {
	return json.Unmarshal(dr.Body, resource)
}

// ----------------------------------------------------------------------

func upsertAnyDatastoreResource(s *googleCloudStore, r model.Resource) (model.UpdateStatus, error) {
	// TODO if resource type and kind get out of sync, this will cause issues
	switch r.GetKind() {
	case model.KindConfiguration:
		return upsertDatastoreResource(s, r.(*model.Configuration))
	case model.KindSource:
		return upsertDatastoreResource(s, r.(*model.Source))
	case model.KindSourceType:
		return upsertDatastoreResource(s, r.(*model.SourceType))
	case model.KindDestination:
		return upsertDatastoreResource(s, r.(*model.Destination))
	case model.KindDestinationType:
		return upsertDatastoreResource(s, r.(*model.DestinationType))
	default:
		return model.StatusError, fmt.Errorf("unable to use ApplyResource with %s", string(r.GetKind()))
	}
}

func upsertDatastoreAgent(s *googleCloudStore, agent *model.Agent) error {
	dsr, err := newDatastoreAgent(agent)
	if err != nil {
		return err
	}

	_, err = s.client.Put(context.TODO(), dsr.Key, dsr)
	if err != nil {
		return err
	}

	return nil
}

func upsertDatastoreResource[R model.Resource](s *googleCloudStore, r R) (model.UpdateStatus, error) {
	successStatus := model.StatusCreated
	existing, exists, err := getDatastoreResource[R](s, r.GetKind(), r.Name())
	if err != nil {
		return model.StatusError, err
	} else if exists {
		successStatus = model.StatusConfigured
		// preserve the id (if possible)
		r.SetID(existing.ID())
	}

	dsr, err := newDatastoreResource(r)
	if err != nil {
		return model.StatusUnchanged, fmt.Errorf("failed to marshal the resource: %w", err)
	}

	_, err = s.client.Put(context.TODO(), dsr.Key, dsr)
	if err != nil {
		return model.StatusUnchanged, fmt.Errorf("failed to put the resource: %w", err)
	}

	// note that we don't know if the item existed before
	return successStatus, err
}

func getDatastoreResource[R any](s *googleCloudStore, kind model.Kind, name string) (resource R, exists bool, err error) {
	var dsr datastoreResource

	if err = s.client.Get(context.TODO(), datastoreKey(kind, name), &dsr); err != nil {
		if errors.Is(err, datastore.ErrNoSuchEntity) {
			return resource, false, nil
		}
		s.logger.Error("error in client.Get", zap.Error(err))
		return resource, true, fmt.Errorf("failed to get the resource: %w", err)
	}

	if err = decodeDatastoreResource(&dsr, &resource); err != nil {
		return resource, true, fmt.Errorf("failed to unmarshal the resource: %w", err)
	}

	return resource, true, nil
}

func deleteDatastoreResourceAndNotify[R model.Resource](s *googleCloudStore, kind model.Kind, name string) (resource R, exists bool, err error) {
	deleted, exists, err := deleteDatastoreResource[R](s, kind, name)

	if err == nil && exists {
		updates := NewUpdates()
		updates.IncludeResource(deleted, EventTypeRemove)
		s.notify(updates)
	}

	return deleted, exists, err
}

func deleteAnyDatastoreResource(s *googleCloudStore, r model.Resource) (model.Resource, bool, error) {
	// TODO if resource type and kind get out of sync, this will cause issues
	switch r.GetKind() {
	case model.KindConfiguration:
		return deleteDatastoreResource[*model.Configuration](s, r.GetKind(), r.Name())
	case model.KindSource:
		return deleteDatastoreResource[*model.Source](s, r.GetKind(), r.Name())
	case model.KindSourceType:
		return deleteDatastoreResource[*model.SourceType](s, r.GetKind(), r.Name())
	case model.KindDestination:
		return deleteDatastoreResource[*model.Destination](s, r.GetKind(), r.Name())
	case model.KindDestinationType:
		return deleteDatastoreResource[*model.DestinationType](s, r.GetKind(), r.Name())
	default:
		return nil, false, fmt.Errorf("unable to use DeleteResources with %s", string(r.GetKind()))
	}
}

func deleteDatastoreResource[R model.Resource](s *googleCloudStore, kind model.Kind, name string) (resource R, exists bool, err error) {
	var dsr datastoreResource
	if err = s.client.Get(context.TODO(), datastoreKey(kind, name), &dsr); err != nil {
		if errors.Is(err, datastore.ErrNoSuchEntity) {
			return resource, false, nil
		}
		return resource, true, fmt.Errorf("failed to check if resource exists: %w", err)
	}
	if err = decodeDatastoreResource(&dsr, &resource); err != nil {
		return resource, true, fmt.Errorf("failed to unmarshal existing resource: %w", err)
	}

	// Check if the resources is referenced by another
	dependencies, err := FindDependentResources(context.TODO(), s, resource)
	if !dependencies.empty() {
		return resource, true, ErrResourceInUse
	}

	if err = s.client.Delete(context.TODO(), datastoreKey(kind, name)); err != nil {
		return resource, true, fmt.Errorf("failed to delete the resource: %w", err)
	}
	return resource, true, nil
}

func getDatastoreResources[R any](s *googleCloudStore, kind model.Kind, opts *queryOptions) ([]R, error) {
	query := datastoreQuery(kind, opts)
	var list []datastoreResource
	if _, err := s.client.GetAll(context.TODO(), query, &list); err != nil {
		return nil, err
	}

	results := make([]R, 0, len(list))
	for _, dsr := range list {
		dsr := dsr // copy to local variable to securely pass a reference to a loop variable
		var result R
		if err := decodeDatastoreResource(&dsr, &result); err != nil {
			s.logger.Error("unable to decode datastore resource", zap.String("name", dsr.Name), zap.String("kind", string(kind)))
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

func getDatastoreResourcesWithQuery[R any](ctx context.Context, s *googleCloudStore, index search.Index, kind model.Kind, opts *queryOptions) ([]R, error) {
	// search is implemented using the search index
	if opts.query != nil {
		ids, err := index.Search(ctx, opts.query)
		if err != nil {
			return nil, err
		}
		return getDatastoreResourcesByID[R](ctx, s, kind, ids)
	}

	return getDatastoreResources[R](s, kind, opts)
}

func getDatastoreResourcesByID[R any](ctx context.Context, s *googleCloudStore, kind model.Kind, ids []string) ([]R, error) {
	ctx, span := tracer.Start(ctx, "store/getDatastoreResourcesByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("kind", string(kind)),
		attribute.Int("count", len(ids)),
	)

	keys := make([]*datastore.Key, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, datastoreKey(kind, id))
	}

	list := make([]datastoreResource, len(keys))
	if err := s.client.GetMulti(ctx, keys, list); err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}

	span.SetAttributes(
		attribute.Int("found", len(list)),
	)

	results := make([]R, 0, len(list))
	for _, dsr := range list {
		dsr := dsr // copy to local variable to securely pass a reference to a loop variable
		var result R
		err := decodeDatastoreResource(&dsr, &result)
		if err != nil {
			s.logger.Error("error decoding resource", zap.String("kind", string(kind)), zap.Error(err))
		}
		results = append(results, result)
	}

	return results, nil
}

func deleteDatastoreAgents(ctx context.Context, s *googleCloudStore, ids []string) ([]*model.Agent, error) {
	ctx, span := tracer.Start(ctx, "store/deleteDatastoreAgents")
	defer span.End()

	agents, err := getDatastoreResourcesByID[*model.Agent](ctx, s, model.KindAgent, ids)
	if err != nil {
		return nil, err
	}

	keys := make([]*datastore.Key, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, datastoreKey(model.KindAgent, id))
	}

	if err := s.client.DeleteMulti(ctx, keys); err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}

	// set deleted status on the agents that have been deleted
	for _, agent := range agents {
		agent.Status = model.Deleted
	}

	return agents, nil
}

// ----------------------------------------------------------------------
// google cloud client creation

func clientOptions(endpoint string, credentialsFile string, options ...option.ClientOption) []option.ClientOption {
	results := []option.ClientOption{}
	results = append(results, options...)

	if endpoint != "" {
		results = append(results,
			option.WithEndpoint(endpoint),
			option.WithoutAuthentication(),
			option.WithGRPCDialOption(grpc.WithInsecure()),
		)
	}

	if credentialsFile != "" {
		results = append(results, option.WithCredentialsFile(credentialsFile))
	}

	return results
}

func createDatastore(ctx context.Context, config *common.GoogleCloudDatastore) (*datastore.Client, error) {
	return datastore.NewClient(ctx, config.ProjectID, clientOptions(config.Endpoint, config.CredentialsFile)...)
}

type pubsubClient struct {
	// TODO: Close()
	client     *pubsub.Client
	publisher  *pubsub.Topic
	subscriber *pubsub.Subscription
}

func createPubSub(ctx context.Context, config *common.GoogleCloudPubSub) (*pubsubClient, error) {
	client, err := pubsub.NewClient(ctx, config.ProjectID, clientOptions(config.Endpoint, config.CredentialsFile)...)
	if err != nil {
		return nil, fmt.Errorf("create pubsub: %w", err)
	}
	topic := client.Topic(config.Topic)
	topicExists, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("check topic: %w", err)
	}
	if !topicExists {
		return nil, errors.New("topic " + config.Topic + " does not exist")
	}
	subID := config.Subscription
	if subID == "" {
		subID, err = os.Hostname()
		if err != nil {
			subID = uuid.NewString()
		}
		subID = config.Topic + "-" + subID
	}
	sub := client.Subscription(subID)
	subExists, err := sub.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("check subscription: %w", err)
	}
	if !subExists {
		sub, err = client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
			Topic:            topic,
			ExpirationPolicy: 7 * 24 * time.Hour,
		})
		if err != nil {
			return nil, fmt.Errorf("create subscription: %w", err)
		}
	}

	return &pubsubClient{
		client:     client,
		publisher:  topic,
		subscriber: sub,
	}, nil
}
