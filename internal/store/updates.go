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
	"github.com/hashicorp/go-multierror"
	"github.com/observiq/bindplane-op/model"
)

// Updates are sent on the channel available from Store.Updates().
type Updates struct {
	Agents           Events[*model.Agent]
	Sources          Events[*model.Source]
	SourceTypes      Events[*model.SourceType]
	Destinations     Events[*model.Destination]
	DestinationTypes Events[*model.DestinationType]
	Configurations   Events[*model.Configuration]
}

// NewUpdates returns a New Updates struct
func NewUpdates() *Updates {
	// TODO: optimize allocate as needed
	return &Updates{
		Agents:           NewEvents[*model.Agent](),
		Sources:          NewEvents[*model.Source](),
		SourceTypes:      NewEvents[*model.SourceType](),
		Destinations:     NewEvents[*model.Destination](),
		DestinationTypes: NewEvents[*model.DestinationType](),
		Configurations:   NewEvents[*model.Configuration](),
	}
}

// IncludeAgent will include the agent in the updates. While updates.Agents.Include can also be used directly, this
// matches the pattern of IncludeResource.
func (updates *Updates) IncludeAgent(agent *model.Agent, eventType EventType) {
	updates.Agents.Include(agent, eventType)
}

// IncludeResource will include the resource in the updates for the appropriate type. If the specified Resource is not
// supported by Updates, this will do nothing.
func (updates *Updates) IncludeResource(r model.Resource, eventType EventType) {
	switch r := r.(type) {
	case *model.Source:
		updates.Sources.Include(r, eventType)
	case *model.SourceType:
		updates.SourceTypes.Include(r, eventType)
	case *model.Destination:
		updates.Destinations.Include(r, eventType)
	case *model.DestinationType:
		updates.DestinationTypes.Include(r, eventType)
	case *model.Configuration:
		updates.Configurations.Include(r, eventType)
	}
}

// Empty returns true if all individual updates are empty
func (updates *Updates) Empty() bool {
	return updates.Size() == 0
}

// Size returns the sum of all updates of all types
func (updates *Updates) Size() int {
	return len(updates.Agents) +
		len(updates.Sources) +
		len(updates.SourceTypes) +
		len(updates.Destinations) +
		len(updates.DestinationTypes) +
		len(updates.Configurations)
}

// ----------------------------------------------------------------------
//
// add transitive updates based on updates that already exist. this could be optimized for a specific store and may not
// be used by all stores.

// TODO: how does this work in a distributed environment?
// pub/sub individual event => pub/sub include dependencies => subscribers
func (updates *Updates) addTransitiveUpdates(s Store) error {
	// for sourceTypes, add sources
	// for destinationTypes, add destinations
	// for sources and sourceTypes, add configurations
	// for destinations and destinationTypes, add configurations

	var errs error

	err := updates.addSourceUpdates(s)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	err = updates.addDestinationUpdates(s)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	err = updates.addConfigurationUpdates(s)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs
}

func (updates *Updates) addSourceUpdates(s Store) error {
	if updates.SourceTypes.Empty() {
		return nil
	}

	// get all of the sources
	sources, err := s.Sources()
	if err != nil {
		return err
	}

	// updates to a SourceType will trigger updates of all of the Sources that use that SourceType.
	for _, sourceTypeEvent := range updates.SourceTypes {
		if sourceTypeEvent.Type == EventTypeUpdate {
			sourceTypeName := sourceTypeEvent.Item.Name()

			for _, source := range sources {
				if source.Spec.Type == sourceTypeName {
					updates.Sources.Include(source, EventTypeUpdate)
				}
			}
		}
	}

	return nil
}

func (updates *Updates) addDestinationUpdates(s Store) error {
	if updates.DestinationTypes.Empty() {
		return nil
	}

	// get all of the destinations
	destinations, err := s.Destinations()
	if err != nil {
		return err
	}

	// updates to a DestinationType will trigger updates of all of the Destinations that use that DestinationType.
	for _, destinationTypeEvent := range updates.DestinationTypes {
		if destinationTypeEvent.Type == EventTypeUpdate {
			destinationTypeName := destinationTypeEvent.Item.Name()

			for _, destination := range destinations {
				if destination.Spec.Type == destinationTypeName {
					updates.Destinations.Include(destination, EventTypeUpdate)
				}
			}
		}
	}

	return nil
}

func (updates *Updates) addConfigurationUpdates(s Store) error {
	configurations, err := s.Configurations()
	if err != nil {
		return err
	}

	for _, configuration := range configurations {
		// as a small optimization, before checking all of the sources and destinations for changes, check to see if we're
		// already updating this configuration.
		if updates.Configurations.Contains(configuration.Name(), EventTypeUpdate) {
			continue
		}
		updates.addConfigurationUpdatesFromComponents(configuration, s)
	}
	return nil
}

func (updates *Updates) addConfigurationUpdatesFromComponents(configuration *model.Configuration, s Store) {
	for _, source := range configuration.Spec.Sources {
		if _, ok := updates.Sources[source.Name]; ok {
			updates.Configurations.Include(configuration, EventTypeUpdate)
			return
		}
		if _, ok := updates.SourceTypes[source.Type]; ok {
			updates.Configurations.Include(configuration, EventTypeUpdate)
			return
		}
	}
	for _, destination := range configuration.Spec.Destinations {
		if _, ok := updates.Destinations[destination.Name]; ok {
			updates.Configurations.Include(configuration, EventTypeUpdate)
			return
		}
		if _, ok := updates.DestinationTypes[destination.Type]; ok {
			updates.Configurations.Include(configuration, EventTypeUpdate)
			return
		}
	}
}

// ----------------------------------------------------------------------
// merge for use with RelayWithMerge

func mergeUpdates(into, single *Updates) bool {
	// first make sure we can safely merge
	safe := into.Agents.CanSafelyMerge(single.Agents) &&
		into.Sources.CanSafelyMerge(single.Sources) &&
		into.SourceTypes.CanSafelyMerge(single.SourceTypes) &&
		into.Destinations.CanSafelyMerge(single.Destinations) &&
		into.DestinationTypes.CanSafelyMerge(single.DestinationTypes) &&
		into.Configurations.CanSafelyMerge(single.Configurations)

	if !safe {
		return false
	}

	// merge individual events
	into.Agents.Merge(single.Agents)
	into.Sources.Merge(single.Sources)
	into.SourceTypes.Merge(single.SourceTypes)
	into.Destinations.Merge(single.Destinations)
	into.DestinationTypes.Merge(single.DestinationTypes)
	into.Configurations.Merge(single.Configurations)

	return true
}
