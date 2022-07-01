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

	"github.com/mitchellh/mapstructure"

	"github.com/observiq/bindplane-op/common"
)

// Profile TODO(doc)
type Profile struct {
	ResourceMeta `yaml:",inline" json:",inline" mapstructure:",squash"`
	Spec         ProfileSpec `json:"spec" yaml:"spec" mapstructure:"spec"`
}

// ProfileSpec TODO(doc)
type ProfileSpec struct {
	common.Common  `mapstructure:",squash" yaml:",inline,omitempty"`
	common.Server  `mapstructure:"server" yaml:"server,omitempty"`
	common.Client  `mapstructure:"client" yaml:"client,omitempty"`
	common.Command `mapstructure:"command" yaml:"command,omitempty"`
}

// NewProfile takes a name and ProfileSpec and returns a *Profile.
func NewProfile(name string, spec ProfileSpec) *Profile {
	return NewProfileWithMetadata(Metadata{
		Name: name,
	}, spec)
}

// NewProfileWithMetadata takes a Metadata and ProfileSpec and returns a *Profile.
func NewProfileWithMetadata(metadata Metadata, spec ProfileSpec) *Profile {
	return &Profile{
		ResourceMeta: ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1beta",
			Kind:       KindProfile,
			Metadata:   metadata,
		},
		Spec: spec,
	}
}

// Context TODO(doc)
type Context struct {
	// ResourceMeta TODO(doc)
	ResourceMeta `yaml:",inline" json:",inline" mapstructure:",squash"`
	// Spec TODO(doc)
	Spec ContextSpec `json:"spec" yaml:"spec,omitempty" mapstructure:"spec"`
}

// ContextSpec TODO(doc)
type ContextSpec struct {
	// CurrentContext TODO(doc)
	CurrentContext string `json:"currentContext" yaml:"currentContext" mapstructure:"currentContext"`
}

// NewContext rtakes a name and ContextSpec and returns a *Context.
func NewContext(name string, spec ContextSpec) *Context {
	return NewContextWithMetadata(Metadata{
		Name: name,
	}, spec)
}

// NewContextWithMetadata takes a Metadata and ContextSpec and returns a *Context.
func NewContextWithMetadata(metadata Metadata, spec ContextSpec) *Context {
	return &Context{
		ResourceMeta: ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1beta",
			Kind:       KindContext,
			Metadata:   metadata,
		},
		Spec: spec,
	}
}

func parseProfile(r *AnyResource) (*Profile, error) {
	if r.Kind != KindProfile {
		return nil, fmt.Errorf("invalid resource kind: %s", r.Kind)
	}

	var spec ProfileSpec
	err := mapstructure.Decode(r.Spec, &spec)
	if err != nil {
		return nil, fmt.Errorf("failed to decode profile: %w", err)
	}
	return &Profile{
		ResourceMeta: r.ResourceMeta,
		Spec:         spec,
	}, nil
}

func parseContext(r *AnyResource) (*Context, error) {
	if r.Kind != KindContext {
		return nil, fmt.Errorf("invalid resource kind: %s", r.Kind)
	}

	var spec ContextSpec
	err := mapstructure.Decode(r.Spec, &spec)
	if err != nil {
		return nil, fmt.Errorf("failed to decode context: %w", err)
	}

	return &Context{
		ResourceMeta: r.ResourceMeta,
		Spec:         spec,
	}, nil
}
