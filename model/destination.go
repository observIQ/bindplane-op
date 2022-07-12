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

// Destination will generate an exporter and be at the end of a pipeline
type Destination struct {
	// ResourceMeta TODO(doc)
	ResourceMeta `yaml:",inline" json:",inline" mapstructure:",squash"`
	// Spec TODO(doc)
	Spec ParameterizedSpec `json:"spec" yaml:"spec" mapstructure:"spec"`
}

var _ parameterizedResource = (*Destination)(nil)

// ValidateWithStore checks that the destination is valid, returning an error if it is not. It uses the store to
// retreive the destination type so that parameter values can be validated against the parameter defintions.
func (d *Destination) ValidateWithStore(store ResourceStore) error {
	errors := validation.NewErrors()

	d.validate(errors)
	d.Spec.validateTypeAndParameters(KindDestination, errors, store)

	return errors.Result()
}

// GetKind returns "Destination"
func (d *Destination) GetKind() Kind { return KindDestination }

// ResourceTypeName is the name of the ResourceType that renders this resource type
func (d *Destination) ResourceTypeName() string {
	return d.Spec.Type
}

// ResourceParameters are the parameters passed to the ResourceType to generate the configuration
func (d *Destination) ResourceParameters() []Parameter {
	return d.Spec.Parameters
}

// ComponentID provides a unique component id for the specified component name
func (d *Destination) ComponentID(name string) otel.ComponentID {
	return otel.UniqueComponentID(name, d.Spec.Type, d.Name())
}

// NewDestination creates a new Destination with the specified name, type, and parameters
func NewDestination(name string, typeValue string, parameters []Parameter) *Destination {
	return NewDestinationWithSpec(name, ParameterizedSpec{
		Type:       typeValue,
		Parameters: parameters,
	})
}

// NewDestinationWithSpec creates a new Destination with the specified spec
func NewDestinationWithSpec(name string, spec ParameterizedSpec) *Destination {
	return &Destination{
		ResourceMeta: ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1beta",
			Kind:       KindDestination,
			Metadata: Metadata{
				Name:   name,
				Labels: MakeLabels(),
			},
		},
		Spec: spec,
	}
}

// FindDestination returns a Destination from the store if it exists. If it doesn't exist, it creates a new Destination with the
// specified defaultName.
func FindDestination(destination *ResourceConfiguration, defaultName string, store ResourceStore) (*Destination, error) {
	if destination.Name == "" {
		// inline destination
		return NewDestination(defaultName, destination.Type, destination.Parameters), nil
	}
	// find the destination and override parameters
	dest, err := store.Destination(destination.Name)
	if err != nil {
		return nil, err
	}
	if dest == nil {
		return nil, fmt.Errorf("unknown %s: %s", KindDestination, destination.Name)
	}
	spec := dest.Spec.overrideParameters(destination.Parameters)
	return NewDestinationWithSpec(dest.Name(), spec), nil
}

// ----------------------------------------------------------------------

// PrintableFieldTitles returns the list of field titles, used for printing a table of resources
func (d *Destination) PrintableFieldTitles() []string {
	return []string{"Name", "Type", "Description"}
}

// PrintableFieldValue returns the field value for a title, used for printing a table of resources
func (d *Destination) PrintableFieldValue(title string) string {
	switch title {
	case "ID":
		return d.ID()
	case "Name":
		return d.Name()
	case "Type":
		return d.ResourceTypeName()
	case "Description":
		return d.Metadata.Description
	default:
		return "-"
	}
}
