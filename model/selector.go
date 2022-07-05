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
	"fmt"

	"github.com/observiq/bindplane-op/model/validation"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// HasAgentSelector TODO(docs)
type HasAgentSelector interface {
	// AgentSelector TODO(docs)
	AgentSelector() Selector
	IsForAgent(agent *Agent) bool
}

// Selector TODO(docs)
type Selector struct {
	labels.Selector
}

// AgentSelector specifies a selector to use to match resources to agents
type AgentSelector struct {
	MatchLabels `json:"matchLabels" yaml:"matchLabels" mapstructure:"matchLabels"`
}

// MatchLabels represents the labels used to match Pipelines with Agents.
type MatchLabels map[string]string

// SelectorFromString takes a string and returns a Selector and error
func SelectorFromString(selector string) (Selector, error) {
	l, err := labels.ConvertSelectorToLabelsMap(selector)
	if err != nil {
		return EmptySelector(), err
	}
	return Selector{l.AsSelector()}, nil
}

// SelectorFromMap takes a map[string]string and returns a Selector and error.
func SelectorFromMap(m map[string]string) (Selector, error) {
	labels, err := LabelsFromMap(m)
	if err != nil {
		return EmptySelector(), err
	}
	selector, err := labels.AsValidatedSelector()
	if err != nil {
		return EmptySelector(), err
	}
	return Selector{selector}, nil
}

// EmptySelector returns a Selector that has no labels and matches nothing
func EmptySelector() Selector {
	return Selector{labels.Nothing()}
}

// EverythingSelector returns a Selector that matches everything
func EverythingSelector() Selector {
	return Selector{labels.Everything()}
}

// isResourceForAgent returns true if the resource selector matches a given agent's labels.
func isResourceForAgent(hasAgentSelector HasAgentSelector, agent *Agent) bool {
	return hasAgentSelector.AgentSelector().Matches(agent.Labels)
}

// Selector creates a Selector struct from an AgentSelector
func (s AgentSelector) Selector() Selector {
	selector, err := SelectorFromMap(s.MatchLabels)
	if err != nil {
		return EmptySelector()
	}
	return selector
}

// validate ensures that the selector is valid
func (s AgentSelector) validate(errors validation.Errors) {
	_, err := SelectorFromMap(s.MatchLabels)
	if err != nil {
		errors.Add(fmt.Errorf("selector is invalid: %w", err))
	}
}

// MatchLabels returns the portion of the Selector that consists of simple name=value label matches. A Selector supports
// more complex selection and this should only be used in cases where Matches() would have terrible performance and
// partial selection is ok. This will return false for complete if there are selectors requirements that cannot be
// expressed with match labels.
func (s *Selector) MatchLabels() (labels MatchLabels, complete bool) {
	selectorRequirements, selectable := s.Selector.Requirements()
	if !selectable {
		// not selectable means this selects nothing and Matches will always be false.
		return nil, false
	}
	complete = true
	labels = MatchLabels{}
	for _, r := range selectorRequirements {
		op := r.Operator()
		if op == selection.Equals || op == selection.DoubleEquals {
			values := r.Values().UnsortedList()
			if len(values) > 0 {
				labels[r.Key()] = values[0]
			}
		} else {
			// operator not supported with simple match labels
			complete = false
		}
	}
	return labels, complete
}
