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

// ProcessorType is a ResourceType used to define sources
type ProcessorType struct {
	ResourceType `yaml:",inline" json:",inline" mapstructure:",squash"`
}

// NewProcessorType creates a new processor-type with the specified name,
func NewProcessorType(name string, parameters []ParameterDefinition) *ProcessorType {
	return NewProcessorTypeWithSpec(name, ResourceTypeSpec{
		Parameters: parameters,
	})
}

// NewProcessorTypeWithSpec creates a new processor-type with the specified name and spec.
func NewProcessorTypeWithSpec(name string, spec ResourceTypeSpec) *ProcessorType {
	return &ProcessorType{
		ResourceType: ResourceType{
			ResourceMeta: ResourceMeta{
				APIVersion: V1Alpha,
				Kind:       KindProcessorType,
				Metadata: Metadata{
					Name: name,
				},
			},
			Spec: spec,
		},
	}
}

// GetKind returns "ProcessorType"
func (s *ProcessorType) GetKind() Kind {
	return KindProcessorType
}
