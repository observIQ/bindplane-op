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

type packageStatusesSyncer struct{}

var _ messageSyncer[*protobufs.PackageStatuses] = (*packageStatusesSyncer)(nil)

func (s *packageStatusesSyncer) name() string {
	return "PackageStatuses"
}

func (s *packageStatusesSyncer) message(msg *protobufs.AgentToServer) (result *protobufs.PackageStatuses, exists bool) {
	result = msg.GetPackageStatuses()
	return result, result != nil
}

func (s *packageStatusesSyncer) agentCapabilitiesFlag() protobufs.AgentCapabilities {
	return protobufs.AgentCapabilities_ReportsPackageStatuses
}

func (s *packageStatusesSyncer) update(ctx context.Context, logger *zap.Logger, state *agentState, conn opamp.Connection, agent *model.Agent, value *protobufs.PackageStatuses) error {
	state.Status.PackageStatuses = value
	return nil
}
