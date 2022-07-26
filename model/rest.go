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

// AgentResponse is the REST API response to GET /v1/agent/:name
type AgentResponse struct {
	Agent *Agent `json:"agent"`
}

// AgentsResponse is the REST API response to GET /v1/agents endpoint.
type AgentsResponse struct {
	Agents []*Agent `json:"agents"`
}

// DeleteAgentsPayload is the REST API body to DELETE /v1/agents
type DeleteAgentsPayload struct {
	IDs []string `json:"ids"`
}

// DeleteAgentsResponse is the REST API response to DELETE /v1/agents
type DeleteAgentsResponse = AgentsResponse

// AgentLabelsResponse is the REST API response to GET /v1/agents/{id}/labels
type AgentLabelsResponse struct {
	Errors []string `json:"errors"`
	Labels *Labels  `json:"labels"`
}

// AgentLabelsPayload is the REST API body for PATCH /v1/agents/{id}/labels
type AgentLabelsPayload struct {
	Labels map[string]string `json:"labels"`
}

// BulkAgentLabelsPayload is the REST API body for PATCH /v1/agents/labels
type BulkAgentLabelsPayload struct {
	IDs       []string          `json:"ids"`
	Labels    map[string]string `json:"labels"`
	Overwrite bool              `json:"overwrite"`
}

// BulkAgentLabelsResponse is the REST API response to PATCH /v1/agents/labels
type BulkAgentLabelsResponse struct {
	Errors []string `json:"errors"`
}

// ConfigurationsResponse is the REST API response to GET /v1/configurations
type ConfigurationsResponse struct {
	Configurations []*Configuration `json:"configurations"`
}

// ConfigurationResponse is the REST API response to GET /v1/configuration/:name
type ConfigurationResponse struct {
	Configuration *Configuration `json:"configuration"`
	Raw           string         `json:"raw"`
}

// SourcesResponse is the REST API response to GET /v1/sources
type SourcesResponse struct {
	Sources []*Source `json:"sources"`
}

// SourceResponse is the REST API response to GET /v1/sources/:name
type SourceResponse struct {
	Source *Source `json:"source"`
}

// SourceTypesResponse is the REST API response to GET /v1/sourceTypes
type SourceTypesResponse struct {
	SourceTypes []*SourceType `json:"sourceTypes"`
}

// SourceTypeResponse is the REST API response to GET /v1/sourceType/:name
type SourceTypeResponse struct {
	SourceType *SourceType `json:"sourceType"`
}

// ProcessorsResponse is the REST API response to GET /v1/processors
type ProcessorsResponse struct {
	Processors []*Processor `json:"processors"`
}

// ProcessorResponse is the REST API response to GET /v1/processors/:name
type ProcessorResponse struct {
	Processor *Processor `json:"processor"`
}

// ProcessorTypesResponse is the REST API response to GET /v1/processorTypes
type ProcessorTypesResponse struct {
	ProcessorTypes []*ProcessorType `json:"processorTypes"`
}

// ProcessorTypeResponse is the REST API response to GET /v1/processorType/:name
type ProcessorTypeResponse struct {
	ProcessorType *ProcessorType `json:"processorType"`
}

// DestinationsResponse is the REST API response to GET /v1/destinations
type DestinationsResponse struct {
	Destinations []*Destination `json:"destinations"`
}

// DestinationResponse is the REST API response to GET /v1/destinations/:name
type DestinationResponse struct {
	Destination *Destination `json:"destination"`
}

// DestinationTypesResponse is the REST API response to GET /v1/destinationTypes
type DestinationTypesResponse struct {
	DestinationTypes []*DestinationType `json:"destinationTypes"`
}

// DestinationTypeResponse is the REST API response to GET /v1/destinationType/:name
type DestinationTypeResponse struct {
	DestinationType *DestinationType `json:"destinationType"`
}

// ApplyResponse is the REST API response to POST /v1/apply.  This is used on
// the server side to return updates consisting of generic ResourceStatuses.
type ApplyResponse struct {
	Updates []ResourceStatus `json:"updates"`
}

// ApplyResponseClientSide is the REST API Response.  This is used on the client
// side where updates consists of AnyResourceStatuses.
type ApplyResponseClientSide struct {
	Updates []*AnyResourceStatus `json:"updates"`
}

// ApplyPayload is the REST API body for POST /v1/apply
type ApplyPayload struct {
	Resources []*AnyResource `json:"resources"`
}

// DeletePayload is the REST API body for POST /v1/delete.  Though resources
// with full Spec can be included its only necessary for the Kind and Metadata.Name
// fields to be present.
type DeletePayload struct {
	Resources []*AnyResource `json:"resources"`
}

// DeleteResponse is the REST API response to POST /v1/delete
type DeleteResponse struct {
	Errors  []string         `json:"errors"`
	Updates []ResourceStatus `json:"updates"`
}

// DeleteResponseClientSide is the REST API response to POST /v1/delete
type DeleteResponseClientSide struct {
	Errors  []string             `json:"errors"`
	Updates []*AnyResourceStatus `json:"updates"`
}

// InstallCommandResponse is the REST API response to GET /v1/agent-versions/{version}/install-command
type InstallCommandResponse struct {
	Command string `json:"command"`
}

// PostAgentVersionRequest is the REST API body for POST /v1/agents/{id}/version
type PostAgentVersionRequest struct {
	Version string `json:"version"`
}

// PutDuplicateConfigRequest is the REST API body for PUT /v1/configurations/{name}/duplicate
type PutDuplicateConfigRequest struct {
	// The intended name of the duplicated config
	Name string `json:"name"`
}

// PutDuplicateConfigResponse is the REST API response to PUT /v1/configurations/{name}/duplicate
type PutDuplicateConfigResponse = PutDuplicateConfigRequest
