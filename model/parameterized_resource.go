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

	"github.com/observiq/bindplane-op/model/otel"
	"github.com/observiq/bindplane-op/model/validation"
)

// ParameterizedSpec is the spec for a ParameterizedResource
type ParameterizedSpec struct {
	Type       string                  `yaml:"type" json:"type" mapstructure:"type"`
	Parameters []Parameter             `yaml:"parameters" json:"parameters" mapstructure:"parameters"`
	Processors []ResourceConfiguration `yaml:"processors" json:"processors" mapstructure:"processors"`
}

// parameterizedResource is a resource based on a resource type which provides a specific resource value via templated
type parameterizedResource interface {
	otel.ComponentIDProvider

	// Name returns the name for this resource
	Name() string

	// ResourceTypeName is the name of the ResourceType that renders this resource type
	ResourceTypeName() string

	// ResourceParameters are the parameters passed to the ResourceType to generate the configuration
	ResourceParameters() []Parameter
}

// overrideParameters overrides the parameters in the spec and returns a new spec with the overrides applied
func (s ParameterizedSpec) overrideParameters(parameters []Parameter) ParameterizedSpec {
	result := make([]Parameter, len(s.Parameters))
	nameIndex := map[string]int{}
	for i, p := range s.Parameters {
		result[i] = p
		nameIndex[p.Name] = i
	}
	for _, p := range parameters {
		index, ok := nameIndex[p.Name]
		if ok {
			// override
			s.Parameters[index] = p
		} else {
			// append
			s.Parameters = append(s.Parameters, p)
		}
	}
	return ParameterizedSpec{Type: s.Type, Parameters: result, Processors: s.Processors}
}

// validateTypeAndParameters is used by Source and Destination validation and uses methods created for Configuration
// validation.
func (s *ParameterizedSpec) validateTypeAndParameters(kind Kind, errors validation.Errors, store ResourceStore) {
	fmt.Printf("validate %s %s with %d processors\n", kind, s.Type, len(s.Processors))
	// ResourceConfiguration is a resource embedded in a Configuration, but it works equally well for Source and
	// Destination validation.
	rc := &ResourceConfiguration{
		Type:       s.Type,
		Parameters: s.Parameters,
		Processors: s.Processors,
	}
	rc.validateParameters(kind, errors, store)
	rc.validateProcessors(kind, errors, store)
}
