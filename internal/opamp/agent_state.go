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

package opamp

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/mitchellh/mapstructure"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/model/observiq"
	"github.com/open-telemetry/opamp-go/protobufs"
	opamp "github.com/open-telemetry/opamp-go/server/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/runtime/protoiface"
)

var (
	syncAgentDescription   = agentDescriptionSyncer{}
	syncEffectiveConfig    = effectiveConfigSyncer{}
	syncRemoteConfigStatus = remoteConfigStatusSyncer{}
	syncPackageStatuses    = packageStatusesSyncer{}
)

func (s *opampServer) updateAgentState(ctx context.Context, agentID string, conn opamp.Connection, msg *protobufs.AgentToServer, response *protobufs.ServerToAgent) (agent *model.Agent, state *agentState, err error) {
	agent, err = s.manager.UpsertAgent(ctx, agentID, func(agent *model.Agent) {
		// we're using opamp
		agent.Protocol = ProtocolName

		// decode the state which we will update
		state, err = decodeState(agent)
		if err != nil {
			s.logger.Error("error encountered while decoding agent state", zap.Error(err))
			return
		}

		// TODO(andy): abort upsert if we know nothing changed?
		syncOne[*protobufs.AgentDescription](ctx, s.logger, msg, state, conn, agent, response, &syncAgentDescription)
		syncOne[*protobufs.EffectiveConfig](ctx, s.logger, msg, state, conn, agent, response, &syncEffectiveConfig)
		syncOne[*protobufs.RemoteConfigStatus](ctx, s.logger, msg, state, conn, agent, response, &syncRemoteConfigStatus)
		syncOne[*protobufs.PackageStatuses](ctx, s.logger, msg, state, conn, agent, response, &syncPackageStatuses)

		// after sync, update sequence number
		state.SequenceNum = msg.GetSequenceNum()

		// always update the agent status, regardless of RemoteConfigStatus message being present
		updateAgentStatus(s.logger, agent, state.Status.GetRemoteConfigStatus())

		// update ConnectedAt, etc
		if msg.GetAgentDisconnect() != nil {
			agent.Disconnect()
		} else {
			agent.Connect(agent.Version)
		}

		// the state could be new
		agent.State = state
	})

	return agent, state, err
}

// ----------------------------------------------------------------------

// agentState stores OpAMP state for the agent
type agentState struct {
	SequenceNum uint64
	Status      *protobufs.AgentToServer
}

func decodeState(agent *model.Agent) (*agentState, error) {
	result := &agentState{}
	err := mapstructure.Decode(agent.State, result)
	if result.Status == nil {
		result.Status = &protobufs.AgentToServer{}
	}
	return result, err
}

func (s *agentState) Configuration() *observiq.RawAgentConfiguration {
	if ec := s.Status.GetEffectiveConfig(); ec != nil {
		return agentCurrentConfiguration(ec)
	}
	return nil
}

func (s *agentState) UpdateSequenceNumber(agentToServer *protobufs.AgentToServer) {
	s.SequenceNum = agentToServer.GetSequenceNum()
}

func (s *agentState) IsMissingMessage(agentToServer *protobufs.AgentToServer) bool {
	return agentToServer.GetSequenceNum()-s.SequenceNum > 1
}

// ----------------------------------------------------------------------

// interface that defines how to sync each message
type messageSyncer[T protoiface.MessageV1] interface {
	// message returns the message within the AgentToServer the is being synced
	message(msg *protobufs.AgentToServer) (T, bool)

	// apply applies the updated message to the specified AgentToServer
	update(ctx context.Context, logger *zap.Logger, state *agentState, conn opamp.Connection, agent *model.Agent, value T) error

	// agentCapabilitiesFlag returns the flag to check on the agent to determine if it supports this message. If
	// unsupported, the reportFlag will not be specified.
	agentCapabilitiesFlag() protobufs.AgentCapabilities
}

// ----------------------------------------------------------------------

func syncOne[T protoiface.MessageV1](ctx context.Context, logger *zap.Logger, agentToServer *protobufs.AgentToServer, state *agentState, conn opamp.Connection, agent *model.Agent, response *protobufs.ServerToAgent, syncer messageSyncer[T]) (updated bool) {
	agentMessage, agentMessageExists := syncer.message(agentToServer)
	localMessage, localMessageExists := syncer.message(state.Status)

	// make sure we have a message
	if !agentMessageExists || state.IsMissingMessage(agentToServer) {
		// Either:
		//
		// 1) agent doesn't have the message at all => request contents
		//
		// 2) we missed a messages in sequence => request contents
		//
		if hasCapability(agentToServer, syncer.agentCapabilitiesFlag()) {
			response.Flags |= protobufs.ServerToAgent_ReportFullState
		}
		return false
	}

	if localMessageExists && proto.Equal(agentMessage, localMessage) {
		// data on the server is present and matches content => do nothing
		return false
	}

	// update
	if err := syncer.update(ctx, logger, state, conn, agent, agentMessage); err != nil {
		errorMessage := err.Error()
		if response.ErrorResponse != nil {
			errorMessage = response.ErrorResponse.ErrorMessage + ", " + errorMessage
		}
		response.ErrorResponse = &protobufs.ServerErrorResponse{
			Type:         protobufs.ServerErrorResponse_Unknown,
			ErrorMessage: errorMessage,
		}
	}

	return true
}

func hasCapability(agentToServer *protobufs.AgentToServer, capability protobufs.AgentCapabilities) bool {
	return capability&agentToServer.GetCapabilities() != 0
}
