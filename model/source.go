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

// Source will generate an exporter and be at the end of a pipeline
type Source struct {
	// ResourceMeta TODO(doc)
	ResourceMeta `yaml:",inline" json:",inline" mapstructure:",squash"`
	// Spec TODO(doc)
	Spec ParameterizedSpec `json:"spec" yaml:"spec" mapstructure:"spec"`
}

var _ parameterizedResource = (*Source)(nil)

// ValidateWithStore checks that the source is valid, returning an error if it is not. It uses the store to retreive the
// source type so that parameter values can be validated against the parameter defintions.
func (s *Source) ValidateWithStore(store ResourceStore) error {
	errors := validation.NewErrors()

	s.validate(errors)
	s.Spec.validateTypeAndParameters(KindSource, errors, store)

	return errors.Result()
}

// GetKind returns "Source"
func (s *Source) GetKind() Kind { return KindSource }

// ResourceTypeName is the name of the ResourceType that renders this resource type
func (s *Source) ResourceTypeName() string {
	return s.Spec.Type
}

// ResourceParameters are the parameters passed to the ResourceType to generate the configuration
func (s *Source) ResourceParameters() []Parameter {
	return s.Spec.Parameters
}

// ComponentID provides a unique component id for the specified component name
func (s *Source) ComponentID(name string) otel.ComponentID {
	return otel.UniqueComponentID(name, s.Spec.Type, s.Name())
}

// NewSource creates a new Source with the specified name, type, and parameters
func NewSource(name string, sourceTypeName string, parameters []Parameter) *Source {
	return NewSourceWithSpec(name, ParameterizedSpec{
		Type:       sourceTypeName,
		Parameters: parameters,
	})
}

// NewSourceWithSpec creates a new Source with the specified spec
func NewSourceWithSpec(name string, spec ParameterizedSpec) *Source {
	return &Source{
		ResourceMeta: ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1beta",
			Kind:       KindSource,
			Metadata: Metadata{
				Name:   name,
				Labels: MakeLabels(),
			},
		},
		Spec: spec,
	}
}

// FindSource returns a Source from the store if it exists. If it doesn't exist, it creates a new Source with the
// specified defaultName.
func FindSource(source *ResourceConfiguration, defaultName string, store ResourceStore) (*Source, error) {
	if source.Name == "" {
		// inline source
		return NewSource(defaultName, source.Type, source.Parameters), nil
	}
	// find the source and override parameters
	src, err := store.Source(source.Name)
	if err != nil {
		return nil, err
	}
	if src == nil {
		return nil, fmt.Errorf("unknown %s: %s", KindSource, source.Name)
	}
	spec := src.Spec.overrideParameters(source.Parameters)
	return NewSourceWithSpec(src.Name(), spec), nil
}

// ----------------------------------------------------------------------

// PrintableFieldTitles returns the list of field titles, used for printing a table of resources
func (s *Source) PrintableFieldTitles() []string {
	return []string{"Name", "Type", "Description"}
}

// PrintableFieldValue returns the field value for a title, used for printing a table of resources
func (s *Source) PrintableFieldValue(title string) string {
	switch title {
	case "ID":
		return s.ID()
	case "Name":
		return s.Name()
	case "Type":
		return s.ResourceTypeName()
	case "Description":
		return s.Metadata.Description
	default:
		return "-"
	}
}
