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
	"fmt"
	"io"
)

// ResourceStatus contains a resource and its status after an update, which is of type UpdateStatus
// used in store and rest packages.
type ResourceStatus struct {
	// Resource TODO(doc)
	Resource Resource `json:"resource" mapstructure:"resource"`
	// Status TODO(doc)
	Status UpdateStatus `json:"status" mapstructure:"status"`
	// Reason will be set if status is invalid or error
	Reason string `json:"reason" mapstructure:"reason"`
}

// AnyResourceStatus TODO(doc)
// Same as ResourceStatus but used by cli to parse response from the rest api.
type AnyResourceStatus struct {
	Resource AnyResource  `json:"resource" mapstructure:"resource"`
	Status   UpdateStatus `json:"status" mapstructure:"status"`
	Reason   string       `json:"reason" mapstructure:"reason"`
}

// Message returns the summary of the ResourceStatus, e.g. "exporter updated"
func (s *AnyResourceStatus) Message() string {
	if s.Reason != "" {
		return fmt.Sprintf("%s %s %s\n\t%s", s.Resource.Kind, s.Resource.Name(), s.Status, s.Reason)
	}
	return fmt.Sprintf("%s %s %s", s.Resource.Kind, s.Resource.Name(), s.Status)
}

func (s *ResourceStatus) String() string {
	return fmt.Sprintf("%s %s %s", s.Resource.GetKind(), s.Resource.Name(), s.Status)
}

// NewResourceStatus TODO(doc)
func NewResourceStatus(r Resource, s UpdateStatus) *ResourceStatus {
	return &ResourceStatus{Resource: r, Status: s}
}

// NewResourceStatusWithReason returns a status for an invalid resource
func NewResourceStatusWithReason(r Resource, s UpdateStatus, reason string) *ResourceStatus {
	return &ResourceStatus{Resource: r, Status: s, Reason: reason}
}

// UpdateStatus is part of ResourceStatus that indicates the result of ApplyResources and DeleteResources on the Store.
type UpdateStatus string

const (
	// StatusUnchanged indicates that there were no changes to a modified resource because the existing resource is the same
	StatusUnchanged UpdateStatus = "unchanged"

	// StatusConfigured indicates that changes were applied to an existing resource
	StatusConfigured UpdateStatus = "configured"

	// StatusCreated indicates that a new resource was created
	StatusCreated UpdateStatus = "created"

	// StatusDeleted indicates that a resource was deleted, either from the store or the current filtered view of resources
	StatusDeleted UpdateStatus = "deleted"

	// StatusInvalid represents an attempt to add or update a resource with an invalid resource
	StatusInvalid UpdateStatus = "invalid"

	// StatusError is used when an individual resource cannot be applied because of an error
	StatusError UpdateStatus = "error"

	// StatusInUse is used when attempting to delete a resource that is being referenced by another
	StatusInUse UpdateStatus = "in-use"
)

// PrintResourceUpdates TODO(doc)
func PrintResourceUpdates(writer io.Writer, resourceStatuses []*AnyResourceStatus) {
	for _, update := range resourceStatuses {
		fmt.Fprintln(writer, update.Message())
	}
}
