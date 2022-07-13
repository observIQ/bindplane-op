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

// Processor will generate an exporter and be at the end of a pipeline
type Processor struct {
	// ResourceMeta TODO(doc)
	ResourceMeta `yaml:",inline" json:",inline" mapstructure:",squash"`
	// Spec TODO(doc)
	Spec ParameterizedSpec `json:"spec" yaml:"spec" mapstructure:"spec"`
}

var _ parameterizedResource = (*Processor)(nil)

// ValidateWithStore checks that the processor is valid, returning an error if it is not. It uses the store to retrieve the
// processor type so that parameter values can be validated against the parameter definitions.
func (s *Processor) ValidateWithStore(store ResourceStore) error {
	errors := validation.NewErrors()

	s.validate(errors)
	s.Spec.validateTypeAndParameters(KindProcessor, errors, store)

	return errors.Result()
}

// GetKind returns "Processor"
func (s *Processor) GetKind() Kind { return KindProcessor }

// ResourceTypeName is the name of the ResourceType that renders this resource type
func (s *Processor) ResourceTypeName() string {
	return s.Spec.Type
}

// ResourceParameters are the parameters passed to the ResourceType to generate the configuration
func (s *Processor) ResourceParameters() []Parameter {
	return s.Spec.Parameters
}

// ComponentID provides a unique component id for the specified component name
func (s *Processor) ComponentID(name string) otel.ComponentID {
	return otel.UniqueComponentID(name, s.Spec.Type, s.Name())
}

// NewProcessor creates a new Processor with the specified name, type, and parameters
func NewProcessor(name string, processorTypeName string, parameters []Parameter) *Processor {
	return NewProcessorWithSpec(name, ParameterizedSpec{
		Type:       processorTypeName,
		Parameters: parameters,
	})
}

// NewProcessorWithSpec creates a new Processor with the specified spec
func NewProcessorWithSpec(name string, spec ParameterizedSpec) *Processor {
	return &Processor{
		ResourceMeta: ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1beta",
			Kind:       KindProcessor,
			Metadata: Metadata{
				Name:   name,
				Labels: MakeLabels(),
			},
		},
		Spec: spec,
	}
}

// FindProcessor returns a Processor from the store if it exists. If it doesn't exist, it creates a new Processor with the
// specified defaultName.
func FindProcessor(processor *ResourceConfiguration, defaultName string, store ResourceStore) (*Processor, error) {
	if processor.Name == "" {
		// inline source
		return NewProcessor(defaultName, processor.Type, processor.Parameters), nil
	}
	// find the processor and override parameters
	prc, err := store.Processor(processor.Name)
	if err != nil {
		return nil, err
	}
	if prc == nil {
		return nil, fmt.Errorf("unknown %s: %s", KindProcessor, processor.Name)
	}
	spec := prc.Spec.overrideParameters(processor.Parameters)
	return NewProcessorWithSpec(prc.Name(), spec), nil
}

// ----------------------------------------------------------------------

// PrintableFieldTitles returns the list of field titles, used for printing a table of resources
func (s *Processor) PrintableFieldTitles() []string {
	return []string{"Name", "Type", "Description"}
}

// PrintableFieldValue returns the field value for a title, used for printing a table of resources
func (s *Processor) PrintableFieldValue(title string) string {
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
