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
	"github.com/observiq/bindplane-op/model"
)

// EventType is the type of event
type EventType uint8

// Insert, Update, Remove, and Label are the possible changes to a resource. Remove can indicate that the resource has
// been deleted or that it no longer matches a filtered list of resources being observed.
//
// Label is currently used to indicate that Agent labels have changed but there may also be other updates so it can
// be considered a subset of EventTypeUpdate where every EventTypeLabel is also an EventTypeUpdate.
const (
	EventTypeInsert EventType = 1
	EventTypeUpdate EventType = 2
	EventTypeRemove EventType = 3
	EventTypeLabel  EventType = 4
)

// Event represents an insert, update, or remove of something stored in Store.
type Event[T model.HasUniqueKey] struct {
	Type EventType `json:"type"`
	Item T         `json:"item"`
}

// Events is a map of ID to Event
type Events[T model.HasUniqueKey] map[string]Event[T]

// NewEvents returns a new empty set of events
func NewEvents[T model.HasUniqueKey]() Events[T] {
	return Events[T](map[string]Event[T]{})
}

// NewEventsWithItem returns a new set of events with an initial item and eventType
func NewEventsWithItem[T model.HasUniqueKey](item T, eventType EventType) Events[T] {
	e := NewEvents[T]()
	e.Include(item, eventType)
	return e
}

// Empty returns true if there are no events
func (e Events[T]) Empty() bool {
	return len(e) == 0
}

// Contains returns true if the item already exists
func (e Events[T]) Contains(uniqueKey string, eventType EventType) bool {
	if uniqueKey == "" {
		return false
	}
	if event, ok := e[uniqueKey]; ok {
		return event.Type == eventType
	}
	return false
}

func (e Events[T]) Keys() []string {
	result := make([]string, 0, len(e))
	for k := range e {
		result = append(result, k)
	}
	return result
}

// Include an item of with the specified event type.
func (e Events[T]) Include(item T, eventType EventType) {
	key := item.UniqueKey()
	e[key] = Event[T]{
		Item: item,
		Type: eventType,
	}
}

func (e Events[T]) Updates() []Event[T] {
	return e.ByType(EventTypeUpdate)
}

func (e Events[T]) ByType(eventType EventType) []Event[T] {
	var results []Event[T]
	for _, event := range e {
		if event.Type == eventType {
			results = append(results, event)
		}
	}
	return results
}
