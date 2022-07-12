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

	"github.com/observiq/bindplane-op/model"
	"github.com/open-telemetry/opamp-go/protobufs"
	opamp "github.com/open-telemetry/opamp-go/server/types"
	"go.uber.org/zap"
)

// ----------------------------------------------------------------------
// RemoteConfigStatus

type remoteConfigStatusSyncer struct{}

var _ messageSyncer[*protobufs.RemoteConfigStatus] = (*remoteConfigStatusSyncer)(nil)

func (s *remoteConfigStatusSyncer) message(msg *protobufs.AgentToServer) (result *protobufs.RemoteConfigStatus, exists bool) {
	result = msg.GetRemoteConfigStatus()
	return result, result != nil
}

func (s *remoteConfigStatusSyncer) agentCapabilitiesFlag() protobufs.AgentCapabilities {
	return protobufs.AgentCapabilities_AcceptsRemoteConfig
}

func (s *remoteConfigStatusSyncer) update(ctx context.Context, logger *zap.Logger, state *agentState, conn opamp.Connection, agent *model.Agent, value *protobufs.RemoteConfigStatus) error {
	state.Status.RemoteConfigStatus = value
	return nil
}

// updateAgentStatus modifies the agent status based on the RemoteConfigStatus, if available
func updateAgentStatus(logger *zap.Logger, agent *model.Agent, remoteStatus *protobufs.RemoteConfigStatus) {
	// if we failed the apply, enter or update an error state
	if remoteStatus.GetStatus() == protobufs.RemoteConfigStatus_FAILED {
		logger.Info("got RemoteConfigStatus_FAILED", zap.String("ErrorMessage", remoteStatus.ErrorMessage))
		agent.Status = model.Error
		agent.ErrorMessage = remoteStatus.ErrorMessage
		return
	}
	switch agent.Status {
	case model.Error:
		// only way to clear the error is to have a successful apply
		if remoteStatus.GetStatus() == protobufs.RemoteConfigStatus_APPLIED {
			agent.Status = model.Connected
			agent.ErrorMessage = ""
		}
	default:
		// either RemoteConfigStatus wasn't sent or wasn't failed
		agent.Status = model.Connected
		agent.ErrorMessage = ""
	}
}
