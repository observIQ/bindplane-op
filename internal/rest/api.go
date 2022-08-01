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

package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/internal/server"
	"github.com/observiq/bindplane-op/internal/store"
	"github.com/observiq/bindplane-op/internal/store/search"
	"github.com/observiq/bindplane-op/internal/version"
	"github.com/observiq/bindplane-op/model"
)

var tracer = otel.Tracer("rest")

// AddRestRoutes adds all API routes to the gin HTTP router
func AddRestRoutes(router gin.IRouter, bindplane server.BindPlane) {
	router.GET("/agents", func(c *gin.Context) { agents(c, bindplane) })
	router.GET("/agents/:id", func(c *gin.Context) { getAgent(c, bindplane) })
	router.DELETE("/agents", func(c *gin.Context) { deleteAgents(c, bindplane) })
	router.PATCH("/agents/labels", func(c *gin.Context) { labelAgents(c, bindplane) })
	router.GET("/agents/:id/labels", func(c *gin.Context) { getAgentLabels(c, bindplane) })
	router.PATCH("/agents/:id/labels", func(c *gin.Context) { patchAgentLabels(c, bindplane) })
	router.PUT("/agents/:id/restart", func(c *gin.Context) { restartAgent(c, bindplane) })
	router.POST("/agents/:id/version", func(c *gin.Context) { updateAgent(c, bindplane) })
	router.GET("/agents/:id/configuration", func(c *gin.Context) { getAgentConfiguration(c, bindplane) })

	router.GET("/configurations", func(c *gin.Context) { configurations(c, bindplane) })
	router.GET("/configurations/:name", func(c *gin.Context) { configuration(c, bindplane) })
	router.DELETE("/configurations/:name", func(c *gin.Context) { deleteConfiguration(c, bindplane) })
	router.POST("/configurations/:name/duplicate", func(c *gin.Context) { duplicateConfig(c, bindplane) })

	router.GET("/sources", func(c *gin.Context) { sources(c, bindplane) })
	router.GET("/sources/:name", func(c *gin.Context) { source(c, bindplane) })
	router.DELETE("/sources/:name", func(c *gin.Context) { deleteSource(c, bindplane) })

	router.GET("/source-types", func(c *gin.Context) { sourceTypes(c, bindplane) })
	router.GET("/source-types/:name", func(c *gin.Context) { sourceType(c, bindplane) })
	router.DELETE("/source-types/:name", func(c *gin.Context) { deleteSourceType(c, bindplane) })

	router.GET("/processors", func(c *gin.Context) { processors(c, bindplane) })
	router.GET("/processors/:name", func(c *gin.Context) { processor(c, bindplane) })
	router.DELETE("/processors/:name", func(c *gin.Context) { deleteProcessor(c, bindplane) })

	router.GET("/processor-types", func(c *gin.Context) { processorTypes(c, bindplane) })
	router.GET("/processor-types/:name", func(c *gin.Context) { processorType(c, bindplane) })
	router.DELETE("/processor-types/:name", func(c *gin.Context) { deleteProcessorType(c, bindplane) })

	router.GET("/destinations", func(c *gin.Context) { destinations(c, bindplane) })
	router.GET("/destinations/:name", func(c *gin.Context) { destination(c, bindplane) })
	router.DELETE("/destinations/:name", func(c *gin.Context) { deleteDestination(c, bindplane) })

	router.GET("/destination-types", func(c *gin.Context) { destinationTypes(c, bindplane) })
	router.GET("/destination-types/:name", func(c *gin.Context) { destinationType(c, bindplane) })
	router.DELETE("/destination-types/:name", func(c *gin.Context) { deleteDestinationType(c, bindplane) })

	router.POST("/apply", func(c *gin.Context) { applyResources(c, bindplane) })
	router.POST("/delete", func(c *gin.Context) { deleteResources(c, bindplane) })

	router.GET("/version", func(c *gin.Context) { bindplaneVersion(c) })
	router.GET("/agent-versions/:version/install-command", func(c *gin.Context) { getInstallCommand(c, bindplane) })
}

// @Summary List agents
// @Produce json
// @Router /agents [get]
// @Success 200 {object} model.AgentsResponse
// @Failure 500 {object} ErrorResponse
func agents(c *gin.Context, bindplane server.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/agents")
	defer span.End()

	options := []store.QueryOption{}

	selectorString := c.DefaultQuery("selector", "")
	selector, err := model.SelectorFromString(selectorString)
	if err != nil {
		handleErrorResponse(c, http.StatusBadRequest, err)
		return
	}
	options = append(options, store.WithSelector(selector))

	query := c.DefaultQuery("query", "")
	if query != "" {
		q := search.ParseQuery(query)
		q.ReplaceVersionLatest(bindplane.Versions())
		options = append(options, store.WithQuery(q))
	}

	offset := c.DefaultQuery("offset", "0")
	offsetValue, err := strconv.Atoi(offset)
	if err != nil {
		handleErrorResponse(c, http.StatusBadRequest, fmt.Errorf("offset must be a number: %v", err))
		return
	}
	options = append(options, store.WithOffset(offsetValue))

	limit := c.DefaultQuery("limit", "0")
	limitValue, err := strconv.Atoi(limit)
	if err != nil {
		handleErrorResponse(c, http.StatusBadRequest, fmt.Errorf("limit must be a number: %v", err))
		return
	}
	options = append(options, store.WithLimit(limitValue))

	sort := c.DefaultQuery("sort", "")
	if sort != "" {
		options = append(options, store.WithSort(sort))
	}

	agents, err := bindplane.Store().Agents(ctx, options...)
	if err != nil {
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, model.AgentsResponse{
		Agents: agents,
	})
}

// @Summary delete agents by ids
// @Produce json
// @Router /agents [delete]
// @Param 	id	body	[]string	true "list of agent ids to delete"
// @Success 200 {object} model.DeleteAgentsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func deleteAgents(c *gin.Context, bindplane server.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/agents")
	defer span.End()

	p := &model.DeleteAgentsPayload{}

	if err := c.BindJSON(p); err != nil {
		handleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	deleted, err := bindplane.Store().DeleteAgents(ctx, p.IDs)
	if err != nil {
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &model.DeleteAgentsResponse{
		Agents: deleted,
	})
}

// @Summary Get agent by id
// @Produce json
// @Router /agents/{id} [get]
// @Param 	id	path	string	true "the id of the agent"
// @Success 200 {object} model.AgentResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func getAgent(c *gin.Context, bindplane server.BindPlane) {
	id := c.Param("id")

	agent, err := bindplane.Store().Agent(id)

	switch {
	case err != nil:
		handleErrorResponse(c, http.StatusInternalServerError, err)
	case agent == nil:
		handleErrorResponse(c, http.StatusNotFound, store.ErrResourceMissing)
	default:
		c.JSON(http.StatusOK, model.AgentResponse{
			Agent: agent,
		})
	}
}

// @Summary Get agent labels by agent id
// @Produce json
// @Router /agents/{id}/labels [get]
// @Param 	id	path	string	true "the id of the agent"
// @Success 200 {object} model.AgentLabelsResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func getAgentLabels(c *gin.Context, bindplane server.BindPlane) {
	id := c.Param("id")

	agent, err := bindplane.Store().Agent(id)

	switch {
	case err != nil:
		handleErrorResponse(c, http.StatusInternalServerError, err)
	case agent == nil:
		handleErrorResponse(c, http.StatusNotFound, store.ErrResourceMissing)
	default:
		c.JSON(http.StatusOK, model.AgentLabelsResponse{
			Labels: &agent.Labels,
		})
	}
}

// @Summary Get configuration for a given agent
// @Produce json
// @Router /agents/{id}/configuration [get]
// @Param 	id	path	string	true "the id of the agent"
// @Success 200 {object} model.ConfigurationResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func getAgentConfiguration(c *gin.Context, bindplane server.BindPlane) {
	id := c.Param("id")

	agent, err := bindplane.Store().Agent(id)
	switch {
	case err != nil:
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	case agent == nil:
		handleErrorResponse(c, http.StatusNotFound, store.ErrResourceMissing)
		return
	}

	config, err := bindplane.Store().AgentConfiguration(id)
	if err != nil {
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &model.ConfigurationResponse{Configuration: config})
}

// @Summary Bulk apply labels to agents
// @Produce json
// @Router /agents/labels [patch]
// @Param ids 	body	[]string	true "agent IDs"
// @Param labels 	body	map[string]string	true "labels to apply"
// @Param labels body boolean false "overwrite labels"
// @Success 200 {object} model.BulkAgentLabelsResponse
func labelAgents(c *gin.Context, bindplane server.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/labelAgents")
	defer span.End()

	p := &model.BulkAgentLabelsPayload{}

	if err := c.BindJSON(p); err != nil {
		handleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if p.Labels == nil {
		handleErrorResponse(c, http.StatusBadRequest, fmt.Errorf("body is missing the required labels field"))
		return
	}

	if p.IDs == nil {
		handleErrorResponse(c, http.StatusBadRequest, fmt.Errorf(("body is missing the required ids field")))
		return
	}

	newLabels, err := model.LabelsFromMap(p.Labels)
	if err != nil {
		handleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Accumulate API errors outside of upsert, and then upsert agents with valid label operations
	// Check to see if 1) agent exists and 2) there are no label conflicts if overwrite=false.
	upsertIDs := make([]string, 0, len(p.IDs))
	apiErrors := make([]string, 0)
	for _, id := range p.IDs {
		curAgent, err := bindplane.Store().Agent(id)

		switch {
		case err != nil:
			handleErrorResponse(c, http.StatusInternalServerError, err)
			apiErrors = append(apiErrors, fmt.Sprintf("failed to apply labels for agent with id %s, %s", id, err.Error()))
			continue
		case curAgent == nil:
			apiErrors = append(apiErrors, fmt.Sprintf("failed to apply labels for agent with id %s, agent not found", id))
			continue
		case !p.Overwrite && curAgent.Labels.Conflicts(newLabels):
			apiErrors = append(apiErrors, fmt.Sprintf("failed to apply labels for agent with id %s, labels conflict, include overwrite: true in body to overwrite", id))
			continue
		}
		// Agent is cleared to patch - add it to upsertIDs
		upsertIDs = append(upsertIDs, id)
	}

	updater := func(current *model.Agent) {
		current.Labels = model.LabelsFromMerge(current.Labels, newLabels)
	}

	bindplane.Logger().Info("bulkApplyAgentLabels", zap.String("payloadLabels", newLabels.String()), zap.Any("ids", p.IDs), zap.Error(err))

	_, err = bindplane.Store().UpsertAgents(ctx, p.IDs, updater)

	if err != nil {
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &model.BulkAgentLabelsResponse{
		Errors: apiErrors,
	})
}

// @Summary Patch agent labels by agent id
// @Produce json
// @Router /agents/{id}/labels [patch]
// @Param 	id	path	string	true "the id of the agent"
// @Param overwrite query string false "if true, overwrite any existing labels with the same names"
// @Param labels 	body	model.AgentLabelsPayload	true "Labels to be merged with existing labels, empty values will delete existing labels"
// @Success 200 {object} model.AgentLabelsResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func patchAgentLabels(c *gin.Context, bindplane server.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/patchAgentLabels")
	defer span.End()

	id := c.Param("id")
	overwrite := c.DefaultQuery("overwrite", "false") == "true"
	p := &model.AgentLabelsPayload{}
	if err := c.BindJSON(p); err != nil {
		handleErrorResponse(c, http.StatusBadRequest, err)
		return
	}
	if p.Labels == nil {
		handleErrorResponse(c, http.StatusBadRequest, fmt.Errorf("body is missing the required labels field"))
		return
	}

	newLabels, err := model.LabelsFromMap(p.Labels)
	if err != nil {
		handleErrorResponse(c, http.StatusBadRequest, err)
	}

	curAgent, err := bindplane.Store().Agent(id)
	switch {
	case err != nil:
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	case curAgent == nil:
		handleErrorResponse(c, http.StatusNotFound, store.ErrResourceMissing)
		return
	case !overwrite && curAgent.Labels.Conflicts(newLabels):
		err := fmt.Errorf("new labels conflict with existing labels, add ?overwrite=true to replace labels")
		c.Error(err)
		c.JSON(http.StatusConflict, model.AgentLabelsResponse{
			Errors: []string{err.Error()},
			Labels: &curAgent.Labels,
		})
		return
	}

	newAgent, err := bindplane.Store().UpsertAgent(ctx, id, func(agent *model.Agent) {
		agent.Labels = model.LabelsFromMerge(agent.Labels, newLabels)
	})

	if err != nil {
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	bindplane.Logger().Info("patchAgentLabels", zap.String("payloadLabels", newLabels.String()), zap.String("newLabels", newAgent.Labels.String()))
	c.JSON(http.StatusOK, model.AgentLabelsResponse{
		Labels: &newAgent.Labels,
	})
}

// @Summary TODO restart agent
// @Produce json
// @Router /agents/{id}/restart [put]
// @Param 	id	path	string	true "the id of the agent"
func restartAgent(c *gin.Context, bindplane server.BindPlane) {
	id := c.Param("id")

	// TODO(andy): Do a restart
	bindplane.Logger().Info("TODO Restart agent", zap.String("id", id))

	c.Status(http.StatusAccepted)
}

// @Summary TODO update agent
// @Produce json
// TODO (dsvanlani): document body params
// @Router /agents/{id}/version [post]
// @Param 	name	path	string	true "the id of the agent"
func updateAgent(c *gin.Context, bindplane server.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/updateAgent")
	defer span.End()

	id := c.Param("id")
	var req model.PostAgentVersionRequest

	if err := c.BindJSON(&req); err != nil {
		handleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// set the pending version
	_, err := bindplane.Store().UpsertAgent(ctx, id, func(current *model.Agent) {
		current.VersionPending = req.Version
	})
	if err != nil {
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary List Configurations
// @Produce json
// @Router /configurations [get]
// @Success 200 {object} model.ConfigurationsResponse
// @Failure 500 {object} ErrorResponse
func configurations(c *gin.Context, bindplane server.BindPlane) {
	configs, err := bindplane.Store().Configurations()
	if err != nil {
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, model.ConfigurationsResponse{
		Configurations: configs,
	})
}

// @Summary Get configuration by name
// @Produce json
// @Router /configurations/{name} [get]
// @Param 	name	path	string	true "the name of the configuration"
// @Success 200 {object} model.ConfigurationResponse
// @Failure 500 {object} ErrorResponse
func configuration(c *gin.Context, bindplane server.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/configuration")
	defer span.End()

	name := c.Param("name")

	config, err := bindplane.Store().Configuration(name)
	if !okResource(c, config == nil, err) {
		return
	}

	raw, err := config.Render(ctx, bindplane.Store())
	if !okResponse(c, err) {
		return
	}

	c.JSON(http.StatusOK, model.ConfigurationResponse{
		Configuration: config,
		Raw:           raw,
	})
}

// @Summary Delete configuration by name
// @Produce json
// @Router /configurations/{name} [delete]
// @Param 	name	path	string	true "the name of the configuration to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func deleteConfiguration(c *gin.Context, bindplane server.BindPlane) {
	name := c.Param("name")
	configuration, err := bindplane.Store().DeleteConfiguration(name)
	if okResource(c, configuration == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// @Summary Duplicate an existing configuration
// @Produce json
// @Router /configurations/{name}/duplicate [post]
// @Param 	name	path	string	true "the name of the configuration to duplicate"
// @Param name	body	string	true "the desired name of the duplicate configuration"
// @Success 201	"Successful Duplication, created"
// @Failure 404 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func duplicateConfig(c *gin.Context, bindplane server.BindPlane) {
	name := c.Param("name")

	// The config to make a duplicate of
	config, err := bindplane.Store().Configuration(name)
	if err != nil {
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	if config == nil {
		handleErrorResponse(c, http.StatusNotFound, fmt.Errorf("no configuration with name %s found", name))
		return
	}

	var req model.PostDuplicateConfigRequest
	if err := c.BindJSON(&req); err != nil {
		handleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	duplicateName := req.Name
	duplicateConfig, err := bindplane.Store().Configuration(duplicateName)
	if err != nil {
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	if duplicateConfig != nil {
		handleErrorResponse(c, http.StatusConflict, errors.New("a configuration with that name already exists"))
		return
	}

	duplicateConfig = config.Duplicate(duplicateName)

	_, err = bindplane.Store().ApplyResources([]model.Resource{duplicateConfig})
	if err != nil {
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, &model.PostDuplicateConfigResponse{
		Name: duplicateName,
	})
}

// ----------------------------------------------------------------------

// @Summary List sources
// @Produce json
// @Router /sources [get]
// @Success 200 {object} model.SourcesResponse
// @Failure 500 {object} ErrorResponse
func sources(c *gin.Context, bindplane server.BindPlane) {
	sources, err := bindplane.Store().Sources()
	if okResponse(c, err) {
		c.JSON(http.StatusOK, model.SourcesResponse{
			Sources: sources,
		})
	}
}

// @Summary Get source by name
// @Produce json
// @Router /sources/{name} [get]
// @Param 	name	path	string	true "the name of the source"
// @Success 200 {object} model.SourceResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func source(c *gin.Context, bindplane server.BindPlane) {
	name := c.Param("name")
	source, err := bindplane.Store().Source(name)
	if okResource(c, source == nil, err) {
		c.JSON(http.StatusOK, model.SourceResponse{
			Source: source,
		})
	}
}

// @Summary Delete source by name
// @Produce json
// @Router /sources/{name} [delete]
// @Param 	name	path	string	true "the name of the source to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func deleteSource(c *gin.Context, bindplane server.BindPlane) {
	name := c.Param("name")
	source, err := bindplane.Store().DeleteSource(name)

	if okResource(c, source == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// ----------------------------------------------------------------------

// @Summary List source types
// @Produce json
// @Router /source-types [get]
// @Success 200 {object} model.SourceTypesResponse
// @Failure 500 {object} ErrorResponse
func sourceTypes(c *gin.Context, bindplane server.BindPlane) {
	sourceTypes, err := bindplane.Store().SourceTypes()
	if okResponse(c, err) {
		c.JSON(http.StatusOK, model.SourceTypesResponse{
			SourceTypes: sourceTypes,
		})
	}
}

// @Summary Get source type by name
// @Produce json
// @Router /source-types/{name} [get]
// @Param 	name	path	string	true "the name of the source type"
// @Success 200 {object} model.SourceTypeResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func sourceType(c *gin.Context, bindplane server.BindPlane) {
	name := c.Param("name")
	sourceType, err := bindplane.Store().SourceType(name)
	if okResource(c, sourceType == nil, err) {
		c.JSON(http.StatusOK, model.SourceTypeResponse{
			SourceType: sourceType,
		})
	}
}

// @Summary Delete source type by name
// @Produce json
// @Router /source-types/{name} [delete]
// @Param 	name	path	string	true "the name of the source type to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func deleteSourceType(c *gin.Context, bindplane server.BindPlane) {
	name := c.Param("name")
	sourceType, err := bindplane.Store().DeleteSourceType(name)
	if okResource(c, sourceType == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// ----------------------------------------------------------------------

// @Summary List processors
// @Produce json
// @Router /processors [get]
// @Success 200 {object} model.ProcessorsResponse
// @Failure 500 {object} ErrorResponse
func processors(c *gin.Context, bindplane server.BindPlane) {
	processors, err := bindplane.Store().Processors()
	if okResponse(c, err) {
		c.JSON(http.StatusOK, model.ProcessorsResponse{
			Processors: processors,
		})
	}
}

// @Summary Get processor by name
// @Produce json
// @Router /processors/{name} [get]
// @Param 	name	path	string	true "the name of the processor"
// @Success 200 {object} model.ProcessorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func processor(c *gin.Context, bindplane server.BindPlane) {
	name := c.Param("name")
	processor, err := bindplane.Store().Processor(name)
	if okResource(c, processor == nil, err) {
		c.JSON(http.StatusOK, model.ProcessorResponse{
			Processor: processor,
		})
	}
}

// @Summary Delete processor by name
// @Produce json
// @Router /processors/{name} [delete]
// @Param 	name	path	string	true "the name of the processor to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func deleteProcessor(c *gin.Context, bindplane server.BindPlane) {
	name := c.Param("name")
	processor, err := bindplane.Store().DeleteProcessor(name)
	if okResource(c, processor == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// ----------------------------------------------------------------------

// @Summary List processor types
// @Produce json
// @Router /processor-types [get]
// @Success 200 {object} model.ProcessorTypesResponse
// @Failure 500 {object} ErrorResponse
func processorTypes(c *gin.Context, bindplane server.BindPlane) {
	processorTypes, err := bindplane.Store().ProcessorTypes()
	if okResponse(c, err) {
		c.JSON(http.StatusOK, model.ProcessorTypesResponse{
			ProcessorTypes: processorTypes,
		})
	}
}

// @Summary Get processor type by name
// @Produce json
// @Router /processor-types/{name} [get]
// @Param 	name	path	string	true "the name of the processor type"
// @Success 200 {object} model.ProcessorTypeResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func processorType(c *gin.Context, bindplane server.BindPlane) {
	name := c.Param("name")
	processorType, err := bindplane.Store().ProcessorType(name)
	if okResource(c, processorType == nil, err) {
		c.JSON(http.StatusOK, model.ProcessorTypeResponse{
			ProcessorType: processorType,
		})
	}
}

// @Summary Delete processor type by name
// @Produce json
// @Router /processor-types/{name} [delete]
// @Param 	name	path	string	true "the name of the processor type to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func deleteProcessorType(c *gin.Context, bindplane server.BindPlane) {
	name := c.Param("name")
	processorType, err := bindplane.Store().DeleteProcessorType(name)
	if okResource(c, processorType == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// ----------------------------------------------------------------------

// @Summary List destinations
// @Produce json
// @Router /destinations [get]
// @Success 200 {object} model.DestinationsResponse
// @Failure 500 {object} ErrorResponse
func destinations(c *gin.Context, bindplane server.BindPlane) {
	destinations, err := bindplane.Store().Destinations()
	if okResponse(c, err) {
		c.JSON(http.StatusOK, model.DestinationsResponse{
			Destinations: destinations,
		})
	}
}

// @Summary Get destination by name
// @Produce json
// @Router /destinations/{name} [get]
// @Param 	name	path	string	true "the name of the destination"
// @Success 200 {object} model.DestinationResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func destination(c *gin.Context, bindplane server.BindPlane) {
	name := c.Param("name")
	destination, err := bindplane.Store().Destination(name)
	if okResource(c, destination == nil, err) {
		c.JSON(http.StatusOK, model.DestinationResponse{
			Destination: destination,
		})
	}
}

// @Summary Delete destination by name
// @Produce json
// @Router /destinations/{name} [delete]
// @Param 	name	path	string	true "the name of the destination to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func deleteDestination(c *gin.Context, bindplane server.BindPlane) {
	name := c.Param("name")
	destination, err := bindplane.Store().DeleteDestination(name)
	if okResource(c, destination == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// ----------------------------------------------------------------------

// @Summary List destination types
// @Produce json
// @Router /destination-types [get]
// @Success 200 {object} model.DestinationTypesResponse
// @Failure 500 {object} ErrorResponse
func destinationTypes(c *gin.Context, bindplane server.BindPlane) {
	destinationTypes, err := bindplane.Store().DestinationTypes()
	if okResponse(c, err) {
		c.JSON(http.StatusOK, model.DestinationTypesResponse{
			DestinationTypes: destinationTypes,
		})
	}
}

// @Summary Get destination type by name
// @Produce json
// @Router /destination-types/{name} [get]
// @Param 	name	path	string	true "the name of the destination type"
// @Success 200 {object} model.DestinationTypeResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func destinationType(c *gin.Context, bindplane server.BindPlane) {
	name := c.Param("name")
	destinationType, err := bindplane.Store().DestinationType(name)
	if okResource(c, destinationType == nil, err) {
		c.JSON(http.StatusOK, model.DestinationTypeResponse{
			DestinationType: destinationType,
		})
	}
}

// @Summary Delete destination type by name
// @Produce json
// @Router /destination-types/{name} [delete]
// @Param 	name	path	string	true "the name of the destination type to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func deleteDestinationType(c *gin.Context, bindplane server.BindPlane) {
	name := c.Param("name")
	destinationType, err := bindplane.Store().DeleteDestinationType(name)
	if okResource(c, destinationType == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// ----------------------------------------------------------------------

// @Summary Create, edit, and configure multiple resources.
// @Description The /apply route will try to parse resources
// @Description and upsert them into the store.  Additionally
// @Description it will send reconfigure tasks to affected agents.
// @Produce json
// @Router /apply [post]
// @Param resources 	body	[]model.AnyResource	true "Resources"
// @Success 200 {object} model.ApplyResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func applyResources(c *gin.Context, bindplane server.BindPlane) {
	p := &model.ApplyPayload{}
	if err := c.BindJSON(p); err != nil {
		handleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// parse the resources
	resources := []model.Resource{}
	for _, res := range p.Resources {
		parsed, err := model.ParseResource(res)
		// TODO (dsvanlani): Go through all resources and gather errors.
		if err != nil {
			handleErrorResponse(c, http.StatusBadRequest, err)
			return
		}

		resources = append(resources, parsed)
	}

	bindplane.Logger().Info("/apply", zap.Int("count", len(resources)))

	resourceStatuses, err := bindplane.Store().ApplyResources(resources)
	if err != nil {
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusAccepted, &model.ApplyResponse{
		Updates: resourceStatuses,
	})
}

// @Summary Delete multiple resources
// @Description /delete endpoint will try to parse resources
// @Description and delete them from the store.  Additionally
// @Description it will send reconfigure tasks to affected agents.
// @Produce json
// @Router /delete [post]
// @Param resources 	body	[]model.AnyResource	true "Resources"
// @Success 200 {object} model.DeleteResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func deleteResources(c *gin.Context, bindplane server.BindPlane) {
	p := &model.DeletePayload{}
	if err := c.BindJSON(p); err != nil {
		handleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// parse the resources
	resources := []model.Resource{}
	for _, res := range p.Resources {
		parsed, err := model.ParseResource(res)
		if err != nil {
			handleErrorResponse(c, http.StatusBadRequest, err)
			return
		}
		resources = append(resources, parsed)
	}

	bindplane.Logger().Info("/delete", zap.Int("count", len(resources)))

	resourceStatuses, err := bindplane.Store().DeleteResources(resources)
	if err != nil {
		handleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusAccepted, &model.DeleteResponse{
		Updates: resourceStatuses,
	})
}

// @Summary Server version
// @Description Returns the current bindplane version of the server.
// @Produce json
// @Router /version [get]
// @Success 200 {string} version.Version
func bindplaneVersion(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, version.NewVersion())
}

// @Summary Get Install Command
// @Description Get the proper install command for the provided parameters.
// @Produce json
// @Router /agent-versions/{version}/install-command [get]
// @Param version 	path	string	true "2.1.1"
// @Param secret-key query string false "uuid"
// @Param remote-url query string false "http%3A%2F%2Flocalhost%3A3001"
// @Param platform query string false "windows-amd64"
// @Param labels query string false "env=stage,app=bindplane"
// @Success 200 {object} model.InstallCommandResponse
func getInstallCommand(c *gin.Context, bindplane server.BindPlane) {
	config := bindplane.Config()

	// note: don't use DefaultQuery because caller may specify secret-key=(empty string) but we want to use the default
	// value in that case
	secretKey := c.Query("secret-key")
	if secretKey == "" {
		secretKey = config.SecretKey
	}

	remoteURL := c.Query("remote-url")
	if remoteURL == "" {
		remoteURL = fmt.Sprintf("%s/v1/opamp", config.WebsocketURL())
	}

	serverURL := bindplane.Config().BindPlaneURL()

	// if version is empty or "latest", find the latest version
	version := c.Param("version")
	// if version == "" || version == "latest" {
	// 	v, err := bindplane.Versions().LatestVersion()
	// 	if err != nil {
	// 		handleErrorResponse(c, http.StatusInternalServerError,
	// 			fmt.Errorf("unable to get the latest version of the agent: %w", err),
	// 		)
	// 		c.Status(http.StatusInternalServerError)
	// 		return
	// 	}
	// 	version = v.Version
	// }

	platform, ok := normalizePlatform(c.Query("platform"))
	if !ok {
		handleErrorResponse(c, http.StatusBadRequest,
			fmt.Errorf("unknown platform: %s", c.Query("platform")),
		)
	}

	params := installCommandParameters{
		platform:  platform,
		version:   version,
		labels:    c.Query("labels"),
		secretKey: secretKey,
		remoteURL: remoteURL,
		serverURL: serverURL,
	}
	response := model.InstallCommandResponse{
		Command: params.installCommand(),
	}
	c.JSON(http.StatusOK, response)
}

// ----------------------------------------------------------------------

// okResponse returns true if there should be an OK response based on the error provided. It will set an error response on the
// gin.Context if appropriate.
func okResponse(c *gin.Context, err error) bool {
	switch {
	case err == nil:
		return true
	case errors.Is(err, store.ErrResourceMissing):
		handleErrorResponse(c, http.StatusNotFound, err)
	case isDependencyError(err):
		handleErrorResponse(c, http.StatusConflict, err)
	default:
		handleErrorResponse(c, http.StatusInternalServerError, err)
	}
	return false
}

// okResource returns true if there should be an OK response based on the resource and error provided. It will set an
// error response on the gin.Context if appropriate.
func okResource(c *gin.Context, resourceIsNil bool, err error) bool {
	if !okResponse(c, err) {
		return false
	}
	if resourceIsNil {
		handleErrorResponse(c, http.StatusNotFound, store.ErrResourceMissing)
		return false
	}
	return true
}

func isDependencyError(err error) bool {
	_, ok := err.(*store.DependencyError)
	return ok
}
