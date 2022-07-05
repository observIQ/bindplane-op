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
	"strconv"

	"github.com/observiq/bindplane-op/model/validation"
	"github.com/observiq/stanza/errors"
)

const (
	stringType  = "string"
	boolType    = "bool"
	intType     = "int"
	stringsType = "strings"
	enumType    = "enum"
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
	case stringType, intType, boolType, stringsType, enumType: // ok
	default:
		return errors.NewError(
			fmt.Sprintf("invalid type '%s' for '%s'", p.Type, p.Name),
			"ensure that the type is one of 'string', 'int', 'bool', 'strings', or 'enum'",
		)
	}
	return nil
}

func (p ParameterDefinition) validateValidValues() error {
	switch p.Type {
	case stringType, intType, boolType, stringsType:
		if len(p.ValidValues) > 0 {
			return errors.NewError(
				fmt.Sprintf("validValues is undefined for parameter of type '%s'", p.Type),
				"remove 'validValues' field or change type to 'enum'",
			)
		}
	case enumType:
		if len(p.ValidValues) == 0 {
			return errors.NewError(
				"parameter of type 'enum' must have 'validValues' specified",
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
