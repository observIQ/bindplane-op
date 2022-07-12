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
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/sessions"
	"github.com/hashicorp/go-multierror"
	"github.com/observiq/bindplane-op/internal/eventbus"
	"github.com/observiq/bindplane-op/internal/store/search"
	"github.com/observiq/bindplane-op/model"
	embedded "github.com/observiq/bindplane-op/resources"
	"go.uber.org/zap"
)

// Store handles interacting with a storage backend,
type Store interface {
	Clear()

	Agent(string) (*model.Agent, error)
	Agents(ctx context.Context, options ...QueryOption) ([]*model.Agent, error)
	AgentsCount(context.Context, ...QueryOption) (int, error)
	// UpsertAgent adds a new Agent to the Store or updates an existing one
	UpsertAgent(ctx context.Context, agentID string, updater AgentUpdater) (*model.Agent, error)
	UpsertAgents(ctx context.Context, agentIDs []string, updater AgentUpdater) ([]*model.Agent, error)
	DeleteAgents(ctx context.Context, agentIDs []string) ([]*model.Agent, error)

	Configurations(options ...QueryOption) ([]*model.Configuration, error)
	Configuration(string) (*model.Configuration, error)
	DeleteConfiguration(string) (*model.Configuration, error)

	Source(name string) (*model.Source, error)
	Sources() ([]*model.Source, error)
	DeleteSource(name string) (*model.Source, error)

	SourceType(name string) (*model.SourceType, error)
	SourceTypes() ([]*model.SourceType, error)
	DeleteSourceType(name string) (*model.SourceType, error)

	Destination(name string) (*model.Destination, error)
	Destinations() ([]*model.Destination, error)
	DeleteDestination(name string) (*model.Destination, error)

	DestinationType(name string) (*model.DestinationType, error)
	DestinationTypes() ([]*model.DestinationType, error)
	DeleteDestinationType(name string) (*model.DestinationType, error)

	ApplyResources([]model.Resource) ([]model.ResourceStatus, error)
	// Batch delete of a slice of resources, returns the successfully deleted resources or an error.
	DeleteResources([]model.Resource) ([]model.ResourceStatus, error)

	// AgentConfiguration returns the configuration that should be applied to an agent.
	AgentConfiguration(agentID string) (*model.Configuration, error)

	// AgentsIDsMatchingConfiguration returns the list of agent IDs that are using the specified configuration
	AgentsIDsMatchingConfiguration(*model.Configuration) ([]string, error)

	// CleanupDisconnectedAgents removes agents that have disconnected before the specified time
	CleanupDisconnectedAgents(since time.Time) error

	// Updates will receive pipelines and configurations that have been updated or deleted, either because the
	// configuration changed or a component in them was updated. Agents inserted/updated from UpsertAgent and agents
	// removed from CleanupDisconnectedAgents are also sent with Updates.
	Updates() eventbus.Source[*Updates]

	// AgentIndex provides access to the search AgentIndex implementation managed by the Store
	AgentIndex() search.Index

	// ConfigurationIndex provides access to the search Index for Configurations
	ConfigurationIndex() search.Index

	// UserSessions must implement the gorilla sessions.Store interface
	UserSessions() sessions.Store
}

// AgentUpdater is given the current Agent model (possibly empty except for ID) and should update the Agent directly. We
// take this approach so that appropriate locking and/or transactions can be used for the operation as needed by the
// Store implementation.
type AgentUpdater func(current *model.Agent)

// ErrResourceMissing is used in delete functions to indicate the delete
// could not be performed because no such resource exists
var ErrResourceMissing = errors.New("resource not found")

// ErrResourceInUse is used in delete functions to indicate the delete
// could not be performed because the Resource is a dependency of another.
// i.e. the Source that is being deleted is being referenced in a Configuration.
var ErrResourceInUse = errors.New("resource in use")

// ----------------------------------------------------------------------

// queryOptions represents the set of options available for a store query
type queryOptions struct {
	selector model.Selector
	query    *search.Query
	offset   int
	limit    int
	sort     string
}

func makeQueryOptions(options []QueryOption) queryOptions {
	opts := queryOptions{
		selector: model.EverythingSelector(),
	}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

// QueryOption is an option used in Store queries
type QueryOption func(*queryOptions)

// WithSelector adds a selector to the query options
func WithSelector(selector model.Selector) QueryOption {
	return func(opts *queryOptions) {
		opts.selector = selector
	}
}

// WithQuery adds a search query string to the query options
func WithQuery(query *search.Query) QueryOption {
	return func(opts *queryOptions) {
		opts.query = query
	}
}

// WithOffset sets the offset for the results to return. For paging, if the pages have 10 items per page and this is the
// 3rd page, set the offset to 20.
func WithOffset(offset int) QueryOption {
	return func(opts *queryOptions) {
		opts.offset = offset
	}
}

// WithLimit sets the maximum number of results to return. For paging, if the pages have 10 items per page, set the
// limit to 10.
func WithLimit(limit int) QueryOption {
	return func(opts *queryOptions) {
		opts.limit = limit
	}
}

// WithSort sets the sort order for the request. The sort value is the name of the field, sorted ascending. To sort
// descending, prefix the field with a minus sign (-). Some Stores only allow sorting by certain fields. Sort values not
// supported will be ignored.
func WithSort(field string) QueryOption {
	return func(opts *queryOptions) {
		opts.sort = field
	}
}

// ----------------------------------------------------------------------
// seeding resources

// Seed adds bundled resources to the store
func Seed(store Store, logger *zap.Logger) error {
	var errs error
	for _, dir := range []string{"source-types", "destination-types"} {
		err := seedDir(dir, store, logger)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

// seedDir adds bundled resources from the specified dir to the store
func seedDir(dir string, store Store, logger *zap.Logger) error {
	filesystem := embedded.Files
	resourceTypes := make([]model.Resource, 0)

	_ = fs.WalkDir(filesystem, dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		file, err := filesystem.Open(path)
		if err != nil {
			logger.Error("error opening file", zap.String("path", path), zap.Error(err))
			return nil
		}

		r, err := model.ResourcesFromReader(file)
		if err != nil {
			logger.Error("failed to get resource from reader", zap.String("path", path), zap.Error(err))
			return nil
		}

		parsed, err := model.ParseResources(r)
		if err != nil {
			logger.Error("error parsing resources", zap.Error(err))
			return nil
		}
		resourceTypes = append(resourceTypes, parsed...)

		return nil
	})

	updates, err := store.ApplyResources(resourceTypes)
	if err != nil {
		return err
	}

	messages := make([]string, len(updates))
	for i, update := range updates {
		messages[i] = fmt.Sprintf("%s %s", update.Resource.Name(), update.Status)
	}

	logger.Info("Seeded ResourceTypes", zap.String("dir", dir), zap.Any("resourceTypes", messages))

	return nil
}

type dependency struct {
	name string
	kind model.Kind
}

// DependentResources is the return type of store.dependentResources
// and used to construct DependencyError.
// It has help methods empty(), message(), and add().
type DependentResources []dependency

func (r *DependentResources) empty() bool {
	return len(*r) == 0
}

func (r *DependentResources) message() string {
	msg := "Dependent resources:\n"
	for _, item := range *r {
		msg += fmt.Sprintf("%s %s\n", item.kind, item.name)
	}
	return msg
}

func (r *DependentResources) add(d dependency) {
	*r = append(*r, d)
}

// DependencyError is returned when trying to delete a resource
// that is being referenced by other resources.
type DependencyError struct {
	dependencies DependentResources
}

func (de *DependencyError) Error() string {
	return de.dependencies.message()
}

func newDependencyError(d DependentResources) error {
	return &DependencyError{
		dependencies: d,
	}
}

// ----------------------------------------------------------------------

// FindDependentResources finds the dependent resources using the ConfigurationIndex provided by the Store.
func FindDependentResources(ctx context.Context, s Store, r model.Resource) (DependentResources, error) {
	var dependencies DependentResources

	switch r.GetKind() {
	case model.KindSource:
		ids, err := search.Field(ctx, s.ConfigurationIndex(), "source", r.Name())
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			dependencies.add(dependency{name: id, kind: model.KindConfiguration})
		}

	case model.KindDestination:
		ids, err := search.Field(ctx, s.ConfigurationIndex(), "destination", r.Name())
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			dependencies.add(dependency{name: id, kind: model.KindConfiguration})
		}
	}

	return dependencies, nil
}

// ----------------------------------------------------------------------
// generic helpers for sorting and paging

func applySortOffsetAndLimit[T any](list []T, opts queryOptions, fieldAccessor fieldAccessor[T]) []T {
	if opts.sort != "" {
		sortField := opts.sort
		ascending := true
		if opts.sort[0] == '-' {
			sortField = opts.sort[1:]
			ascending = false
		}
		sort.Sort(byField[T]{
			list:          list,
			field:         sortField,
			ascending:     ascending,
			fieldAccessor: fieldAccessor,
		})
	}
	if opts.offset != 0 {
		offset := opts.offset
		if offset > len(list) {
			offset = len(list)
		}
		list = list[offset:]
	}
	if opts.limit != 0 {
		limit := opts.limit
		if limit > len(list) {
			limit = len(list)
		}
		list = list[:limit]
	}
	return list
}

func newBPCookieStore(secret string) *sessions.CookieStore {
	store := sessions.NewCookieStore([]byte(secret))
	store.Options.MaxAge = 60 * 60 // 1 hour
	store.Options.SameSite = http.SameSiteStrictMode
	return store
}

type fieldAccessor[T any] func(field string, item T) string

type byField[T any] struct {
	list          []T
	field         string
	ascending     bool
	fieldAccessor fieldAccessor[T]
}

var _ sort.Interface = (*byField[any])(nil)

// Len returns the length of the list
func (r byField[T]) Len() int {
	return len(r.list)
}

// Swap swaps to items in the list
func (r byField[T]) Swap(i, j int) {
	r.list[i], r.list[j] = r.list[j], r.list[i]
}

// Less returns true if the i'th item is less than the j'th
func (r byField[T]) Less(i, j int) bool {
	f1 := r.fieldAccessor(r.field, r.list[i])
	f2 := r.fieldAccessor(r.field, r.list[j])
	if r.ascending {
		return f1 < f2
	}
	return f1 > f2
}
