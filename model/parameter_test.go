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

func TestValidateDefault(t *testing.T) {
	testCases := []struct {
		name      string
		expectErr bool
		param     ParameterDefinition
	}{
		{
			"ValidStringDefault",
			false,
			ParameterDefinition{
				Type:    "string",
				Default: "test",
			},
		},
		{
			"InvalidStringDefault",
			true,
			ParameterDefinition{
				Type:    "string",
				Default: 5,
			},
		},
		{
			"ValidIntDefault",
			false,
			ParameterDefinition{
				Type:    "int",
				Default: 5,
			},
		},
		{
			"InvalidStringDefault",
			true,
			ParameterDefinition{
				Type:    "int",
				Default: "test",
			},
		},
		{
			"ValidBoolDefault",
			false,
			ParameterDefinition{
				Type:    "bool",
				Default: true,
			},
		},
		{
			"InvalidBoolDefault",
			true,
			ParameterDefinition{
				Type:    "bool",
				Default: "test",
			},
		},
		{
			"ValidStringsDefault",
			false,
			ParameterDefinition{
				Type:    "strings",
				Default: []interface{}{"test"},
			},
		},
		{
			"InvalidStringsDefault",
			true,
			ParameterDefinition{
				Type:    "strings",
				Default: []interface{}{5},
			},
		},
		{
			"ValidEnumDefault",
			false,
			ParameterDefinition{
				Type:        "enum",
				ValidValues: []string{"test"},
				Default:     "test",
			},
		},
		{
			"InvalidEnumDefault",
			true,
			ParameterDefinition{
				Type:        "enum",
				ValidValues: []string{"test"},
				Default:     "invalid",
			},
		},
		{
			"ValidEnumsDefaultEmpty",
			false,
			ParameterDefinition{
				Type:        "enums",
				ValidValues: []string{"foo", "bar", "baz", "blah"},
				Default:     []any{},
			},
		},
		{
			"ValidEnumsDefaultAll",
			false,
			ParameterDefinition{
				Type:        "enums",
				ValidValues: []string{"foo", "bar", "baz", "blah"},
				Default:     []any{"foo", "bar", "baz", "blah"},
			},
		},
		{
			"NonStringEnumDefault",
			true,
			ParameterDefinition{
				Type:        "enum",
				ValidValues: []string{"test"},
				Default:     5,
			},
		},
		{
			"InvalidTypeDefault",
			true,
			ParameterDefinition{
				Type:    "float",
				Default: 5,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.param.validateDefault()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateValue(t *testing.T) {
	testCases := []struct {
		name      string
		expectErr bool
		param     ParameterDefinition
		value     interface{}
	}{
		{
			"ValidString",
			false,
			ParameterDefinition{
				Type:    "string",
				Default: "test",
			},
			"string",
		},
		{
			"InvalidString",
			true,
			ParameterDefinition{
				Type:    "string",
				Default: "test",
			},
			5,
		},
		{
			"ValidInt",
			false,
			ParameterDefinition{
				Type:    "int",
				Default: 5,
			},
			5,
		},
		{
			"InvalidInt",
			true,
			ParameterDefinition{
				Type:    "int",
				Default: 5,
			},
			"test",
		},
		{
			"ValidBool",
			false,
			ParameterDefinition{
				Type:    "bool",
				Default: true,
			},
			false,
		},
		{
			"InvalidBool",
			true,
			ParameterDefinition{
				Type:    "bool",
				Default: true,
			},
			"test",
		},
		{
			"ValidStringsAsInterface",
			false,
			ParameterDefinition{
				Type:    "strings",
				Default: []interface{}{"test"},
			},
			[]interface{}{"test"},
		},
		{
			"ValidStrings",
			false,
			ParameterDefinition{
				Type:    "strings",
				Default: []interface{}{"test"},
			},
			[]string{"test"},
		},
		{
			"InvalidStringsAsInterface",
			true,
			ParameterDefinition{
				Type:    "strings",
				Default: []interface{}{"test"},
			},
			[]interface{}{5},
		},
		{
			"InvalidStrings",
			true,
			ParameterDefinition{
				Type:    "strings",
				Default: []interface{}{"test"},
			},
			[]int{5},
		},
		{
			"ValidEnum",
			false,
			ParameterDefinition{
				Type:        "enum",
				ValidValues: []string{"test"},
				Default:     "test",
			},
			"test",
		},
		{
			"InvalidEnumValue",
			true,
			ParameterDefinition{
				Type:        "enum",
				ValidValues: []string{"test"},
				Default:     "test",
			},
			"missing",
		},
		{
			"InvalidEnumtype",
			true,
			ParameterDefinition{
				Type:        "enum",
				ValidValues: []string{"test"},
				Default:     "test",
			},
			5,
		},
		{
			"InvalidType",
			true,
			ParameterDefinition{
				Type:    "float",
				Default: 5,
			},
			5,
		},
		{
			"ValidMap",
			false,
			ParameterDefinition{
				Type: "map",
			},
			map[string]string{
				"foo":  "bar",
				"blah": "baz",
			},
		},
		{
			"InvalidMap",
			true,
			ParameterDefinition{
				Type: "map",
			},
			5,
		},
		{
			"InvalidMapType",
			true,
			ParameterDefinition{
				Type: "map",
			},
			map[string]interface{}{
				"blah": 1,
				"foo":  "blah",
			},
		},
		{
			"ValidYaml",
			false,
			ParameterDefinition{
				Type: "yaml",
			},
			`blah: foo
bar: baz
baz:
- one
- two
`,
		}, {
			"ValidYaml",
			false,
			ParameterDefinition{
				Type: "yaml",
			},
			`- one
- two
`,
		},
		{
			"InvalidYaml",
			true,
			ParameterDefinition{
				Type: "yaml",
			},
			`one: two
three: four
- five: 6
seven:
	- eight
	- nine
	- 10
eleven:
	- twelve: thirteen
	fourteen: 15
`,
		},
		{
			"InvalidYaml",
			true,
			ParameterDefinition{
				Type: "yaml",
			},
			`{{{}}}`,
		},
		{
			"ValidEnums",
			false,
			ParameterDefinition{
				Type:        "enums",
				ValidValues: []string{"one", "two", "three", "four"},
			},
			[]any{
				"two", "four",
			},
		},
		{
			"InvalidEnums",
			true,
			ParameterDefinition{
				Type:        "enums",
				ValidValues: []string{"one", "two", "three", "four"},
			},
			[]any{
				"one", "seven",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// set the name for better error messages
			tc.param.Name = tc.name
			err := tc.param.validateValue(tc.value)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
