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

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/observiq/bindplane-op/internal/eventbus"
	"github.com/observiq/bindplane-op/internal/graphql/generated"
	model1 "github.com/observiq/bindplane-op/internal/graphql/model"
	"github.com/observiq/bindplane-op/internal/store"
	"github.com/observiq/bindplane-op/model"
	"go.uber.org/zap"
)

// Labels is the resolver for the labels field.
func (r *agentResolver) Labels(ctx context.Context, obj *model.Agent) (map[string]interface{}, error) {
	labels := map[string]interface{}{}
	for k := range obj.Labels.Set {
		labels[k] = obj.Labels.Get(k)
	}
	return labels, nil
}

// Status is the resolver for the status field.
func (r *agentResolver) Status(ctx context.Context, obj *model.Agent) (int, error) {
	return int(obj.Status), nil
}

// Configuration is the resolver for the configuration field.
func (r *agentResolver) Configuration(ctx context.Context, obj *model.Agent) (*model1.AgentConfiguration, error) {
	ac := &model1.AgentConfiguration{}
	if err := mapstructure.Decode(obj.Configuration, ac); err != nil {
		return &model1.AgentConfiguration{}, err
	}

	return ac, nil
}

// ConfigurationResource is the resolver for the configurationResource field.
func (r *agentResolver) ConfigurationResource(ctx context.Context, obj *model.Agent) (*model.Configuration, error) {
	return r.bindplane.Store().AgentConfiguration(obj.ID)
}

// MatchLabels is the resolver for the matchLabels field.
func (r *agentSelectorResolver) MatchLabels(ctx context.Context, obj *model.AgentSelector) (map[string]interface{}, error) {
	labels := map[string]interface{}{}
	for k := range obj.MatchLabels {
		labels[k] = obj.MatchLabels[k]
	}
	return labels, nil
}

// Kind is the resolver for the kind field.
func (r *configurationResolver) Kind(ctx context.Context, obj *model.Configuration) (string, error) {
	return string(obj.GetKind()), nil
}

// Kind is the resolver for the kind field.
func (r *destinationResolver) Kind(ctx context.Context, obj *model.Destination) (string, error) {
	return string(obj.GetKind()), nil
}

// Kind is the resolver for the kind field.
func (r *destinationTypeResolver) Kind(ctx context.Context, obj *model.DestinationType) (string, error) {
	return string(obj.GetKind()), nil
}

// Labels is the resolver for the labels field.
func (r *metadataResolver) Labels(ctx context.Context, obj *model.Metadata) (map[string]interface{}, error) {
	labels := map[string]interface{}{}
	for k := range obj.Labels.Set {
		labels[k] = obj.Labels.Get(k)
	}
	return labels, nil
}

// Type is the resolver for the type field.
func (r *parameterDefinitionResolver) Type(ctx context.Context, obj *model.ParameterDefinition) (model1.ParameterType, error) {
	switch obj.Type {
	case "strings":
		return model1.ParameterTypeStrings, nil
	case "string":
		return model1.ParameterTypeString, nil

	case "enum":
		return model1.ParameterTypeEnum, nil

	case "bool":
		return model1.ParameterTypeBool, nil

	case "int":
		return model1.ParameterTypeInt, nil

	default:
		return "", errors.New("unknown parameter type")
	}
}

// Kind is the resolver for the kind field.
func (r *processorResolver) Kind(ctx context.Context, obj *model.Processor) (string, error) {
	return string(obj.GetKind()), nil
}

// Kind is the resolver for the kind field.
func (r *processorTypeResolver) Kind(ctx context.Context, obj *model.ProcessorType) (string, error) {
	return string(obj.GetKind()), nil
}

// Agents is the resolver for the agents field.
func (r *queryResolver) Agents(ctx context.Context, selector *string, query *string) (*model1.Agents, error) {
	ctx, span := tracer.Start(ctx, "graphql/Agents")
	defer span.End()

	options, suggestions, err := r.queryOptionsAndSuggestions(selector, query, r.Resolver.bindplane.Store().AgentIndex())
	agents, err := r.Resolver.bindplane.Store().Agents(ctx, options...)
	if err != nil {
		r.bindplane.Logger().Error("error in graphql Agents", zap.Error(err))
		return nil, err
	}
	return &model1.Agents{
		Agents:      agents,
		Query:       query,
		Suggestions: suggestions,
	}, nil
}

// Agent is the resolver for the agent field.
func (r *queryResolver) Agent(ctx context.Context, id string) (*model.Agent, error) {
	return r.Resolver.bindplane.Store().Agent(id)
}

// Configurations is the resolver for the configurations field.
func (r *queryResolver) Configurations(ctx context.Context, selector *string, query *string) (*model1.Configurations, error) {
	options, suggestions, err := r.queryOptionsAndSuggestions(selector, query, r.Resolver.bindplane.Store().ConfigurationIndex())
	configurations, err := r.Resolver.bindplane.Store().Configurations(options...)
	if err != nil {
		return nil, err
	}
	return &model1.Configurations{
		Configurations: configurations,
		Query:          query,
		Suggestions:    suggestions,
	}, nil
}

// Configuration is the resolver for the configuration field.
func (r *queryResolver) Configuration(ctx context.Context, name string) (*model.Configuration, error) {
	return r.Resolver.bindplane.Store().Configuration(name)
}

// Sources is the resolver for the sources field.
func (r *queryResolver) Sources(ctx context.Context) ([]*model.Source, error) {
	return r.Resolver.bindplane.Store().Sources()
}

// Source is the resolver for the source field.
func (r *queryResolver) Source(ctx context.Context, name string) (*model.Source, error) {
	return r.Resolver.bindplane.Store().Source(name)
}

// SourceTypes is the resolver for the sourceTypes field.
func (r *queryResolver) SourceTypes(ctx context.Context) ([]*model.SourceType, error) {
	return r.Resolver.bindplane.Store().SourceTypes()
}

// SourceType is the resolver for the sourceType field.
func (r *queryResolver) SourceType(ctx context.Context, name string) (*model.SourceType, error) {
	return r.Resolver.bindplane.Store().SourceType(name)
}

// Processors is the resolver for the processors field.
func (r *queryResolver) Processors(ctx context.Context) ([]*model.Processor, error) {
	return r.Resolver.bindplane.Store().Processors()
}

// Processor is the resolver for the processor field.
func (r *queryResolver) Processor(ctx context.Context, name string) (*model.Processor, error) {
	return r.Resolver.bindplane.Store().Processor(name)
}

// ProcessorTypes is the resolver for the processorTypes field.
func (r *queryResolver) ProcessorTypes(ctx context.Context) ([]*model.ProcessorType, error) {
	return r.Resolver.bindplane.Store().ProcessorTypes()
}

// ProcessorType is the resolver for the processorType field.
func (r *queryResolver) ProcessorType(ctx context.Context, name string) (*model.ProcessorType, error) {
	return r.Resolver.bindplane.Store().ProcessorType(name)
}

// Destinations is the resolver for the destinations field.
func (r *queryResolver) Destinations(ctx context.Context) ([]*model.Destination, error) {
	return r.Resolver.bindplane.Store().Destinations()
}

// Destination is the resolver for the destination field.
func (r *queryResolver) Destination(ctx context.Context, name string) (*model.Destination, error) {
	return r.Resolver.bindplane.Store().Destination(name)
}

// DestinationWithType is the resolver for the destinationWithType field.
func (r *queryResolver) DestinationWithType(ctx context.Context, name string) (*model1.DestinationWithType, error) {
	resp := &model1.DestinationWithType{}

	dest, err := r.Resolver.bindplane.Store().Destination(name)
	if err != nil {
		return resp, err
	}

	if dest == nil {
		return resp, nil
	}

	destinationType, err := r.Resolver.bindplane.Store().DestinationType(dest.Spec.Type)
	if err != nil {
		return resp, err
	}

	return &model1.DestinationWithType{
		Destination:     dest,
		DestinationType: destinationType,
	}, nil
}

// DestinationTypes is the resolver for the destinationTypes field.
func (r *queryResolver) DestinationTypes(ctx context.Context) ([]*model.DestinationType, error) {
	return r.Resolver.bindplane.Store().DestinationTypes()
}

// DestinationType is the resolver for the destinationType field.
func (r *queryResolver) DestinationType(ctx context.Context, name string) (*model.DestinationType, error) {
	return r.Resolver.bindplane.Store().DestinationType(name)
}

// Components is the resolver for the components field.
func (r *queryResolver) Components(ctx context.Context) (*model1.Components, error) {
	sources := make([]*model.Source, 0)
	destinations := make([]*model.Destination, 0)
	var err error

	sources, err = r.bindplane.Store().Sources()
	if err != nil {
		return &model1.Components{
			Destinations: destinations,
			Sources:      sources,
		}, err
	}

	destinations, err = r.bindplane.Store().Destinations()
	if err != nil {
		return &model1.Components{
			Destinations: destinations,
			Sources:      sources,
		}, err
	}

	return &model1.Components{
		Destinations: destinations,
		Sources:      sources,
	}, nil
}

// Operator is the resolver for the operator field.
func (r *relevantIfConditionResolver) Operator(ctx context.Context, obj *model.RelevantIfCondition) (model1.RelevantIfOperatorType, error) {
	return model1.RelevantIfOperatorType(obj.Operator), nil
}

// Kind is the resolver for the kind field.
func (r *sourceResolver) Kind(ctx context.Context, obj *model.Source) (string, error) {
	return string(obj.GetKind()), nil
}

// Kind is the resolver for the kind field.
func (r *sourceTypeResolver) Kind(ctx context.Context, obj *model.SourceType) (string, error) {
	return string(obj.GetKind()), nil
}

// AgentChanges is the resolver for the agentChanges field.
func (r *subscriptionResolver) AgentChanges(ctx context.Context, selector *string, query *string) (<-chan []*model1.AgentChange, error) {
	parsedSelector, parsedQuery, err := r.parseSelectorAndQuery(selector, query)
	if err != nil {
		return nil, err
	}

	// we can ignore the unsubscribe function because this will automatically unsubscribe when the context is done. we
	// could subscribe directly to store.AgentChanges, but the resolver is setup to relay events and the filter and
	// dispatch will happen in a separate goroutine.
	channel, _ := eventbus.SubscribeWithFilterUntilDone(ctx, r.updates, func(updates *store.Updates) (result []*model1.AgentChange, accept bool) {
		// if the observer is using a selector or query, we want to change Update to Remove if it no longer matches the
		// selector or query
		events := applySelectorToChanges(parsedSelector, updates.Agents)
		events = applyQueryToChanges(parsedQuery, r.bindplane.Store().AgentIndex(), events)

		return model1.ToAgentChangeArray(events), !events.Empty()
	})

	return channel, nil
}

// ConfigurationChanges is the resolver for the configurationChanges field.
func (r *subscriptionResolver) ConfigurationChanges(ctx context.Context, selector *string, query *string) (<-chan []*model1.ConfigurationChange, error) {
	parsedSelector, parsedQuery, err := r.parseSelectorAndQuery(selector, query)
	if err != nil {
		return nil, err
	}

	// we can ignore the unsubscribe function because this will automatically unsubscribe when the context is done.
	channel, _ := eventbus.SubscribeWithFilterUntilDone(ctx, r.updates, func(updates *store.Updates) (result []*model1.ConfigurationChange, accept bool) {
		// if the observer is using a selector or query, we want to change Update to Remove if it no longer matches the
		// selector or query
		events := applySelectorToEvents(parsedSelector, updates.Configurations)
		events = applyQueryToEvents(parsedQuery, r.bindplane.Store().ConfigurationIndex(), events)

		return model1.ToConfigurationChanges(events), len(events) > 0
	})

	return channel, nil
}

// Livetail is the resolver for the livetail field.
func (r *subscriptionResolver) Livetail(ctx context.Context, agentIds []string, filters []string) (<-chan []*model1.LiveTailMessage, error) {
	// TODO (observathon)
	panic(fmt.Errorf("not implemented"))
}

// Agent returns generated.AgentResolver implementation.
func (r *Resolver) Agent() generated.AgentResolver { return &agentResolver{r} }

// AgentSelector returns generated.AgentSelectorResolver implementation.
func (r *Resolver) AgentSelector() generated.AgentSelectorResolver { return &agentSelectorResolver{r} }

// Configuration returns generated.ConfigurationResolver implementation.
func (r *Resolver) Configuration() generated.ConfigurationResolver { return &configurationResolver{r} }

// Destination returns generated.DestinationResolver implementation.
func (r *Resolver) Destination() generated.DestinationResolver { return &destinationResolver{r} }

// DestinationType returns generated.DestinationTypeResolver implementation.
func (r *Resolver) DestinationType() generated.DestinationTypeResolver {
	return &destinationTypeResolver{r}
}

// Metadata returns generated.MetadataResolver implementation.
func (r *Resolver) Metadata() generated.MetadataResolver { return &metadataResolver{r} }

// ParameterDefinition returns generated.ParameterDefinitionResolver implementation.
func (r *Resolver) ParameterDefinition() generated.ParameterDefinitionResolver {
	return &parameterDefinitionResolver{r}
}

// Processor returns generated.ProcessorResolver implementation.
func (r *Resolver) Processor() generated.ProcessorResolver { return &processorResolver{r} }

// ProcessorType returns generated.ProcessorTypeResolver implementation.
func (r *Resolver) ProcessorType() generated.ProcessorTypeResolver { return &processorTypeResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// RelevantIfCondition returns generated.RelevantIfConditionResolver implementation.
func (r *Resolver) RelevantIfCondition() generated.RelevantIfConditionResolver {
	return &relevantIfConditionResolver{r}
}

// Source returns generated.SourceResolver implementation.
func (r *Resolver) Source() generated.SourceResolver { return &sourceResolver{r} }

// SourceType returns generated.SourceTypeResolver implementation.
func (r *Resolver) SourceType() generated.SourceTypeResolver { return &sourceTypeResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type agentResolver struct{ *Resolver }
type agentSelectorResolver struct{ *Resolver }
type configurationResolver struct{ *Resolver }
type destinationResolver struct{ *Resolver }
type destinationTypeResolver struct{ *Resolver }
type metadataResolver struct{ *Resolver }
type parameterDefinitionResolver struct{ *Resolver }
type processorResolver struct{ *Resolver }
type processorTypeResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type relevantIfConditionResolver struct{ *Resolver }
type sourceResolver struct{ *Resolver }
type sourceTypeResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
