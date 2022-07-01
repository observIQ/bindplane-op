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
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/observiq/bindplane-op/common"
)

var (
	testConfigCommonConfig = common.Common{
		ServerURL: "https://remote-address.com",
		Host:      "192.168.64.1",
		Port:      "5000",
		Username:  "admin",
		Password:  "admin",
	}

	testProfile = NewProfileWithMetadata(Metadata{
		Name: "local",
	}, ProfileSpec{
		Server:  common.Server{},
		Client:  common.Client{},
		Command: common.Command{},
		Common:  testConfigCommonConfig,
	})

	testContext = NewContextWithMetadata(Metadata{}, ContextSpec{
		CurrentContext: "local",
	})
)

func TestResourcesFromFileErrors(t *testing.T) {
	var tests = []struct {
		file string
	}{
		{
			file: "source-malformed-yaml.yaml",
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("testfile:%s", test.file), func(t *testing.T) {
			_, err := ResourcesFromFile(filepath.Join("testfiles", test.file))
			assert.Error(t, err)
		})
	}
}

func TestResourceValidate(t *testing.T) {
	tests := []struct {
		name      string
		resource  Resource
		errorMsgs []string
	}{
		{
			name: "invalid name",
			resource: &Configuration{
				ResourceMeta: ResourceMeta{
					Metadata: Metadata{
						Name: "invalid=name",
					},
				},
			},
			errorMsgs: []string{
				"invalid=name is not a valid resource name",
			},
		},
		{
			name: "invalid kind unknown",
			resource: &AnyResource{
				ResourceMeta: ResourceMeta{
					Kind: KindUnknown,
					Metadata: Metadata{
						Name: "invalid-kind",
					},
				},
			},
			errorMsgs: []string{
				"Unknown is not a valid resource kind",
			},
		},
		{
			name: "invalid kind string",
			resource: &AnyResource{
				ResourceMeta: ResourceMeta{
					Kind: Kind("invalid"),
					Metadata: Metadata{
						Name: "invalid-kind",
					},
				},
			},
			errorMsgs: []string{
				"1 error occurred:\n\t* invalid is not a valid resource kind",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.resource.Validate()
			if len(test.errorMsgs) == 0 {
				require.NoError(t, err)
			} else {
				for _, errorMsg := range test.errorMsgs {
					require.Contains(t, err.Error(), errorMsg)
				}
			}
		})
	}
}

func TestParseSourceType(t *testing.T) {
	resources, err := ResourcesFromFile(filepath.Join("testfiles", "sourcetype-macos.yaml"))
	assert.NoError(t, err)

	parsed, err := ParseResources(resources)
	require.NoError(t, err)

	sourceType, ok := parsed[0].(*SourceType)
	require.True(t, ok)

	expect := &SourceType{
		ResourceType: ResourceType{
			ResourceMeta: ResourceMeta{
				APIVersion: "bindplane.observiq.com/v1beta",
				Kind:       "SourceType",
				Metadata: Metadata{
					Name:        "MacOS",
					DisplayName: "Mac OS",
					Description: "Log parser for MacOS",
					Icon:        "/public/bindplane-logo.png",
				},
			},
			Spec: ResourceTypeSpec{
				Version:            "0.0.2",
				SupportedPlatforms: []string{"macos"},
				Parameters: []ParameterDefinition{
					{
						Name:        "enable_system_log",
						Label:       "System Logs",
						Description: "Enable to collect MacOS system logs",
						Type:        "bool",
						Default:     true,
					},
					{
						Name:        "system_log_path",
						Label:       "System Log Path",
						Description: "The absolute path to the System log",
						Type:        "string",
						Default:     "/var/log/system.log",
						RelevantIf: []RelevantIfCondition{
							{
								Name:     "enable_system_log",
								Operator: "equals",
								Value:    true,
							},
						},
					},
					{
						Name:        "enable_install_log",
						Label:       "Install Logs",
						Description: "Enable to collect MacOS install logs",
						Type:        "bool",
						Default:     true,
					},
					{
						Name:        "install_log_path",
						Label:       "Install Log Path",
						Description: "The absolute path to the Install log",
						Type:        "string",
						Default:     "/var/log/install.log",
						RelevantIf: []RelevantIfCondition{
							{
								Name:     "enable_install_log",
								Operator: "equals",
								Value:    true,
							},
						},
					},
					{
						Name:    "collection_interval_seconds",
						Label:   "Collection Interval",
						Type:    "int",
						Default: "30",
					},
					{
						Name:        "start_at",
						Label:       "Start At",
						Description: "Start reading file from 'beginning' or 'end'",
						Type:        "enum",
						ValidValues: []string{"beginning", "end"},
						Default:     "end",
					},
				},
				Logs: ResourceTypeOutput{
					Receivers: ResourceTypeTemplate(strings.TrimLeft(`
- plugin/macos:
    plugin:
      name: macos
    parameters:
    - name: enable_system_log
      value: {{ .enable_system_log }}
    - name: system_log_path
      value: {{ .system_log_path }}
    - name: enable_install_log
      value: {{ .enable_install_log }}
    - name: install_log_path
      value: {{ .install_log_path }}
    - name: start_at
      value: {{ .start_at }}
- plugin/journald:
    plugin:
      name: journald
`, "\n")),
				},
				Metrics: ResourceTypeOutput{
					Receivers: ResourceTypeTemplate(strings.TrimLeft(`
- hostmetrics:
    collection_interval: 1m
    scrapers:
      load:
`, "\n")),
				},
			},
		},
	}
	require.Equal(t, expect, sourceType)
}
