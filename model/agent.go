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

import (
	"sort"
	"time"

	"github.com/observiq/bindplane-op/internal/store/search"
)

// AgentStatus TODO(doc)
type AgentStatus uint8

const (
	// Disconnected is the state of an agent that was formerly Connected to the management platform but is no longer
	// connected. This could mean that the agent has stopped running, the network connection has been interrupted, or that
	// the server terminated the connection.
	Disconnected AgentStatus = 0

	// Connected is the normal state of a healthy agent that is Connected to the management platform.
	Connected AgentStatus = 1

	// Error occurs if there is an error running or Configuring the agent.
	Error AgentStatus = 2

	// ComponentFailed is deprecated.
	ComponentFailed AgentStatus = 4

	// Deleted is set on a deleted Agent before notifying observers of the change.
	Deleted AgentStatus = 5

	// Configuring is set on an Agent when it is sent a new configuration that has not been applied. After successful
	// Configuring, it will transition back to Connected. If there is an error Configuring, it will transition to Error.
	Configuring AgentStatus = 6
)

// Agent TODO(doc)
type Agent struct {
	ID              string `json:"id" yaml:"id"`
	Name            string `json:"name" yaml:"name"`
	Type            string `json:"type" yaml:"type"`
	Architecture    string `json:"arch" yaml:"arch"`
	HostName        string `json:"hostname" yaml:"hostname"`
	Labels          Labels `json:"labels,omitempty" yaml:"labels"`
	Version         string `json:"version" yaml:"version"`
	Home            string `json:"home" yaml:"home"`
	Platform        string `json:"platform" yaml:"platform"`
	OperatingSystem string `json:"operatingSystem" yaml:"operatingSystem"`
	MacAddress      string `json:"macAddress" yaml:"macAddress"`
	RemoteAddress   string `json:"remoteAddress,omitempty" yaml:"remoteAddress,omitempty"`

	// SecretKey is provided by the agent to authenticate
	SecretKey string `json:"-" yaml:"-"`

	// reported by Status messages
	Status       AgentStatus `json:"status"`
	ErrorMessage string      `json:"errorMessage,omitempty" yaml:"errorMessage,omitempty"`

	// tracked by BindPlane
	Configuration  interface{} `json:"configuration,omitempty" yaml:"configuration,omitempty"`
	ConnectedAt    *time.Time  `json:"connectedAt,omitempty" yaml:"connectedAt,omitempty"`
	DisconnectedAt *time.Time  `json:"disconnectedAt,omitempty" yaml:"disconnectedAt,omitempty"`
}

var _ search.Indexed = (*Agent)(nil)
var _ HasUniqueKey = (*Agent)(nil)
var _ Labeled = (*Agent)(nil)

// UniqueKey returns the agent ID to uniquely identify an Agent
func (a *Agent) UniqueKey() string {
	return a.ID
}

// StatusDisplayText returns the string representation of the agent's status.
func (a *Agent) StatusDisplayText() string {
	switch a.Status {
	case Disconnected:
		return "Disconnected"
	case Connected:
		return "Connected"
	case Error:
		return "Error"
	case ComponentFailed:
		return "Component Failed"
	case Deleted:
		return "Deleted"
	case Configuring:
		return "Configuring"
	default:
		return "Unknown"
	}
}

// GetLabels implements the Labeled interface for Agents
func (a *Agent) GetLabels() Labels {
	return a.Labels
}

// ConnectedDurationDisplayText TODO(doc)
func (a *Agent) ConnectedDurationDisplayText() string {
	if a.Status == Disconnected {
		return "-"
	}
	return durationDisplay(a.ConnectedAt)
}

// DisconnectedDurationDisplayText TODO(doc) What RFC?
func (a *Agent) DisconnectedDurationDisplayText() string {
	return durationDisplay(a.DisconnectedAt)
}

// MatchesSelector returns true if the given selector matches the agent's labels.
func (a *Agent) MatchesSelector(selector Selector) bool {
	return selector.Matches(a.Labels)
}

// DisconnectedSince returns true if the agent has been disconnected since a given time.
func (a *Agent) DisconnectedSince(since time.Time) bool {
	return a.DisconnectedAt != nil || a.DisconnectedAt.Before(since)
}

// Connect updates the ConnectedAt and DisconnectedAt fields of the agent and should be called when the
// agent connects.
func (a *Agent) Connect(newAgentVersion string) {
	// only update ConnectedAt if this is a new version or never connected
	if a.Version != newAgentVersion || a.ConnectedAt == nil {
		now := time.Now()
		a.ConnectedAt = &now
	}
	a.DisconnectedAt = nil
}

// Disconnect updates the DisconnectedAt and Status fields of the agent and should be called when the agent disconnects.
func (a *Agent) Disconnect() {
	now := time.Now()
	a.DisconnectedAt = &now
	a.Status = Disconnected
}

func durationDisplay(t *time.Time) string {
	if t == nil || t.IsZero() {
		return "-"
	}
	return time.Since(*t).Round(time.Second).String()
}

// ----------------------------------------------------------------------
// sorting

type byName []*Agent

func (s byName) Len() int {
	return len(s)
}
func (s byName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

// SortAgentsByName TODO(doc)
func SortAgentsByName(agents []*Agent) {
	sort.Sort(byName(agents))
}

// ----------------------------------------------------------------------
// indexing

// IndexID returns an ID used to identify the resource that is indexed
func (a *Agent) IndexID() string {
	return a.ID
}

// IndexFields returns a map of field name to field value to be stored in the index
func (a *Agent) IndexFields(index search.Indexer) {
	index("id", a.ID)
	index("arch", a.Architecture)
	index("hostname", a.HostName)
	index("platform", a.Platform)
	index("version", a.Version)
	index("name", a.Name)
	index("home", a.Home)
	index("os", a.OperatingSystem)
	index("macAddress", a.MacAddress)
	index("type", a.Type)
	index("status", a.StatusDisplayText())
}

// IndexLabels returns a map of label name to label value to be stored in the index
func (a *Agent) IndexLabels(index search.Indexer) {
	for n, v := range a.Labels.Set {
		index(n, v)
	}
}

// ----------------------------------------------------------------------
// Printable

// PrintableKindSingular returns the singular form of the Kind, e.g. "Agent"
func (a *Agent) PrintableKindSingular() string {
	return "Agent"
}

// PrintableKindPlural returns the singular form of the Kind, e.g. "Agents"
func (a *Agent) PrintableKindPlural() string {
	return "Agents"
}

// PrintableFieldTitles returns the list of field titles, used for printing a table of resources
func (a *Agent) PrintableFieldTitles() []string {
	return []string{"ID", "Name", "Version", "Status", "Connected", "Disconnected", "Labels"}
}

// PrintableFieldValue returns the field value for a title, used for printing a table of resources
func (a *Agent) PrintableFieldValue(title string) string {
	switch title {
	case "ID":
		return a.ID
	case "Name":
		return a.Name
	case "Version":
		return a.Version
	case "Status":
		return a.StatusDisplayText()
	case "Connected":
		return a.ConnectedDurationDisplayText()
	case "Disconnected":
		return a.DisconnectedDurationDisplayText()
	case "Labels":
		return a.Labels.Custom().String()
	}
	return ""
}
