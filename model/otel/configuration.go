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

package otel

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

// PipelineType is the telemetry type specified in pipeline names, e.g. metrics in metrics/redis.
type PipelineType string

// OpenTelemetry currently supports "metrics", "logs", and "traces"
const (
	Metrics PipelineType = "metrics"
	Logs    PipelineType = "logs"
	Traces  PipelineType = "traces"
)

// ComponentID is a the name of an individual receiver, processor, exporter, or extension.
type ComponentID string

// ComponentMap is a map of individual receivers, processors, etc.
type ComponentMap map[ComponentID]any

// ComponentList is an ordered list of individual receivers, processors, etc. The order is especially important for
// processors.
type ComponentList []map[ComponentID]any

// NewComponentID creates a new ComponentID from its type and name
func NewComponentID(pipelineType, name string) ComponentID {
	return ComponentID(fmtComponentID(pipelineType, name))
}

func fmtComponentID(pipelineType, name string) string {
	if name == "" {
		return pipelineType
	}
	return fmt.Sprintf("%s/%s", pipelineType, name)
}

// ParseComponentID returns the typeName and name from a ComponentID
func ParseComponentID(id ComponentID) (pipelineType, name string) {
	parts := strings.SplitN(string(id), "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// Configuration is a rough approximation of an OpenTelemetry configuration. It is used to help assemble a configuration
// and marshal it to a string to send to an agent.
type Configuration struct {
	Receivers  ComponentMap `yaml:"receivers,omitempty"`
	Processors ComponentMap `yaml:"processors,omitempty"`
	Exporters  ComponentMap `yaml:"exporters,omitempty"`
	Extensions ComponentMap `yaml:"extensions,omitempty"`
	Service    Service      `yaml:"service"`
}

// NewConfiguration creates a new configuration with initialized fields
func NewConfiguration() *Configuration {
	return &Configuration{
		Receivers:  ComponentMap{},
		Processors: ComponentMap{},
		Exporters:  ComponentMap{},
		Extensions: ComponentMap{},
		Service: Service{
			Pipelines: Pipelines{},
		},
	}
}

// YAML marshals the configuration to yaml
func (c *Configuration) YAML() (string, error) {
	if c == nil || !c.HasPipelines() {
		return NoopConfig, nil
	}
	bytes, err := yaml.Marshal(c)
	return string(bytes), err
}

// HasPipelines returns true if there are pipelines
func (c *Configuration) HasPipelines() bool {
	return len(c.Service.Pipelines) > 0
}

// Service is the part of the configuration that defines the pipelines which consist of references to the components in
// the Configuration.
type Service struct {
	Extensions []ComponentID `yaml:"extensions,omitempty"`
	Pipelines  Pipelines     `yaml:"pipelines"`
}

// Pipelines are identified by a pipeline type and name in the form type/name where type is "metrics", "logs", or
// "traces"
type Pipelines map[string]Pipeline

// Pipeline is an ordered list of receivers, processors, and exporters.
type Pipeline struct {
	Name       string        `yaml:"-"`
	Receivers  []ComponentID `yaml:"receivers"`
	Processors []ComponentID `yaml:"processors"`
	Exporters  []ComponentID `yaml:"exporters"`
}

// Incomplete returns true if there are zero Receivers or zero Exporters
func (p *Pipeline) Incomplete() bool {
	return len(p.Receivers) == 0 || len(p.Exporters) == 0
}

// Partial represents a fragment of configuration produced by an individual resource.
type Partial struct {
	Receivers  ComponentList
	Processors ComponentList
	Exporters  ComponentList
	Extensions ComponentList
}

// Size returns the number of components in the partial configuration
func (p *Partial) Size() int {
	return len(p.Receivers) + len(p.Processors) + len(p.Exporters) + len(p.Extensions)
}

// Add adds components from another partial by appending each of the component lists together
func (p *Partial) Add(o *Partial) {
	p.Receivers = append(p.Receivers, o.Receivers...)
	p.Processors = append(p.Processors, o.Processors...)
	p.Exporters = append(p.Exporters, o.Exporters...)
	p.Extensions = append(p.Extensions, o.Extensions...)
}

// Partials represents a fragments of configuration for each type of telemetry.
type Partials map[PipelineType]*Partial

// Add combines the individual Logs, Metrics, and Traces Partial configurations
func (p Partials) Add(o Partials) {
	p[Logs].Add(o[Logs])
	p[Metrics].Add(o[Metrics])
	p[Traces].Add(o[Traces])
}

// ComponentIDProvider can provide ComponentIDs for component names
type ComponentIDProvider interface {
	ComponentID(componentName string) ComponentID
}

// UniqueComponentID ensures that each ComponentID is unique by including the type and resource name. To make them easy
// to find in a completed configuration, we preserve the part before the / and then insert the type and resource name
// separated by 2 underscores.
func UniqueComponentID(original, typeName, resourceName string) ComponentID {
	// replace type/name with type/resourceType__resourceName__name
	pipelineType, name := ParseComponentID(ComponentID(original))

	var newName string
	if name != "" {
		newName = fmt.Sprintf("%s__%s__%s", typeName, resourceName, name)
	} else {
		newName = fmt.Sprintf("%s__%s", typeName, resourceName)
	}
	return NewComponentID(pipelineType, newName)
}

// AddExtensions adds all of the extensions to the configuration, replacing any extensions with the same id
func (c *Configuration) AddExtensions(extensions ComponentList) {
	for _, extension := range extensions {
		for n, v := range extension {
			c.AddExtension(n, v)
		}
	}
}

// AddExtension adds the specified extension with the specified id, replace any extension with the same id
func (c *Configuration) AddExtension(name ComponentID, extension any) {
	c.Extensions[name] = extension
	if !slices.Contains(c.Service.Extensions, name) {
		c.Service.Extensions = append(c.Service.Extensions, name)
	}
}

// AddPipeline adds a pipeline and all of the corresponding components to the configuration
func (c *Configuration) AddPipeline(name string, pipelineType PipelineType, source, destination Partials) {
	s := source[pipelineType]
	d := destination[pipelineType]
	if s.Size() == 0 || d.Size() == 0 {
		// not all pipelineType will have components, ignore these
		return
	}

	p := Pipeline{}

	// add any receivers specified
	p.AddReceivers(c.Receivers.addComponents(s.Receivers))
	p.AddReceivers(c.Receivers.addComponents(d.Receivers))

	// add any processors specified
	p.AddProcessors(c.Processors.addComponents(s.Processors))
	p.AddProcessors(c.Processors.addComponents(d.Processors))

	// add any exporters specified
	p.AddExporters(c.Exporters.addComponents(s.Exporters))
	p.AddExporters(c.Exporters.addComponents(d.Exporters))

	// skip any incomplete pipelines
	if p.Incomplete() {
		return
	}

	// any extensions are added and shared by all pipelines
	c.AddExtensions(s.Extensions)
	c.AddExtensions(d.Extensions)

	pipelineID := fmt.Sprintf("%s/%s", pipelineType, name)
	c.Service.Pipelines[pipelineID] = p
}

// addComponents adds the components to the map and returns their ids as a convenience to build the pipeline
func (c ComponentMap) addComponents(componentList ComponentList) []ComponentID {
	ids := []ComponentID{}
	for _, components := range componentList {
		for id, component := range components {
			c[id] = component
			ids = append(ids, id)
		}
	}
	return ids
}

// AddReceivers adds receivers to the pipeline
func (p *Pipeline) AddReceivers(id []ComponentID) {
	p.Receivers = append(p.Receivers, id...)
}

// AddProcessors adds processors to the pipeline
func (p *Pipeline) AddProcessors(id []ComponentID) {
	p.Processors = append(p.Processors, id...)
}

// AddExporters adds exporters to the pipeline
func (p *Pipeline) AddExporters(id []ComponentID) {
	p.Exporters = append(p.Exporters, id...)
}
