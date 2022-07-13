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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigurationValidate(t *testing.T) {
	tests := []struct {
		testfile                     string
		expectValidateError          string
		expectValidateWithStoreError string
	}{
		{
			testfile:                     "configuration-invalid-spec-fields.yaml",
			expectValidateError:          "1 error occurred:\n\t* configuration must specify raw or sources and destinations\n\n",
			expectValidateWithStoreError: "1 error occurred:\n\t* configuration must specify raw or sources and destinations\n\n",
		},

		{
			testfile:                     "configuration-raw-malformed.yaml",
			expectValidateError:          "1 error occurred:\n\t* unable to parse spec.raw as yaml: yaml: line 29: did not find expected key\n\n",
			expectValidateWithStoreError: "1 error occurred:\n\t* unable to parse spec.raw as yaml: yaml: line 29: did not find expected key\n\n",
		},
		{
			testfile:                     "configuration-bad-name.yaml",
			expectValidateError:          "1 error occurred:\n\t* bad name is not a valid resource name: a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')\n\n",
			expectValidateWithStoreError: "1 error occurred:\n\t* bad name is not a valid resource name: a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')\n\n",
		},
		{
			testfile:                     "configuration-bad-labels.yaml",
			expectValidateError:          "1 error occurred:\n\t* bad label name is not a valid label name: name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')\n\n",
			expectValidateWithStoreError: "1 error occurred:\n\t* bad label name is not a valid label name: name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')\n\n",
		},
		{
			// sources and destinations must have valid resources with name and/or type (name takes precedence)
			testfile:                     "configuration-bad-resources.yaml",
			expectValidateError:          "",
			expectValidateWithStoreError: "4 errors occurred:\n\t* all Source parameters must have a name\n\t* all Source must have either a name or type\n\t* unknown Source: valid\n\t* unknown SourceType: unknown\n\n",
		},
		{
			testfile:                     "configuration-bad-parameter-values.yaml",
			expectValidateError:          "",
			expectValidateWithStoreError: "4 errors occurred:\n\t* parameter value for 'enable_system_log' must be a bool\n\t* parameter value for 'install_log_path' must be a string\n\t* parameter value for 'start_at' must be one of [beginning end]\n\t* parameter unknown not defined in type MacOS\n\n",
		},
		{
			testfile:                     "configuration-bad-selector.yaml",
			expectValidateError:          "1 error occurred:\n\t* selector is invalid: 1 error occurred:\n\t* bad key is not a valid label name: name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')\n\n\n\n",
			expectValidateWithStoreError: "1 error occurred:\n\t* selector is invalid: 1 error occurred:\n\t* bad key is not a valid label name: name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')\n\n\n\n",
		},
		{
			testfile:                     "configuration-ok.yaml",
			expectValidateError:          "",
			expectValidateWithStoreError: "",
		},
		{
			testfile:                     "configuration-ok-empty.yaml",
			expectValidateError:          "",
			expectValidateWithStoreError: "",
		},
	}

	store := newTestResourceStore()

	macos := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes[macos.Name()] = macos

	cabin := testResource[*Destination](t, "destination-cabin.yaml")
	store.destinations[cabin.Name()] = cabin

	cabinType := testResource[*DestinationType](t, "destinationtype-cabin.yaml")
	store.destinationTypes[cabinType.Name()] = cabinType

	for _, test := range tests {
		t.Run(test.testfile, func(t *testing.T) {
			config := validateResource[*Configuration](t, test.testfile)

			// test normal Validate() used by all resources
			err := config.Validate()
			if test.expectValidateError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, test.expectValidateError, err.Error())
			}

			// test special ValidateWithStore which can validate sources and destinations
			err = config.ValidateWithStore(store)
			if test.expectValidateWithStoreError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, test.expectValidateWithStoreError, err.Error())
			}
		})
	}
}

func TestSourceTypeValidate(t *testing.T) {
	tests := []struct {
		testfile           string
		expectErrorMessage string
	}{
		{
			testfile:           "sourcetype-ok.yaml",
			expectErrorMessage: "",
		},
		{
			testfile:           "sourcetype-bad-name.yaml",
			expectErrorMessage: "1 error occurred:\n\t* Mac OS is not a valid resource name: a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')\n\n",
		},
		{
			testfile:           "sourcetype-bad-labels.yaml",
			expectErrorMessage: "1 error occurred:\n\t* bad name is not a valid label name: name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')\n\n",
		},
		{
			testfile:           "sourcetype-bad-parameter-definitions.yaml",
			expectErrorMessage: "20 errors occurred:\n\t* missing type for 'no_type'\n\t* missing name for parameter\n\t* invalid name 'bad-name' for parameter\n\t* missing type for 'bad-name'\n\t* invalid type 'bad-type' for 'bad_type'\n\t* parameter of type 'enum' must have 'validValues' specified\n\t* validValues is undefined for parameter of type 'strings'\n\t* default value for 'bad_string_default' must be a string\n\t* default value for 'bad_bool_default' must be a bool\n\t* default value for 'bad_strings_default' must be an array of strings\n\t* default value for 'bad_int_default' must be an integer\n\t* default value for 'bad_int_default_as_float' must be an integer\n\t* default value for 'bad_enum_default' must be one of [1 2 3]\n\t* relevantIf for 'bad_relevant_if_2' must have a name\n\t* relevantIf for 'bad_relevant_if_2' refers to nonexistant parameter 'does_not_exist'\n\t* relevantIf 'string_default_1' for 'bad_relevant_if_2': relevantIf value for 'string_default_1' must be a string\n\t* relevantIf 'string_default_2' for 'bad_relevant_if_2' must have an operator\n\t* relevantIf 'string_default_3' for 'bad_relevant_if_2' must have a value\n\t* relevantIf 'bad_enum_default' for 'bad_relevant_if_2': relevantIf value for 'bad_enum_default' must be one of [1 2 3]\n\t* relevantIf 'bad_bool_default' for 'bad_relevant_if_2': relevantIf value for 'bad_bool_default' must be a bool\n\n",
		},
		{
			testfile:           "sourcetype-bad-templates.yaml",
			expectErrorMessage: "2 errors occurred:\n\t* template: logs.receivers:6: unexpected \"}\" in operand\n\t* template: logs.processors:1:5: executing \"logs.processors\" at <.not_a_variable>: map has no entry for key \"not_a_variable\"\n\n",
		},
	}

	for _, test := range tests {
		t.Run(test.testfile, func(t *testing.T) {
			config := validateResource[*SourceType](t, test.testfile)
			err := config.Validate()
			if test.expectErrorMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, test.expectErrorMessage, err.Error())
			}
		})
	}

}

func TestSourceValidate(t *testing.T) {
	tests := []struct {
		testfile                     string
		expectValidateError          string
		expectValidateWithStoreError string
	}{
		{
			testfile:                     "source-ok.yaml",
			expectValidateError:          "",
			expectValidateWithStoreError: "",
		},
		{
			testfile:                     "source-bad-name.yaml",
			expectValidateError:          "1 error occurred:\n\t* bar foo is not a valid resource name: a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')\n\n",
			expectValidateWithStoreError: "1 error occurred:\n\t* bar foo is not a valid resource name: a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')\n\n",
		},
		{
			testfile:                     "source-bad-labels.yaml",
			expectValidateError:          "1 error occurred:\n\t* bad label name is not a valid label name: name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')\n\n",
			expectValidateWithStoreError: "1 error occurred:\n\t* bad label name is not a valid label name: name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')\n\n",
		},
		{
			testfile:                     "source-bad-parameter-values.yaml",
			expectValidateError:          "",
			expectValidateWithStoreError: "3 errors occurred:\n\t* parameter value for 'install_log_path' must be a string\n\t* parameter value for 'start_at' must be one of [beginning end]\n\t* parameter unknown not defined in type MacOS\n\n",
		},
		{
			testfile:                     "source-bad-processor-type.yaml",
			expectValidateError:          "",
			expectValidateWithStoreError: "1 error occurred:\n\t* unknown ProcessorType: not_valid\n\n",
		},
		{
			testfile:                     "source-bad-processor-name.yaml",
			expectValidateError:          "",
			expectValidateWithStoreError: "1 error occurred:\n\t* unknown Processor: not_found\n\n",
		},
		{
			testfile:                     "source-bad-processor-parameter-values.yaml",
			expectValidateError:          "",
			expectValidateWithStoreError: "1 error occurred:\n\t* unknown ProcessorType: resource-attribute-transposer\n\n",
		},
	}

	store := newTestResourceStore()

	macos := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes[macos.Name()] = macos

	for _, test := range tests {
		t.Run(test.testfile, func(t *testing.T) {
			src := validateResource[*Source](t, test.testfile)

			t.Logf("Parsed config: %v", src)

			// test normal Validate() used by all resources
			err := src.Validate()
			if test.expectValidateError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, test.expectValidateError, err.Error())
			}

			// test special ValidateWithStore which can validate sources and destinations
			err = src.ValidateWithStore(store)
			if test.expectValidateWithStoreError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, test.expectValidateWithStoreError, err.Error())
			}
		})
	}

}
