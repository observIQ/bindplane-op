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

	"github.com/open-telemetry/opamp-go/protobufs"
	opamp "github.com/open-telemetry/opamp-go/server/types"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/model"
)

type agentDescription struct {
	AgentID         string
	AgentName       string
	AgentType       string
	Architecture    string
	Hostname        string
	MacAddress      string
	Labels          string
	Platform        string
	OperatingSystem string
	Version         string
}

func parseAgentDescription(desc *protobufs.AgentDescription) *agentDescription {
	labels := stringValue("service.labels", desc.NonIdentifyingAttributes)
	if labels == "" {
		// check IdentifyingAttributes for compatibility with existing agents
		// TODO: remove when those agents are no longer supported
		labels = stringValue("service.labels", desc.IdentifyingAttributes)
	}
	return &agentDescription{
		AgentID:         stringValue("service.instance.id", desc.IdentifyingAttributes),
		AgentName:       stringValue("service.instance.name", desc.IdentifyingAttributes),
		AgentType:       stringValue("service.name", desc.IdentifyingAttributes),
		Version:         stringValue("service.version", desc.IdentifyingAttributes),
		Labels:          labels,
		Architecture:    stringValue("os.arch", desc.NonIdentifyingAttributes),
		OperatingSystem: stringValue("os.details", desc.NonIdentifyingAttributes),
		Platform:        stringValue("os.family", desc.NonIdentifyingAttributes),
		Hostname:        stringValue("host.name", desc.NonIdentifyingAttributes),
		MacAddress:      stringValue("host.mac_address", desc.NonIdentifyingAttributes),
	}
}

func stringValue(key string, fields []*protobufs.KeyValue) string {
	for _, kv := range fields {
		if key == kv.Key {
			return kv.Value.GetStringValue()
		}
	}
	return ""
}

func (desc *agentDescription) labels() model.Labels {
	// the error from parsing labels is ignored because these are provided by the agents. valid labels will still be
	// parsed and invalid labels will be ignored.
	bindplaneLabels, _ := model.LabelsFromMap(map[string]string{
		model.LabelBindPlaneAgentID:      desc.AgentID,
		model.LabelBindPlaneAgentName:    desc.AgentName,
		model.LabelBindPlaneAgentVersion: desc.Version,
		model.LabelBindPlaneAgentHost:    desc.Hostname,
		model.LabelBindPlaneAgentOS:      desc.Platform,
		model.LabelBindPlaneAgentArch:    desc.Architecture,
	})
	if agentLabels, err := model.LabelsFromSelector(desc.Labels); err == nil {
		return model.LabelsFromMerge(agentLabels, bindplaneLabels)
	}
	return bindplaneLabels
}

// ----------------------------------------------------------------------
// AgentDescription

type agentDescriptionSyncer struct{}

var _ messageSyncer[*protobufs.AgentDescription] = (*agentDescriptionSyncer)(nil)

func (s *agentDescriptionSyncer) message(msg *protobufs.AgentToServer) (result *protobufs.AgentDescription, exists bool) {
	result = msg.GetAgentDescription()
	return result, result != nil
}

func (s *agentDescriptionSyncer) agentCapabilitiesFlag() protobufs.AgentCapabilities {
	// TODO(andy): this flag is ok to check and should be true for all agents, but there should probably be a
	// ReportsAgentDescription capability flag.
	return protobufs.AgentCapabilities_ReportsStatus
}

func (s *agentDescriptionSyncer) update(ctx context.Context, logger *zap.Logger, state *agentState, conn opamp.Connection, agent *model.Agent, value *protobufs.AgentDescription) error {
	state.Status.AgentDescription = value
	updateOpAmpAgentDetails(agent, conn, value)
	return nil
}

func updateOpAmpAgentDetails(agent *model.Agent, conn opamp.Connection, desc *protobufs.AgentDescription) {
	ad := parseAgentDescription(desc)
	if ad.AgentID != "" {
		agent.ID = ad.AgentID
	}
	agent.Type = ad.AgentType
	agent.Architecture = ad.Architecture
	agent.Name = ad.AgentName
	agent.HostName = ad.Hostname
	agent.Platform = ad.Platform
	agent.OperatingSystem = ad.OperatingSystem
	agent.Labels = ad.labels()
	agent.Version = ad.Version
	agent.MacAddress = ad.MacAddress
	if addr := conn.RemoteAddr(); addr != nil {
		agent.RemoteAddress = addr.String()
	} else {
		agent.RemoteAddress = ""
	}
}
