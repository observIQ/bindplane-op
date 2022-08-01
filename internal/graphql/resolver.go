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

package graphql

import (
	"context"

	"github.com/observiq/bindplane-op/internal/eventbus"
	model1 "github.com/observiq/bindplane-op/internal/graphql/model"
	"github.com/observiq/bindplane-op/internal/server"
	"github.com/observiq/bindplane-op/internal/store"
	"github.com/observiq/bindplane-op/internal/store/search"
	"github.com/observiq/bindplane-op/model"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("graphql")

//go:generate gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver TODO(doc)
type Resolver struct {
	bindplane server.BindPlane
	updates   eventbus.Source[*store.Updates]
}

// NewResolver returns a new Resolver and starts a go routine
// that sends agent updates to observers.
func NewResolver(bindplane server.BindPlane) *Resolver {
	resolver := &Resolver{
		bindplane: bindplane,
		updates:   eventbus.NewSource[*store.Updates](),
	}

	// relay events from the store to the resolver where they will be dispatched to individual graphql subscriptions
	eventbus.Relay(context.TODO(), bindplane.Store().Updates(), resolver.updates)

	return resolver
}

func applySelectorToChanges(selector *model.Selector, changes store.Events[*model.Agent]) store.Events[*model.Agent] {
	if selector == nil {
		return changes
	}
	result := store.NewEvents[*model.Agent]()
	for _, change := range changes {
		if change.Type != store.EventTypeRemove && !selector.Matches(change.Item.Labels) {
			result.Include(change.Item, store.EventTypeRemove)
		} else {
			result.Include(change.Item, change.Type)
		}
	}
	return result
}

func applyQueryToChanges(query *search.Query, index search.Index, changes store.Events[*model.Agent]) store.Events[*model.Agent] {
	if query == nil {
		return changes
	}
	result := store.NewEvents[*model.Agent]()
	for _, change := range changes {
		if change.Type != store.EventTypeRemove && !index.Matches(query, change.Item.ID) {
			result.Include(change.Item, store.EventTypeRemove)
		} else {
			result.Include(change.Item, change.Type)
		}
	}
	return result
}

func applySelectorToEvents[T model.Resource](selector *model.Selector, events store.Events[T]) store.Events[T] {
	if selector == nil {
		return events
	}
	result := store.NewEvents[T]()
	for _, event := range events {
		if event.Type != store.EventTypeRemove && !selector.Matches(event.Item.GetLabels()) {
			result.Include(event.Item, store.EventTypeRemove)
		} else {
			result.Include(event.Item, event.Type)
		}
	}
	return result
}

func applyQueryToEvents[T model.Resource](query *search.Query, index search.Index, events store.Events[T]) store.Events[T] {
	if query == nil || index == nil {
		return events
	}
	result := store.NewEvents[T]()
	for _, event := range events {
		if event.Type != store.EventTypeRemove && !index.Matches(query, event.Item.Name()) {
			result.Include(event.Item, store.EventTypeRemove)
		} else {
			result.Include(event.Item, event.Type)
		}
	}
	return result
}

func (r *Resolver) parseSelectorAndQuery(selector *string, query *string) (*model.Selector, *search.Query, error) {
	var parsedSelector *model.Selector
	if selector != nil {
		sel, err := model.SelectorFromString(*selector)
		if err != nil {
			return nil, nil, err
		}
		parsedSelector = &sel
	}

	// parse the parsedQuery, if any
	var parsedQuery *search.Query
	if query != nil && *query != "" {
		q := search.ParseQuery(*query)
		q.ReplaceVersionLatest(r.bindplane.Versions())
		parsedQuery = q
	}

	return parsedSelector, parsedQuery, nil
}

func (r *Resolver) queryOptionsAndSuggestions(selector *string, query *string, index search.Index) ([]store.QueryOption, []*search.Suggestion, error) {
	parsedSelector, parsedQuery, err := r.parseSelectorAndQuery(selector, query)
	if err != nil {
		return nil, nil, err
	}
	return r.queryOptionsAndSuggestionsParsed(parsedSelector, parsedQuery, index)
}

func (r *Resolver) queryOptionsAndSuggestionsParsed(parsedSelector *model.Selector, parsedQuery *search.Query, index search.Index) ([]store.QueryOption, []*search.Suggestion, error) {
	options := []store.QueryOption{}
	if parsedSelector != nil {
		options = append(options, store.WithSelector(*parsedSelector))
	}

	var suggestions []*search.Suggestion
	if parsedQuery != nil {
		options = append(options, store.WithQuery(parsedQuery))

		s, err := index.Suggestions(parsedQuery)
		if err != nil {
			return nil, nil, err
		}

		suggestions = s
	}
	return options, suggestions, nil
}

func (r *Resolver) seedAgents(ctx context.Context, parsedSelector *model.Selector, query *string, parsedQuery *search.Query, channel chan *model1.AgentChanges) {
	// seed with insert AgentChange events to get started, similar to an Agents query.
	options, suggestions, err := r.queryOptionsAndSuggestionsParsed(parsedSelector, parsedQuery, r.bindplane.Store().AgentIndex())
	if err != nil {
		return
	}

	agents, _ := r.bindplane.Store().Agents(ctx, options...)
	changes := make([]*model1.AgentChange, 0, len(agents))
	for _, agent := range agents {
		changes = append(changes, &model1.AgentChange{
			Agent:      agent,
			ChangeType: model1.AgentChangeTypeInsert,
		})
	}
	channel <- &model1.AgentChanges{
		AgentChanges: changes,
		Query:        query,
		Suggestions:  suggestions,
	}
}
