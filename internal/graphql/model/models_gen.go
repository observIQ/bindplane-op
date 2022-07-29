// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"

	"github.com/observiq/bindplane-op/internal/store/search"
	"github.com/observiq/bindplane-op/model"
)

type AgentChange struct {
	Agent      *model.Agent    `json:"agent"`
	ChangeType AgentChangeType `json:"changeType"`
}

type AgentConfiguration struct {
	Collector *string                `json:"Collector"`
	Logging   *string                `json:"Logging"`
	Manager   map[string]interface{} `json:"Manager"`
}

type Agents struct {
	Query       *string              `json:"query"`
	Agents      []*model.Agent       `json:"agents"`
	Suggestions []*search.Suggestion `json:"suggestions"`
}

type Components struct {
	Sources      []*model.Source      `json:"sources"`
	Destinations []*model.Destination `json:"destinations"`
}

type ConfigurationChange struct {
	Configuration *model.Configuration `json:"configuration"`
	EventType     EventType            `json:"eventType"`
}

type Configurations struct {
	Query          *string                `json:"query"`
	Configurations []*model.Configuration `json:"configurations"`
	Suggestions    []*search.Suggestion   `json:"suggestions"`
}

type DestinationWithType struct {
	Destination     *model.Destination     `json:"destination"`
	DestinationType *model.DestinationType `json:"destinationType"`
}

type AgentChangeType string

const (
	AgentChangeTypeInsert AgentChangeType = "INSERT"
	AgentChangeTypeUpdate AgentChangeType = "UPDATE"
	AgentChangeTypeRemove AgentChangeType = "REMOVE"
)

var AllAgentChangeType = []AgentChangeType{
	AgentChangeTypeInsert,
	AgentChangeTypeUpdate,
	AgentChangeTypeRemove,
}

func (e AgentChangeType) IsValid() bool {
	switch e {
	case AgentChangeTypeInsert, AgentChangeTypeUpdate, AgentChangeTypeRemove:
		return true
	}
	return false
}

func (e AgentChangeType) String() string {
	return string(e)
}

func (e *AgentChangeType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = AgentChangeType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AgentChangeType", str)
	}
	return nil
}

func (e AgentChangeType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type EventType string

const (
	EventTypeInsert EventType = "INSERT"
	EventTypeUpdate EventType = "UPDATE"
	EventTypeRemove EventType = "REMOVE"
)

var AllEventType = []EventType{
	EventTypeInsert,
	EventTypeUpdate,
	EventTypeRemove,
}

func (e EventType) IsValid() bool {
	switch e {
	case EventTypeInsert, EventTypeUpdate, EventTypeRemove:
		return true
	}
	return false
}

func (e EventType) String() string {
	return string(e)
}

func (e *EventType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = EventType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid EventType", str)
	}
	return nil
}

func (e EventType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type ParameterType string

const (
	ParameterTypeString  ParameterType = "string"
	ParameterTypeStrings ParameterType = "strings"
	ParameterTypeInt     ParameterType = "int"
	ParameterTypeBool    ParameterType = "bool"
	ParameterTypeEnum    ParameterType = "enum"
	ParameterTypeEnums   ParameterType = "enums"
	ParameterTypeMap     ParameterType = "map"
	ParameterTypeYaml    ParameterType = "yaml"
)

var AllParameterType = []ParameterType{
	ParameterTypeString,
	ParameterTypeStrings,
	ParameterTypeInt,
	ParameterTypeBool,
	ParameterTypeEnum,
	ParameterTypeEnums,
	ParameterTypeMap,
	ParameterTypeYaml,
}

func (e ParameterType) IsValid() bool {
	switch e {
	case ParameterTypeString, ParameterTypeStrings, ParameterTypeInt, ParameterTypeBool, ParameterTypeEnum, ParameterTypeEnums, ParameterTypeMap, ParameterTypeYaml:
		return true
	}
	return false
}

func (e ParameterType) String() string {
	return string(e)
}

func (e *ParameterType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ParameterType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ParameterType", str)
	}
	return nil
}

func (e ParameterType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type RelevantIfOperatorType string

const (
	RelevantIfOperatorTypeEquals RelevantIfOperatorType = "equals"
)

var AllRelevantIfOperatorType = []RelevantIfOperatorType{
	RelevantIfOperatorTypeEquals,
}

func (e RelevantIfOperatorType) IsValid() bool {
	switch e {
	case RelevantIfOperatorTypeEquals:
		return true
	}
	return false
}

func (e RelevantIfOperatorType) String() string {
	return string(e)
}

func (e *RelevantIfOperatorType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = RelevantIfOperatorType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid RelevantIfOperatorType", str)
	}
	return nil
}

func (e RelevantIfOperatorType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
