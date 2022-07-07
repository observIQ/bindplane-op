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

package server

import (
	"context"
	"math"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/eventbus"
	"github.com/observiq/bindplane-op/internal/store"
	"github.com/observiq/bindplane-op/model"
)

var tracer = otel.Tracer("bindplane/manager")

const (
	// AgentCleanupInterval is the default agent cleanup interval.
	AgentCleanupInterval = time.Minute
	// AgentCleanupTTL is the default agent cleanup time to live.
	AgentCleanupTTL = 15 * time.Minute
	// AgentHeartbeatInterval is the default interval for the heartbeat sent to the agent to keep the websocket live.
	AgentHeartbeatInterval = 30 * time.Second
)

// Manager manages agent connects and communications with them
type Manager interface {
	// Start starts the manager and allows it to begin processing configuration changes
	Start(ctx context.Context)
	// EnableProtocol adds the protocol to the manager
	EnableProtocol(Protocol)
	// Agent returns the agent with the specified agentID.
	Agent(ctx context.Context, agentID string) (*model.Agent, error)
	// UpsertAgent adds a new Agent to the Store or updates an existing one
	UpsertAgent(ctx context.Context, agentID string, updater store.AgentUpdater) (*model.Agent, error)
	// AgentUpdates returns the updates that should be applied to an agent based on the current bindplane configuration
	AgentUpdates(ctx context.Context, agent *model.Agent) (*AgentUpdates, error)
	// VerifySecretKey checks to see if the specified secretKey matches configured secretKey
	VerifySecretKey(ctx context.Context, secretKey string) bool
	// ResourceStore provides access to the store to render configurations
	ResourceStore() model.ResourceStore
}

// ----------------------------------------------------------------------

type manager struct {
	// agentCleanupTicker   *time.Ticker
	// agentHeartbeatTicker *time.Ticker
	store     store.Store
	logger    *zap.Logger
	protocols []Protocol
	secretKey string
}

var _ Manager = (*manager)(nil)

// NewManager returns a new implementation of the Manager interface
func NewManager(config *common.Server, store store.Store, logger *zap.Logger) (Manager, error) {
	return &manager{
		// agentCleanupTicker:   time.NewTicker(AgentCleanupInterval),
		// agentHeartbeatTicker: time.NewTicker(AgentHeartbeatInterval),
		store:     store,
		logger:    logger,
		protocols: []Protocol{},
		secretKey: config.SecretKey,
	}, nil
}

func (m *manager) EnableProtocol(protocol Protocol) {
	m.protocols = append(m.protocols, protocol)
}

// Start TODO(doc)
func (m *manager) Start(ctx context.Context) {
	updatesChannel, unsubscribe := eventbus.Subscribe(m.store.Updates(), eventbus.WithChannel(make(chan *store.Updates, 10_000)))
	defer unsubscribe()

	for {
		select {
		case <-ctx.Done():
			// m.agentCleanupTicker.Stop()
			// m.agentHeartbeatTicker.Stop()
			return

		case updates := <-updatesChannel:
			m.logger.Info("Received configuration updates")
			m.handleUpdates(updates)

			// TODO: determine if these need to be replaced and if so, replace them
			// case <-m.agentCleanupTicker.C:
			// 	m.handleAgentCleanup()

			// case <-m.agentHeartbeatTicker.C:
			// 	m.handleAgentHeartbeat()
		}
	}
}

// helper for bookkeeping during updates
type pendingAgentUpdate struct {
	agent   *model.Agent
	updates *AgentUpdates
}

type pendingAgentUpdates map[string]pendingAgentUpdate

func (p pendingAgentUpdates) agent(agent *model.Agent) pendingAgentUpdate {
	u, ok := p[agent.ID]
	if ok {
		return u
	}
	u = pendingAgentUpdate{
		agent:   agent,
		updates: &AgentUpdates{},
	}
	p[agent.ID] = u
	return u
}

func (p pendingAgentUpdates) apply(ctx context.Context, m *manager) {
	ctx, span := tracer.Start(ctx, "manager/apply")
	defer span.End()

	// Number of workers is a quarter of the total or 10
	// This insures for small updates we don't spin up 10 workers for 1 or 2 updates
	numWorkers := int(math.Min(float64(len(p)/4)+1, 10))

	m.logger.Info("Creating workers", zap.Int("numWorkers", numWorkers), zap.Int("pending updates", len(p)))

	startTime := time.Now()
	// Spun up worker group
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	updateChan := make(chan *pendingAgentUpdate, numWorkers)
	for i := 0; i < numWorkers; i++ {
		go updateWorker(ctx, &wg, m, updateChan)
	}

	for _, pending := range p {
		if !pending.updates.Empty() {
			pendingCpy := pending
			updateChan <- &pendingCpy
		}
	}

	close(updateChan)
	wg.Wait()

	execTime := time.Since(startTime)
	m.logger.Info("Update Time", zap.String("dur", execTime.String()))
}

func updateWorker(ctx context.Context, wg *sync.WaitGroup, m *manager, updateChan <-chan *pendingAgentUpdate) {
	ctx, span := tracer.Start(ctx, "manager/updateWorker")
	defer span.End()

	defer wg.Done()

	for {
		pending, ok := <-updateChan
		if !ok {
			return
		}
		m.updateAgent(ctx, pending.agent, pending.updates)
	}
}

func (m *manager) handleUpdates(updates *store.Updates) {
	if updates.Empty() {
		return
	}
	ctx, span := tracer.Start(context.TODO(), "manager/handleUpdates")
	defer span.End()

	pending := pendingAgentUpdates{}

	for _, change := range updates.Agents {
		// on delete, disconnect
		if change.Type == store.EventTypeRemove {
			m.disconnect(change.Item.ID)
			continue
		}
		// otherwise, we only care able label changes
		if change.Type != store.EventTypeLabel {
			continue
		}
		agent := change.Item

		// only consider connected agents
		if !m.connected(agent.ID) {
			continue
		}

		// this is only triggered for label changes right now, so we can just update that field
		m.logger.Info("updating labels for agent", zap.String("agentID", agent.ID), zap.String("labels", agent.Labels.String()))
		labels := agent.Labels.Custom()
		pending.agent(agent).updates.Labels = &labels
	}

	for _, event := range updates.Configurations {
		configuration := event.Item
		agentIDs, err := m.store.AgentsIDsMatchingConfiguration(configuration)
		if err != nil {
			m.logger.Error("unable to apply configuration to agents", zap.String("configuration.name", configuration.Name()), zap.Error(err))
			continue
		}

		for _, agentID := range agentIDs {
			// only consider connected agents
			if !m.connected(agentID) {
				continue
			}

			agent, err := m.store.Agent(agentID)
			if err != nil {
				m.logger.Error("unable to apply configuration to agent", zap.String("agentID", agentID), zap.String("configuration.name", configuration.Name()), zap.Error(err))
				continue
			}

			// TODO(andy): support multiple matches with precedence
			if event.Type == store.EventTypeRemove {
				m.logger.Info("deleting configuration for agent", zap.String("agentID", agent.ID))

				// TODO(andy): we need a default configuration
				// https://github.com/observIQ/bindplane/issues/279
				// agentUpdates.Configuration = otel.EmptyConfig()
			} else {
				m.logger.Info("updating configuration for agent", zap.String("agentID", agent.ID))
				pending.agent(agent).updates.Configuration = configuration
			}
		}

	}

	pending.apply(ctx, m)
}

func (m *manager) Agent(ctx context.Context, agentID string) (*model.Agent, error) {
	return m.store.Agent(agentID)
}

func (m *manager) UpsertAgent(ctx context.Context, agentID string, updater store.AgentUpdater) (*model.Agent, error) {
	return m.store.UpsertAgent(ctx, agentID, updater)
}

// AgentUpdates returns the updates that should be applied to an agent based on the current bindplane configuration
func (m *manager) AgentUpdates(ctx context.Context, agent *model.Agent) (*AgentUpdates, error) {
	newConfiguration, err := m.store.AgentConfiguration(agent.ID)
	if err != nil {
		return nil, err
	}
	newLabels := agent.Labels.Custom()
	return &AgentUpdates{
		Labels:        &newLabels,
		Configuration: newConfiguration,
	}, nil
}

// VerifySecretKey checks to see if the specified secretKey matches configured secretKey. If the BindPlane server does not
// have a configured secretKey, this returns true.
func (m *manager) VerifySecretKey(ctx context.Context, secretKey string) bool {
	return m.secretKey == "" || m.secretKey == secretKey
}

// ResourceStore provides access to the store to render configurations
func (m *manager) ResourceStore() model.ResourceStore {
	return m.store
}

// handleAgentCleanup removes disconnected agents from the store.
func (m *manager) handleAgentCleanup() {
	_, span := tracer.Start(context.TODO(), "manager/handleAgentCleanup")
	defer span.End()

	now := time.Now()
	// TODO: in a cluster, move this to a job
	err := m.store.CleanupDisconnectedAgents(now.Add(-AgentCleanupTTL))
	if err != nil {
		m.logger.Error("error cleaning up disconnected agents", zap.Error(err))
	}
}

func (m *manager) handleAgentHeartbeat() {
	ctx, span := tracer.Start(context.TODO(), "manager/handleAgentHeartbeat")
	defer span.End()

	for _, p := range m.protocols {
		ids, err := p.ConnectedAgentIDs(ctx)
		if err != nil {
			m.logger.Error("unable to get connected agents", zap.String("protocol", p.Name()))
			continue
		}
		for _, id := range ids {
			err = p.SendHeartbeat(id)
			if err != nil {
				m.logger.Error("unable to get send agent heartbeat", zap.String("protocol", p.Name()), zap.String("agentID", id))
				continue
			}
		}
	}
}

// ----------------------------------------------------------------------
// Protocol usage

func (m *manager) disconnect(agentID string) bool {
	for _, p := range m.protocols {
		if p.Disconnect(agentID) {
			return true
		}
	}
	return false
}

func (m *manager) connected(agentID string) bool {
	for _, p := range m.protocols {
		if p.Connected(agentID) {
			return true
		}
	}
	return false
}

// connectedAgentIDs returns the list of agents connected using any protocol
func (m *manager) connectedAgentIDs(ctx context.Context) []string {
	ids := []string{}
	for _, p := range m.protocols {
		list, err := p.ConnectedAgentIDs(ctx)
		if err != nil {
			m.logger.Error("unable to get connected agents", zap.String("protocol", p.Name()))
			continue
		}
		ids = append(ids, list...)
	}
	return ids
}

func (m *manager) updateAgent(ctx context.Context, agent *model.Agent, updates *AgentUpdates) {
	for _, p := range m.protocols {
		err := p.UpdateAgent(ctx, agent, updates)
		if err != nil {
			m.logger.Error("unable to update agent", zap.String("agentID", agent.ID))
		}
	}
}
