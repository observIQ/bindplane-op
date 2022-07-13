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

package model

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/observiq/bindplane-op/internal/store/search"
	"github.com/observiq/bindplane-op/model/otel"
	"github.com/observiq/bindplane-op/model/validation"
	otelExt "go.opentelemetry.io/otel"
	"gopkg.in/yaml.v3"
)

var tracer = otelExt.Tracer("model/configuration")

// ConfigurationType indicates the kind of configuration. It is based on the presence of the Raw, Sources, and
// Destinations fields.
type ConfigurationType string

const (
	// ConfigurationTypeRaw configurations have a configuration in the Raw field that is passed directly to the agent.
	ConfigurationTypeRaw ConfigurationType = "raw"

	// ConfigurationTypeModular configurations have Sources and Destinations that are used to generate the configuration to pass to an agent.
	ConfigurationTypeModular = "modular"
	// TODO(andy): Do we like Modular for configurations with Sources/Destinations?
)

// Configuration is the resource for the entire agent configuration
type Configuration struct {
	// ResourceMeta TODO(doc)
	ResourceMeta `yaml:",inline" json:",inline" mapstructure:",squash"`
	// Spec TODO(doc)
	Spec ConfigurationSpec `json:"spec" yaml:"spec" mapstructure:"spec"`
}

var _ HasAgentSelector = (*Configuration)(nil)

// NewConfiguration creates a new configuration with the specified name
func NewConfiguration(name string) *Configuration {
	return NewConfigurationWithSpec(name, ConfigurationSpec{})
}

// NewRawConfiguration creates a new configuration with the specified name and raw configuration
func NewRawConfiguration(name string, raw string) *Configuration {
	return NewConfigurationWithSpec(name, ConfigurationSpec{
		Raw: raw,
	})
}

// NewConfigurationWithSpec creates a new configuration with the specified name and spec
func NewConfigurationWithSpec(name string, spec ConfigurationSpec) *Configuration {
	return &Configuration{
		ResourceMeta: ResourceMeta{
			APIVersion: V1Alpha,
			Kind:       KindConfiguration,
			Metadata: Metadata{
				Name:   name,
				Labels: MakeLabels(),
			},
		},
		Spec: spec,
	}
}

// GetKind returns "Configuration"
func (c *Configuration) GetKind() Kind {
	return KindConfiguration
}

// ConfigurationSpec is the spec for a configuration resource
type ConfigurationSpec struct {
	ContentType  string                  `json:"contentType" yaml:"contentType" mapstructure:"contentType"`
	Raw          string                  `json:"raw,omitempty" yaml:"raw,omitempty" mapstructure:"raw"`
	Sources      []ResourceConfiguration `json:"sources,omitempty" yaml:"sources,omitempty" mapstructure:"sources"`
	Destinations []ResourceConfiguration `json:"destinations,omitempty" yaml:"destinations,omitempty" mapstructure:"destinations"`
	Selector     AgentSelector           `json:"selector" yaml:"selector" mapstructure:"selector"`
}

// ResourceConfiguration represents a source or destination configuration
type ResourceConfiguration struct {
	Name       string                  `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name"`
	Type       string                  `json:"type,omitempty" yaml:"type,omitempty" mapstructure:"type"`
	Parameters []Parameter             `json:"parameters,omitempty" yaml:"parameters,omitempty" mapstructure:"parameters"`
	Processors []ResourceConfiguration `json:"processors,omitempty" yaml:"processors,omitempty" mapstructure:"processors"`
}

// Validate validates most of the configuration, but if a store is available, ValidateWithStore should be used to
// validate the sources and destinations.
func (c *Configuration) Validate() error {
	errors := validation.NewErrors()
	c.validate(errors)
	return errors.Result()
}

func (c *Configuration) validate(errs validation.Errors) {
	c.ResourceMeta.validate(errs)
	c.Spec.validate(errs)
}

// ValidateWithStore checks that the configuration is valid, returning an error if it is not. It uses the store to
// retrieve source types and destination types so that parameter values can be validated against the parameter
// definitions.
func (c *Configuration) ValidateWithStore(store ResourceStore) error {
	errors := validation.NewErrors()

	c.validate(errors)
	c.Spec.validateSourcesAndDestinations(errors, store)

	return errors.Result()
}

// Type returns the ConfigurationType. It is based on the presence of the Raw, Sources, and Destinations fields.
func (c *Configuration) Type() ConfigurationType {
	if c.Spec.Raw != "" {
		// we always prefer raw
		return ConfigurationTypeRaw
	}
	return ConfigurationTypeModular
}

// AgentSelector returns the Selector for this configuration that can be used to match this resource to agents.
func (c *Configuration) AgentSelector() Selector {
	return c.Spec.Selector.Selector()
}

// IsForAgent returns true if this configuration matches a given agent's labels.
func (c *Configuration) IsForAgent(agent *Agent) bool {
	return isResourceForAgent(c, agent)
}

// ResourceStore provides access to resources required to render configurations that use Sources and Destinations.
type ResourceStore interface {
	Source(name string) (*Source, error)
	SourceType(name string) (*SourceType, error)
	Processor(name string) (*Processor, error)
	ProcessorType(name string) (*ProcessorType, error)
	Destination(name string) (*Destination, error)
	DestinationType(name string) (*DestinationType, error)
}

// Render converts the Configuration model to a configuration that can be sent to an agent
func (c *Configuration) Render(ctx context.Context, store ResourceStore) (string, error) {
	ctx, span := tracer.Start(ctx, "model/Configuration/Render")
	defer span.End()

	if c.Spec.Raw != "" {
		// we always prefer raw
		return c.Spec.Raw, nil
	}
	return c.renderComponents(store)
}

func (c *Configuration) renderComponents(store ResourceStore) (string, error) {
	configuration, err := c.otelConfiguration(store)
	if err != nil {
		return "", err
	}
	return configuration.YAML()
}

func (c *Configuration) otelConfiguration(store ResourceStore) (*otel.Configuration, error) {
	if len(c.Spec.Sources) == 0 || len(c.Spec.Destinations) == 0 {
		return nil, nil
	}

	configuration := otel.NewConfiguration()

	// match each source with each destination to produce a pipeline
	sources, destinations, err := c.evalComponents(store)
	if err != nil {
		return nil, err
	}

	for sourceName, source := range sources {
		for destinationName, destination := range destinations {
			name := fmt.Sprintf("%s__%s", sourceName, destinationName)
			configuration.AddPipeline(name, otel.Logs, source, destination)
			configuration.AddPipeline(name, otel.Metrics, source, destination)
			configuration.AddPipeline(name, otel.Traces, source, destination)
		}
	}

	return configuration, nil
}

func (c *Configuration) evalComponents(store ResourceStore) (sources map[string]otel.Partials, destinations map[string]otel.Partials, err error) {
	errorHandler := func(e error) {
		if e != nil {
			err = multierror.Append(err, e)
		}
	}

	sources = map[string]otel.Partials{}
	destinations = map[string]otel.Partials{}

	for i, source := range c.Spec.Sources {
		source := source // copy to local variable to securely pass a reference to a loop variable
		sourceName, srcParts := evalSource(&source, fmt.Sprintf("source%d", i), store, errorHandler)
		sources[sourceName] = srcParts
	}

	for i, destination := range c.Spec.Destinations {
		destination := destination // copy to local variable to securely pass a reference to a loop variable
		destName, destParts := evalDestination(&destination, fmt.Sprintf("destination%d", i), store, errorHandler)
		destinations[destName] = destParts
	}

	return sources, destinations, err
}

func evalSource(source *ResourceConfiguration, defaultName string, store ResourceStore, errorHandler TemplateErrorHandler) (string, otel.Partials) {
	src, srcType, err := findSourceAndType(source, defaultName, store)
	if err != nil {
		errorHandler(err)
		return "", nil
	}

	srcName := fmt.Sprintf("%s__%s", src.Spec.Type, src.Name())
	partials := srcType.eval(src, errorHandler)

	// evaluate the processors associated with the source
	for i, processor := range source.Processors {
		processor := processor
		_, processorParts := evalProcessor(&processor, fmt.Sprintf("%s__processor%d", srcName, i), store, errorHandler)
		if processorParts == nil {
			continue
		}
		partials.Add(processorParts)
	}

	return srcName, partials
}

func evalProcessor(processor *ResourceConfiguration, defaultName string, store ResourceStore, errorHandler TemplateErrorHandler) (string, otel.Partials) {
	prc, prcType, err := findProcessorAndType(processor, defaultName, store)
	if err != nil {
		errorHandler(err)
		return "", nil
	}

	return prc.Name(), prcType.eval(prc, errorHandler)
}

func evalDestination(destination *ResourceConfiguration, defaultName string, store ResourceStore, errorHandler TemplateErrorHandler) (string, otel.Partials) {
	dest, destType, err := findDestinationAndType(destination, defaultName, store)
	if err != nil {
		errorHandler(err)
		return "", nil
	}

	return dest.Name(), destType.eval(dest, errorHandler)
}

func findSourceAndType(source *ResourceConfiguration, defaultName string, store ResourceStore) (*Source, *SourceType, error) {
	src, err := FindSource(source, defaultName, store)
	if err != nil {
		return nil, nil, err
	}

	srcType, err := store.SourceType(src.Spec.Type)
	if err == nil && srcType == nil {
		err = fmt.Errorf("unknown %s: %s", KindSourceType, src.Spec.Type)
	}
	if err != nil {
		return src, nil, err
	}

	return src, srcType, nil
}

func findProcessorAndType(source *ResourceConfiguration, defaultName string, store ResourceStore) (*Processor, *ProcessorType, error) {
	prc, err := FindProcessor(source, defaultName, store)
	if err != nil {
		return nil, nil, err
	}

	prcType, err := store.ProcessorType(prc.Spec.Type)
	if err == nil && prcType == nil {
		err = fmt.Errorf("unknown %s: %s", KindProcessorType, prc.Spec.Type)
	}
	if err != nil {
		return prc, nil, err
	}

	return prc, prcType, nil
}

func findDestinationAndType(destination *ResourceConfiguration, defaultName string, store ResourceStore) (*Destination, *DestinationType, error) {
	dest, err := FindDestination(destination, defaultName, store)
	if err != nil {
		return nil, nil, err
	}

	destType, err := store.DestinationType(dest.Spec.Type)
	if err == nil && destType == nil {
		err = fmt.Errorf("unknown %s: %s", KindDestinationType, dest.Spec.Type)
	}
	if err != nil {
		return dest, nil, err
	}

	return dest, destType, nil
}

func findResourceAndType(resourceKind Kind, resource *ResourceConfiguration, defaultName string, store ResourceStore) (Resource, *ResourceType, error) {
	switch resourceKind {
	case KindSource:
		src, srcType, err := findSourceAndType(resource, defaultName, store)
		if srcType == nil {
			return src, nil, err
		}
		return src, &srcType.ResourceType, err
	case KindProcessor:
		prc, prcType, err := findProcessorAndType(resource, defaultName, store)
		if prcType == nil {
			return prc, nil, err
		}
		return prc, &prcType.ResourceType, err
	case KindDestination:
		dest, destType, err := findDestinationAndType(resource, defaultName, store)
		if destType == nil {
			return dest, nil, err
		}
		return dest, &destType.ResourceType, err
	}
	return nil, nil, nil
}

// ----------------------------------------------------------------------

func (cs *ConfigurationSpec) validate(errors validation.Errors) {
	cs.validateSpecFields(errors)
	cs.validateRaw(errors)
	cs.Selector.validate(errors)
}

func (cs *ConfigurationSpec) validateSpecFields(errors validation.Errors) {
	if cs.Raw != "" {
		if len(cs.Destinations) > 0 || len(cs.Sources) > 0 {
			errors.Add(fmt.Errorf("configuration must specify raw or sources and destinations"))
		}
	}
}

func (cs *ConfigurationSpec) validateRaw(errors validation.Errors) {
	if cs.Raw == "" {
		return
	}
	parsed := map[string]any{}
	err := yaml.Unmarshal([]byte(cs.Raw), parsed)
	if err != nil {
		errors.Add(fmt.Errorf("unable to parse spec.raw as yaml: %w", err))
	}
}

func (cs *ConfigurationSpec) validateSourcesAndDestinations(errors validation.Errors, store ResourceStore) {
	for _, source := range cs.Sources {
		source.validate(KindSource, errors, store)
	}
	for _, destination := range cs.Destinations {
		destination.validate(KindDestination, errors, store)
	}
}

func (rc *ResourceConfiguration) validate(resourceKind Kind, errors validation.Errors, store ResourceStore) {
	if rc.validateHasNameOrType(resourceKind, errors) {
		rc.validateParameters(resourceKind, errors, store)
	}
	rc.validateProcessors(resourceKind, errors, store)
}

func (rc *ResourceConfiguration) validateHasNameOrType(resourceKind Kind, errors validation.Errors) bool {
	// must have name or type
	if rc.Name == "" && rc.Type == "" {
		errors.Add(fmt.Errorf("all %s must have either a name or type", resourceKind))
		return false
	}
	return true
}

func (rc *ResourceConfiguration) validateParameters(resourceKind Kind, errors validation.Errors, store ResourceStore) {
	// must have a name
	for _, parameter := range rc.Parameters {
		if parameter.Name == "" {
			errors.Add(fmt.Errorf("all %s parameters must have a name", resourceKind))
		}
	}
	_, resourceType, err := findResourceAndType(resourceKind, rc, string(resourceKind), store)
	if err != nil {
		errors.Add(err)
		return
	}
	// ensure parameters are valid
	for _, parameter := range rc.Parameters {
		if parameter.Name == "" {
			continue
		}
		def := resourceType.Spec.ParameterDefinition(parameter.Name)
		if def == nil {
			errors.Add(fmt.Errorf("parameter %s not defined in type %s", parameter.Name, resourceType.Name()))
			continue
		}
		err := def.validateValue(parameter.Value)
		if err != nil {
			errors.Add(err)
		}
	}
}

func (rc *ResourceConfiguration) validateProcessors(resourceKind Kind, errors validation.Errors, store ResourceStore) {
	for _, processor := range rc.Processors {
		processor.validate(KindProcessor, errors, store)
	}
}

// ----------------------------------------------------------------------
// Printable

// PrintableFieldTitles returns the list of field titles, used for printing a table of resources
func (c *Configuration) PrintableFieldTitles() []string {
	return []string{"Name", "Match"}
}

// PrintableFieldValue returns the field value for a title, used for printing a table of resources
func (c *Configuration) PrintableFieldValue(title string) string {
	switch title {
	case "Name":
		return c.Name()
	case "Match":
		return c.AgentSelector().String()
	default:
		return "-"
	}
}

// ----------------------------------------------------------------------
// Indexed

// IndexFields returns a map of field name to field value to be stored in the index
func (c *Configuration) IndexFields(index search.Indexer) {
	c.ResourceMeta.IndexFields(index)

	// add the type of configuration
	index("type", string(c.Type()))

	// add source, sourceType fields
	for _, source := range c.Spec.Sources {
		source.indexFields("source", "sourceType", index)
	}

	// add destination, destinationType fields
	for _, destination := range c.Spec.Destinations {
		destination.indexFields("destination", "destinationType", index)
	}

	// add pipeline fields
	//
	// TODO(andy): I was going to add pipeline:traces, pipeline:logs, and pipeline:metrics because I thought it would be a
	// useful way to filter configurations. However, we need a ResourceStore implementation to call otelConfiguration and
	// we don't have that here, even though indexing is actually done in the store. I think the best solution is to cache
	// the output on the Spec and keep that up to date as any dependent sourceTypes and destinationTypes change. This will
	// improve performance when comparing configurations and displaying the configuration in UI.
}

func (rc *ResourceConfiguration) indexFields(resourceName string, resourceTypeName string, index search.Indexer) {
	index(resourceName, rc.Name)
	index(resourceTypeName, rc.Type)
}
