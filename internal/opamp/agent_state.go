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
	"encoding/base64"

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
		state, err = decodeState(agent.State)
		if err != nil {
			s.logger.Error("error encountered while decoding agent state, starting with fresh state", zap.Error(err))
		}

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
		agent.State = encodeState(state)
	})

	return agent, state, err
}

// ----------------------------------------------------------------------

// state is stored on the model.Agent in a partially serialized form. The status is a base64-encoded protobuf.
type serializedAgentState struct {
	SequenceNum uint64 `json:"sequenceNum" yaml:"sequenceNum" mapstructure:"sequenceNum"`
	Status      string `json:"status,omitempty" yaml:"status,omitempty" mapstructure:"status"`
}

// agentState stores OpAMP state for the agent
type agentState struct {
	SequenceNum uint64                  `json:"sequenceNum" yaml:"sequenceNum" mapstructure:"sequenceNum"`
	Status      protobufs.AgentToServer `json:"status,omitempty" yaml:"status,omitempty" mapstructure:"status"`
}

func encodeState(state *agentState) *serializedAgentState {
	if state == nil {
		return &serializedAgentState{}
	}
	bytes, err := proto.Marshal(&state.Status)
	if err != nil {
		bytes = nil
	}
	serialized := &serializedAgentState{
		SequenceNum: state.SequenceNum,
		Status:      base64.StdEncoding.EncodeToString(bytes),
	}
	return serialized
}

func decodeState(state interface{}) (*agentState, error) {
	serialized := serializedAgentState{}

	if err := mapstructure.Decode(state, &serialized); err != nil {
		return &agentState{
			SequenceNum: serialized.SequenceNum,
		}, err
	}

	result := &agentState{
		SequenceNum: serialized.SequenceNum,
	}

	bytes, err := base64.StdEncoding.DecodeString(serialized.Status)
	if err != nil {
		return result, err
	}

	// unmarshal proto
	if err := proto.Unmarshal(bytes, &result.Status); err != nil {
		return result, err
	}

	return result, nil
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
	// name is useful for debugging
	name() string

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
	localMessage, localMessageExists := syncer.message(&state.Status)

	initialSyncRequired := !localMessageExists && !agentMessageExists
	serverSkippedMessage := state.IsMissingMessage(agentToServer)

	// make sure we have a message
	if initialSyncRequired || serverSkippedMessage {
		// Either:
		//
		// 1) agent doesn't have the message at all => request contents
		//
		// 2) we missed a messages in sequence => request contents
		//
		logger.Debug("not synced or missed message => ReportFullState",
			zap.String("syncer", syncer.name()),
			zap.Bool("serverSkippedMessage", serverSkippedMessage),
			zap.Bool("initialSyncRequired", initialSyncRequired),
		)
		if hasCapability(agentToServer, syncer.agentCapabilitiesFlag()) {
			response.Flags |= protobufs.ServerToAgent_ReportFullState
		}
		return false
	}

	if localMessageExists {
		if !agentMessageExists || proto.Equal(agentMessage, localMessage) {
			// data on the server is present and matches content => do nothing
			logger.Debug("exists locally and unchanged => do nothing", zap.String("syncer", syncer.name()))
			return false
		}
	}

	// before attempting to store, make sure we clone the message
	agentMessage = proto.Clone(agentMessage).(T)

	// update
	if err := syncer.update(ctx, logger, state, conn, agent, agentMessage); err != nil {
		logger.Debug("message different => update error", zap.String("syncer", syncer.name()), zap.Error(err))
		errorMessage := err.Error()
		if response.ErrorResponse != nil {
			errorMessage = response.ErrorResponse.ErrorMessage + ", " + errorMessage
		}
		response.ErrorResponse = &protobufs.ServerErrorResponse{
			Type:         protobufs.ServerErrorResponse_Unknown,
			ErrorMessage: errorMessage,
		}
	} else {
		logger.Debug("message different => update", zap.String("syncer", syncer.name()))
	}

	return true
}

func hasCapability(agentToServer *protobufs.AgentToServer, capability protobufs.AgentCapabilities) bool {
	return capability&agentToServer.GetCapabilities() != 0
}

// ----------------------------------------------------------------------
// misc utils

func messageComponents(agentToServer *protobufs.AgentToServer) []string {
	var components []string
	components = includeComponent(components, agentToServer.AgentDescription, "AgentDescription")
	components = includeComponent(components, agentToServer.EffectiveConfig, "EffectiveConfig")
	components = includeComponent(components, agentToServer.RemoteConfigStatus, "RemoteConfigStatus")
	components = includeComponent(components, agentToServer.PackageStatuses, "PackageStatuses")
	return components
}

func includeComponent(components []string, msg any, name string) []string {
	if msg != nil {
		components = append(components, name)
	}
	return components
}
