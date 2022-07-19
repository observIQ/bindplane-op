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

// SourceType is a ResourceType used to define sources
type SourceType struct {
	ResourceType `yaml:",inline" json:",inline" mapstructure:",squash"`
}

// NewSourceType creates a new source-type with the specified name,
func NewSourceType(name string, parameters []ParameterDefinition) *SourceType {
	return NewSourceTypeWithSpec(name, ResourceTypeSpec{
		Parameters: parameters,
	})
}

// NewSourceTypeWithSpec creates a new source-type with the specified name and spec.
func NewSourceTypeWithSpec(name string, spec ResourceTypeSpec) *SourceType {
	return &SourceType{
		ResourceType: ResourceType{
			ResourceMeta: ResourceMeta{
				APIVersion: V1Alpha,
				Kind:       KindSourceType,
				Metadata: Metadata{
					Name: name,
				},
			},
			Spec: spec,
		},
	}
}

// GetKind returns "SourceType"
func (s *SourceType) GetKind() Kind {
	return KindSourceType
}
