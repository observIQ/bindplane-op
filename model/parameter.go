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
	"go/token"
	"reflect"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/observiq/bindplane-op/model/validation"
	"github.com/observiq/stanza/errors"
	"gopkg.in/yaml.v3"
)

const (
	stringType  = "string"
	boolType    = "bool"
	intType     = "int"
	stringsType = "strings"
	enumType    = "enum"
	enumsType   = "enums"
	yamlType    = "yaml"
	mapType     = "map"
)

// ParameterDefinition is a basic description of a definition's parameter. This implementation comes directly from
// stanza plugin parameters with slight modifications for mapstructure.
type ParameterDefinition struct {
	Name        string `json:"name" yaml:"name"`
	Label       string `json:"label" yaml:"label"`
	Description string `json:"description" yaml:"description"`
	Required    bool   `json:"required" yaml:"required"`

	// "string", "int", "bool", "strings", or "enum"
	Type string `json:"type" yaml:"type"`

	// only useable if Type == "enum"
	ValidValues []string `json:"validValues,omitempty" yaml:"validValues,omitempty" mapstructure:"validValues"`

	// Must be valid according to Type & ValidValues
	Default        interface{}           `json:"default,omitempty" yaml:"default,omitempty"`
	RelevantIf     []RelevantIfCondition `json:"relevantIf,omitempty" yaml:"relevantIf,omitempty" mapstructure:"relevantIf"`
	Hidden         bool                  `json:"hidden" yaml:"hidden"`
	AdvancedConfig bool                  `json:"advancedConfig" yaml:"advancedConfig" mapstructure:"advancedConfig"`

	Options ParameterOptions `json:"options" yaml:"options"`
}

// ParameterOptions specify further customization for inputs in the UI
type ParameterOptions struct {
	// Creatable will modify the "enum" parameter from a select to
	// a creatable select where a user can specify a custom value
	Creatable bool `json:"creatable" yaml:"creatable"`
}

// RelevantIfCondition specifies a condition under which a parameter is deemed relevant.
type RelevantIfCondition struct {
	Name     string `json:"name" yaml:"name" mapstructure:"name"`
	Operator string `json:"operator" yaml:"operator" mapstructure:"operator"`
	Value    any    `json:"value" yaml:"value" mapstructure:"value"`
}

func (p ParameterDefinition) validateValue(value interface{}) error {
	return p.validateValueType(parameterFieldValue, value)
}

func (p ParameterDefinition) validateDefinition(errs validation.Errors) {
	if err := p.validateName(); err != nil {
		errs.Add(err)
	}

	if err := p.validateType(); err != nil {
		errs.Add(err)
	}

	if err := p.validateValidValues(); err != nil {
		errs.Add(err)
	}

	if err := p.validateOptions(); err != nil {
		errs.Add(err)
	}

	if err := p.validateDefault(); err != nil {
		errs.Add(err)
	}
}

func (p ParameterDefinition) validateName() error {
	if p.Name == "" {
		return errors.NewError(
			"missing name for parameter",
			"ensure that the name is a valid go identifier that can be used in go templates",
		)
	}
	if !token.IsIdentifier(p.Name) {
		return errors.NewError(
			fmt.Sprintf("invalid name '%s' for parameter", p.Name),
			"ensure that the name is a valid go identifier that can be used in go templates",
		)
	}
	return nil
}

func (p ParameterDefinition) validateType() error {
	if p.Type == "" {
		return errors.NewError(
			fmt.Sprintf("missing type for '%s'", p.Name),
			"ensure that the type is one of 'string', 'int', 'bool', 'strings', or 'enum'",
		)
	}
	switch p.Type {
	case stringType, intType, boolType, stringsType, enumType, enumsType, mapType, yamlType: // ok
	default:
		return errors.NewError(
			fmt.Sprintf("invalid type '%s' for '%s'", p.Type, p.Name),
			"ensure that the type is one of 'string', 'int', 'bool', 'strings', or 'enum'",
		)
	}
	return nil
}

func (p ParameterDefinition) validateOptions() error {
	err := &multierror.Error{}

	if p.Options.Creatable && p.Type != "enum" {
		multierror.Append(err,
			errors.NewError(
				fmt.Sprintf("creatable is true for parameter of type '%s'", p.Type),
				"remove 'creatable' field or change type to 'enum'",
			),
		)
	}

	return err.ErrorOrNil()
}

func (p ParameterDefinition) validateValidValues() error {
	switch p.Type {
	case stringType, intType, boolType, stringsType, yamlType, mapType:
		if len(p.ValidValues) > 0 {
			return errors.NewError(
				fmt.Sprintf("validValues is undefined for parameter of type '%s'", p.Type),
				"remove 'validValues' field or change type to 'enum' or 'enums",
			)
		}
	case enumType, enumsType:
		if len(p.ValidValues) == 0 {
			return errors.NewError(
				"parameter of type 'enum' or 'enums' must have 'validValues' specified",
				"specify an array that includes one or more valid values",
			)
		}
	}
	return nil
}

func (p ParameterDefinition) validateDefault() error {
	if p.Default == nil {
		return nil
	}

	// Validate that Default corresponds to Type
	return p.validateValueType(parameterFieldDefault, p.Default)
}

type parameterFieldType string

const (
	parameterFieldValue      parameterFieldType = "parameter"
	parameterFieldDefault                       = "default"
	parameterFieldRelevantIf                    = "relevantIf"
)

// validateValueType determines if the specified value is of the right type.
func (p ParameterDefinition) validateValueType(fieldType parameterFieldType, value any) error {
	switch p.Type {
	case stringType:
		return p.validateStringValue(fieldType, value)
	case intType:
		return p.validateIntValue(fieldType, value)
	case boolType:
		return p.validateBoolValue(fieldType, value)
	case stringsType:
		return p.validateStringArrayValue(fieldType, value)
	case enumType:
		return p.validateEnumValue(fieldType, value)
	case enumsType:
		return p.validateEnumsValue(fieldType, value)
	case mapType:
		return p.validateMapValue(fieldType, value)
	case yamlType:
		return p.validateYamlValue(fieldType, value)
	default:
		return errors.NewError(
			"invalid type for parameter",
			"ensure that the type is one of 'string', 'int', 'bool', 'strings', or 'enum'",
		)
	}
}

func (p ParameterDefinition) validateStringValue(fieldType parameterFieldType, value any) error {
	if _, ok := value.(string); !ok {
		return errors.NewError(
			fmt.Sprintf("%s value for '%s' must be a string", fieldType, p.Name),
			fmt.Sprintf("ensure that the %s value is a string", fieldType),
		)
	}
	return nil
}

func (p ParameterDefinition) validateIntValue(fieldType parameterFieldType, value any) error {
	isIntValue := false

	if _, ok := value.(int); ok {
		// obvious case of integer
		isIntValue = true
	} else if floatValue, ok := value.(float64); ok {
		// less obvious case of float64
		if floatValue == float64(int(floatValue)) {
			isIntValue = true
		}
	} else if stringValue, ok := value.(string); ok {
		_, err := strconv.Atoi(stringValue)
		isIntValue = err == nil
	}

	if !isIntValue {
		return errors.NewError(
			fmt.Sprintf("%s value for '%s' must be an integer", fieldType, p.Name),
			fmt.Sprintf("ensure that the %s value is an integer", fieldType),
		)
	}
	return nil
}

func (p ParameterDefinition) validateBoolValue(fieldType parameterFieldType, value any) error {
	isBoolValue := false

	if _, ok := value.(bool); ok {
		isBoolValue = true
	} else if stringValue, ok := value.(string); ok {
		_, err := strconv.ParseBool(stringValue)
		isBoolValue = err == nil
	}

	if !isBoolValue {
		return errors.NewError(
			fmt.Sprintf("%s value for '%s' must be a bool", fieldType, p.Name),
			fmt.Sprintf("ensure that the %s value is a bool", fieldType),
		)
	}
	return nil
}

func (p ParameterDefinition) validateStringArrayValue(fieldType parameterFieldType, value any) error {
	if _, ok := value.([]string); ok {
		return nil
	}
	valueList, ok := value.([]interface{})
	if !ok {
		return errors.NewError(
			fmt.Sprintf("%s value for '%s' must be an array of strings", fieldType, p.Name),
			fmt.Sprintf("ensure that the %s value is an array of string", fieldType),
		)
	}
	for _, s := range valueList {
		if _, ok := s.(string); !ok {
			return errors.NewError(
				fmt.Sprintf("%s value for '%s' must be an array of strings", fieldType, p.Name),
				fmt.Sprintf("ensure that the %s value is an array of string", fieldType),
			)
		}
	}
	return nil
}

func (p ParameterDefinition) validateEnumValue(fieldType parameterFieldType, value any) error {
	def, ok := value.(string)
	if !ok {
		return errors.NewError(
			fmt.Sprintf("%s value for enumerated parameter '%s'.", fieldType, p.Name),
			fmt.Sprintf("ensure that the %s value is a string", fieldType),
		)
	}

	// If the enum is creatable thats all we need to check - any string value is valid.
	if p.Options.Creatable {
		return nil
	}

	for _, val := range p.ValidValues {
		if val == def {
			return nil
		}
	}
	return errors.NewError(
		fmt.Sprintf("%s value for '%s' must be one of %v", fieldType, p.Name, p.ValidValues),
		fmt.Sprintf("ensure %s value is listed as a valid value", fieldType),
	)
}

func (p ParameterDefinition) validateEnumsValue(fieldType parameterFieldType, value any) error {
	def, ok := value.([]any)
	if !ok {
		return errors.NewError(
			fmt.Sprintf("%s value for enums parameter '%s'", fieldType, p.Name),
			fmt.Sprintf("ensure that the %s value is a string array", fieldType),
		)
	}

	// Make sure all strings in the value are a validValue
	err := &multierror.Error{}
	for _, str := range def {
		var containsValue bool
		for _, validValue := range p.ValidValues {
			if str == validValue {
				containsValue = true
				break
			}
		}

		if !containsValue {
			multierror.Append(err,
				errors.NewError(
					fmt.Sprintf("%s value for '%s' must be one of %v", fieldType, p.Name, p.ValidValues),
					fmt.Sprintf("ensure that all values for %s are in %v", p.Name, p.ValidValues),
				),
			)
		}
	}

	return err.ErrorOrNil()
}

func (p ParameterDefinition) validateYamlValue(fieldType parameterFieldType, value any) error {
	str, ok := value.(string)
	if !ok {
		return errors.NewError(
			fmt.Sprintf("expected a string for parameter %s", p.Name),
			fmt.Sprintf("ensure that the value is a string"),
		)
	}

	var into any
	return yaml.Unmarshal([]byte(str), &into)
}

func (p ParameterDefinition) validateMapValue(fieldType parameterFieldType, value any) error {
	reflectValue := reflect.ValueOf(value)
	kind := reflectValue.Kind()
	if kind != reflect.Map {
		return errors.NewError(
			fmt.Sprintf("expected type map for parameter %s but got %s", p.Name, kind),
			"ensure parameter is map[string]string",
		)
	}

	if m, ok := value.(map[string]any); ok {
		for _, v := range m {
			if k, ok := v.(string); !ok {
				return errors.NewError(
					fmt.Sprintf("expected type string for value for key %s in map", k),
					fmt.Sprintf("ensure all values in map are strings"),
				)
			}
		}
	}
	return nil
}
