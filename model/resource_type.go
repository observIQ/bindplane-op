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
	"bytes"
	"fmt"
	"io"
	"text/template"

	"github.com/observiq/bindplane-op/model/otel"
	"github.com/observiq/bindplane-op/model/validation"
	"gopkg.in/yaml.v3"
)

// ResourceType is a resource that describes a type of resource including parameters for creating that resource and a
// template for formatting the resource configuration.
//
// There will be separate ResourceTypes for each type of resource, e.g. SourceType for Source resources.
type ResourceType struct {
	ResourceMeta `yaml:",inline" json:",inline" mapstructure:",squash"`
	Spec         ResourceTypeSpec `json:"spec" yaml:"spec" mapstructure:"spec"`
}

// ResourceTypeSpec is the spec for a resourceType to
type ResourceTypeSpec struct {
	Version string `json:"version" yaml:"version" mapstructure:"version"`

	// Parameters currently uses the model from stanza. Eventually we will probably create a separate definition for
	// BindPlane.
	Parameters         []ParameterDefinition `json:"parameters"  yaml:"parameters"  mapstructure:"parameters"`
	SupportedPlatforms []string              `json:"supportedPlatforms" yaml:"supportedPlatforms" mapstructure:"supportedPlatforms"`

	// individual
	Logs    ResourceTypeOutput `json:"logs,omitempty"    yaml:"logs,omitempty"    mapstructure:"logs"`
	Metrics ResourceTypeOutput `json:"metrics,omitempty" yaml:"metrics,omitempty" mapstructure:"metrics"`
	Traces  ResourceTypeOutput `json:"traces,omitempty"  yaml:"traces,omitempty"  mapstructure:"traces"`

	// pairs (alphabetical order)
	LogsMetrics   ResourceTypeOutput `json:"logs+metrics,omitempty"   yaml:"logs+metrics,omitempty"   mapstructure:"logs+metrics"`
	LogsTraces    ResourceTypeOutput `json:"logs+traces,omitempty"    yaml:"logs+traces,omitempty"    mapstructure:"logs+traces"`
	MetricsTraces ResourceTypeOutput `json:"metrics+traces,omitempty" yaml:"metrics+traces,omitempty" mapstructure:"metrics+traces"`

	// all three (alphabetical order)
	LogsMetricsTraces ResourceTypeOutput `json:"logs+metrics+traces,omitempty" yaml:"logs+metrics+traces,omitempty" mapstructure:"logs+metrics+traces"`
}

// ResourceTypeOutput describes the output of the resource type
type ResourceTypeOutput struct {
	Receivers  ResourceTypeTemplate `json:"receivers,omitempty"  yaml:"receivers,omitempty"  mapstructure:"receivers"`
	Processors ResourceTypeTemplate `json:"processors,omitempty" yaml:"processors,omitempty" mapstructure:"processors"`
	Exporters  ResourceTypeTemplate `json:"exporters,omitempty"  yaml:"exporters,omitempty"  mapstructure:"exporters"`
	Extensions ResourceTypeTemplate `json:"extensions,omitempty" yaml:"extensions,omitempty" mapstructure:"extensions"`
}

// Empty returns true if Receivers, Processors, Exporters, and Extensions are the zero value ""
func (s ResourceTypeOutput) Empty() bool {
	return s.Receivers == "" && s.Processors == "" && s.Exporters == "" && s.Extensions == ""
}

// ResourceTypeTemplate is a go-template that evaluates to an array of OpenTelemetry resources
type ResourceTypeTemplate string

// TemplateErrorHandler handles errors during template evaluation. Typically these will be logged but they could be
// accumulated and reported to the user.
type TemplateErrorHandler func(error)

// ParameterDefinition returns the ParameterDefinition with the specified name or nil if no such parameter exists
func (s *ResourceTypeSpec) ParameterDefinition(name string) *ParameterDefinition {
	for _, p := range s.Parameters {
		if name == p.Name {
			return &p
		}
	}
	return nil
}

// ----------------------------------------------------------------------

// eval executes all of the templates associated with this resource type, returning a partial configuration for each
// telemetry type.
func (rt *ResourceType) eval(resource parameterizedResource, errorHandler TemplateErrorHandler) otel.Partials {
	result := otel.Partials{
		otel.Logs:    rt.evalOutput(&rt.Spec.Logs, resource, errorHandler),
		otel.Metrics: rt.evalOutput(&rt.Spec.Metrics, resource, errorHandler),
		otel.Traces:  rt.evalOutput(&rt.Spec.Traces, resource, errorHandler),
	}

	// add multi-pipelines components
	logsMetrics := rt.evalOutput(&rt.Spec.LogsMetrics, resource, errorHandler)
	result[otel.Logs].Add(logsMetrics)
	result[otel.Metrics].Add(logsMetrics)

	logsTraces := rt.evalOutput(&rt.Spec.LogsTraces, resource, errorHandler)
	result[otel.Logs].Add(logsTraces)
	result[otel.Traces].Add(logsTraces)

	metricsTraces := rt.evalOutput(&rt.Spec.MetricsTraces, resource, errorHandler)
	result[otel.Metrics].Add(metricsTraces)
	result[otel.Traces].Add(metricsTraces)

	logsMetricsTraces := rt.evalOutput(&rt.Spec.LogsMetricsTraces, resource, errorHandler)
	result[otel.Logs].Add(logsMetricsTraces)
	result[otel.Metrics].Add(logsMetricsTraces)
	result[otel.Traces].Add(logsMetricsTraces)

	return result
}

// evalOutput executes the templates associated with the specified output using the specified resource and errorHandler.
func (rt *ResourceType) evalOutput(output *ResourceTypeOutput, resource parameterizedResource, errorHandler TemplateErrorHandler) *otel.Partial {
	params := map[string]any{}
	// start with default parameters
	for _, p := range rt.Spec.Parameters {
		if p.Default != nil {
			params[p.Name] = p.Default
		}
	}
	// resource can overrides the parameters
	for _, p := range resource.ResourceParameters() {
		params[p.Name] = p.Value
	}
	// eval all of the components
	return &otel.Partial{
		Receivers:  rt.evalTemplate(output.Receivers, resource, params, errorHandler),
		Processors: rt.evalTemplate(output.Processors, resource, params, errorHandler),
		Exporters:  rt.evalTemplate(output.Exporters, resource, params, errorHandler),
		Extensions: rt.evalTemplate(output.Extensions, resource, params, errorHandler),
	}
}

// evalTemplate evaluates a single template with the specified paramValues. nameProvider is available to make the name
// unique and the errorHandler will accumulate errors so that they can be reported once.
func (rt *ResourceType) evalTemplate(r ResourceTypeTemplate, nameProvider otel.ComponentIDProvider, paramValues map[string]any, errorHandler TemplateErrorHandler) otel.ComponentList {
	set := otel.ComponentList{}

	// get the template for the key
	t, err := template.New(rt.Name()).Option("missingkey=error").Parse(string(r))
	if err != nil {
		errorHandler(err)
		return set
	}

	// render the template
	var writer bytes.Buffer
	if err := t.Execute(&writer, paramValues); err != nil {
		errorHandler(err)
		return set
	}

	bytes := writer.Bytes()

	// parse as yaml so that we can combine yaml fragments and render
	var parsed []map[string]any
	if err := yaml.Unmarshal(bytes, &parsed); err != nil {
		errorHandler(err)
		return set
	}

	// assemble all of the blocks after renaming them
	for _, block := range parsed {
		for key, value := range block {
			componentID := nameProvider.ComponentID(key)
			set = append(set, map[otel.ComponentID]any{
				componentID: value,
			})
		}
	}

	return set
}

// ----------------------------------------------------------------------

// PrintableFieldTitles returns the list of field titles, used for printing a table of resources
func (rt *ResourceType) PrintableFieldTitles() []string {
	return []string{"Name", "Display"}
}

// ----------------------------------------------------------------------
// validation

// Validate returns an error if any part of the ResourceType is invalid
func (rt *ResourceType) Validate() error {
	errs := validation.NewErrors()

	rt.ResourceMeta.validate(errs)
	rt.Spec.validate(errs)

	return errs.Result()
}

// ValidateWithStore returns an error if any part of the ResourceType is invalid
func (rt *ResourceType) ValidateWithStore(store ResourceStore) error {
	return rt.Validate()
}

func (s *ResourceTypeSpec) validate(errs validation.Errors) {
	s.validateParameterDefinitions(errs)

	// assemble default parameter values for validation
	params := map[string]any{}
	for _, p := range s.Parameters {
		if p.Default != nil {
			params[p.Name] = p.Default
		} else {
			// for template validation, just provide a reasonable default based on the type
			switch p.Type {
			case stringType:
				params[p.Name] = ""
			case boolType:
				params[p.Name] = false
			case intType:
				params[p.Name] = 0
			case stringsType:
				params[p.Name] = []string{}
			case enumType:
				params[p.Name] = "" // p.ValidValues[0] // cannot guarantee this is valid and "" is fine
			}
		}
	}

	s.Logs.validateTemplates(errs, "logs", params)
	s.Metrics.validateTemplates(errs, "metrics", params)
	s.Traces.validateTemplates(errs, "traces", params)
}

func (s *ResourceTypeSpec) validateParameterDefinitions(errs validation.Errors) {
	for _, parameter := range s.Parameters {
		parameter.validateDefinition(errs)
		s.validateParameterRelevantIf(parameter, errs)
	}
}

// validateParameterRelevantIf in ResourceTypeSpec because we need to check against other parameter names
func (s *ResourceTypeSpec) validateParameterRelevantIf(parameter ParameterDefinition, errs validation.Errors) {
	for _, relevantIf := range parameter.RelevantIf {
		if relevantIf.Name == "" {
			errs.Add(fmt.Errorf("relevantIf for '%s' must have a name", parameter.Name))
			continue
		}
		ref := s.ParameterDefinition(relevantIf.Name)
		if ref == nil {
			errs.Add(fmt.Errorf("relevantIf for '%s' refers to nonexistant parameter '%s'", parameter.Name, relevantIf.Name))
			continue
		}
		if relevantIf.Operator == "" {
			errs.Add(fmt.Errorf("relevantIf '%s' for '%s' must have an operator", ref.Name, parameter.Name))
		}
		if relevantIf.Value == nil {
			errs.Add(fmt.Errorf("relevantIf '%s' for '%s' must have a value", ref.Name, parameter.Name))
			continue
		}
		err := ref.validateValueType(parameterFieldRelevantIf, relevantIf.Value)
		if err != nil {
			errs.Add(fmt.Errorf("relevantIf '%s' for '%s': %w", ref.Name, parameter.Name, err))
		}
	}
}

func (s ResourceTypeOutput) validateTemplates(errs validation.Errors, name string, params map[string]any) {
	s.Receivers.validate(errs, fmt.Sprintf("%s.receivers", name), params)
	s.Processors.validate(errs, fmt.Sprintf("%s.processors", name), params)
	s.Exporters.validate(errs, fmt.Sprintf("%s.exporters", name), params)
	s.Extensions.validate(errs, fmt.Sprintf("%s.extensions", name), params)
}

func (s ResourceTypeTemplate) validate(errs validation.Errors, name string, params map[string]any) {
	if s == "" {
		// no validation for empty templates
		return
	}
	// ensure the template is valid
	t, err := template.New(name).Option("missingkey=error").Parse(string(s))
	if err != nil {
		errs.Add(err)
		return
	}
	// ensure that it can be executed with default values
	if err := t.Execute(io.Discard, params); err != nil {
		errs.Add(err)
	}
}

// TelemetryTypes returns the supported telemetry types (logs, metrics, or traces).
// Only applicable to SourceTypes.
func (s *ResourceTypeSpec) TelemetryTypes() []otel.PipelineType {
	telemetryTypes := make([]otel.PipelineType, 0, 3)

	if !s.Logs.Empty() || !s.LogsMetrics.Empty() || !s.LogsTraces.Empty() || !s.LogsMetricsTraces.Empty() {
		telemetryTypes = append(telemetryTypes, otel.Logs)
	}

	if !s.Metrics.Empty() || !s.LogsMetrics.Empty() || !s.MetricsTraces.Empty() || !s.LogsMetricsTraces.Empty() {
		telemetryTypes = append(telemetryTypes, otel.Metrics)
	}

	if !s.Traces.Empty() || !s.LogsTraces.Empty() || !s.MetricsTraces.Empty() || !s.LogsMetricsTraces.Empty() {
		telemetryTypes = append(telemetryTypes, otel.Traces)
	}

	return telemetryTypes
}
