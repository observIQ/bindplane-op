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

package model

import (
	"github.com/observiq/bindplane-op/internal/store"
	"github.com/observiq/bindplane-op/model"
)

// ToAgentChangeArray converts store.AgentChanges to a []*AgentChange for use with graphql
func ToAgentChangeArray(changes store.Events[*model.Agent]) []*AgentChange {
	result := []*AgentChange{}
	for _, change := range changes {
		result = append(result, ToAgentChange(change))
	}
	return result
}

// ToAgentChange converts a store.AgentChange to use for use with graphql
func ToAgentChange(change store.Event[*model.Agent]) *AgentChange {
	agentChangeType := AgentChangeTypeInsert
	switch change.Type {
	case store.EventTypeInsert:
		agentChangeType = AgentChangeTypeInsert
	case store.EventTypeUpdate:
		agentChangeType = AgentChangeTypeUpdate
	case store.EventTypeRemove:
		agentChangeType = AgentChangeTypeRemove
	case store.EventTypeLabel:
		agentChangeType = AgentChangeTypeUpdate
	}
	return &AgentChange{
		Agent:      change.Item,
		ChangeType: agentChangeType,
	}
}

// ToConfigurationChanges converts store.Events for Configuration to an array of ConfigurationChange for use with graphql
func ToConfigurationChanges(events store.Events[*model.Configuration]) []*ConfigurationChange {
	result := []*ConfigurationChange{}
	for _, event := range events {
		result = append(result, ToConfigurationChange(event))
	}
	return result
}

// ToConfigurationChange converts a store.Event for Configuration to a ConfigurationChange for use with graphql
func ToConfigurationChange(event store.Event[*model.Configuration]) *ConfigurationChange {
	return &ConfigurationChange{
		Configuration: event.Item,
		EventType:     ToEventType(event.Type),
	}
}

// ToEventType converts a store.EventType to a graphql EventType
func ToEventType(eventType store.EventType) EventType {
	switch eventType {
	case store.EventTypeInsert:
		return EventTypeInsert
	case store.EventTypeUpdate:
		return EventTypeUpdate
	case store.EventTypeRemove:
		return EventTypeRemove
	case store.EventTypeLabel:
		return EventTypeUpdate
	}
	return EventTypeUpdate
}
