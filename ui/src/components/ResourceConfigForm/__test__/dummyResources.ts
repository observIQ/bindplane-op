import {
  ParameterDefinition,
  ParameterType,
  RelevantIfOperatorType,
  SourceType,
  Destination,
  PipelineType,
} from "../../../graphql/generated";
import { APIVersion } from "../../../types/resources";

/* -------------------------- ParameterDefinitions -------------------------- */

export const stringDef: ParameterDefinition = {
  name: "string_name",
  label: "String Input",
  description: "Here is the description.",
  required: false,

  type: ParameterType.String,

  default: "default-value",
};

export const stringDefRequired: ParameterDefinition = {
  name: "string_required_name",
  label: "String Input",
  description: "Here is the description.",
  required: true,

  type: ParameterType.String,

  default: "default-required-value",
};

export const enumDef: ParameterDefinition = {
  name: "enum_name",
  label: "Enum Input",
  description: "Here is the description.",
  required: false,

  type: ParameterType.Enum,

  default: "option1",
  validValues: ["option1", "option2", "option3"],
};

export const stringsDef: ParameterDefinition = {
  name: "strings_name",
  label: "Multi String Input",
  description: "Here is the description.",
  required: false,

  type: ParameterType.Strings,

  default: ["option1", "option2"],
};

export const boolDef: ParameterDefinition = {
  name: "bool_name",
  label: "Bool Input",
  description: "Here is the description.",
  required: false,

  type: ParameterType.Bool,

  default: true,
};

export const boolDefaultFalseDef: ParameterDefinition = {
  name: "bool_default_false_name",
  label: "Bool Default False Input",
  description: "Here is the description.",
  required: false,

  type: ParameterType.Bool,

  default: false,
};

export const intDef: ParameterDefinition = {
  name: "int_name",
  label: "Int Input",
  description: "Here is the description.",
  required: false,

  type: ParameterType.Int,

  default: 25,
};

export const relevantIfDef: ParameterDefinition = {
  name: "string_name",
  label: "String Input",
  description: "Here is the description.",
  required: false,

  type: ParameterType.String,

  relevantIf: [
    {
      name: "bool_default_false_name",
      operator: RelevantIfOperatorType.Equals,
      value: true,
    },
  ],

  default: "default-value",
};

/* ----------------------------- Resource Types ----------------------------- */

export const ResourceType1: SourceType = {
  apiVersion: APIVersion.V1_BETA,
  kind: "ResourceType",
  metadata: {
    id: "resource-type-1",
    name: "resource-type-1",
    displayName: "ResourceType One",
    description: "A description for resource one.",
    icon: "/icons/destinations/otlp.svg",
  },
  spec: {
    version: "0.0.0",
    parameters: [
      stringDef,
      stringDefRequired,
      enumDef,
      stringsDef,
      boolDef,
      intDef,
    ],
    telemetryTypes: [],

    supportedPlatforms: ["linux", "macos", "windows"],
  },
};

export const ResourceType2: SourceType = {
  apiVersion: APIVersion.V1_BETA,
  kind: "ResourceType",
  metadata: {
    id: "resource-type-2",
    name: "resource-type-2",
    displayName: "ResourceType Two",
    description: "A description for resource one.",
    icon: "/icons/destinations/otlp.svg",
  },
  spec: {
    version: "0.0.0",
    parameters: [boolDefaultFalseDef, relevantIfDef],

    supportedPlatforms: ["linux", "macos", "windows"],
    telemetryTypes: [],
  },
};

export const SupportsLogs: SourceType = {
  apiVersion: APIVersion.V1_BETA,
  kind: "ResourceType",
  metadata: {
    id: "supports-logs",
    name: "supports-logs",
    displayName: "Supports Logs",
    description: "A resource that supports logs.",
    icon: "/icons/destinations/otlp.svg",
  },
  spec: {
    version: "0.0.0",
    parameters: [boolDefaultFalseDef, relevantIfDef],

    supportedPlatforms: ["linux", "macos", "windows"],
    telemetryTypes: [PipelineType.Logs],
  },
};

export const SupportsMetrics: SourceType = {
  apiVersion: APIVersion.V1_BETA,
  kind: "ResourceType",
  metadata: {
    id: "supports-metrics",
    name: "supports-metrics",
    displayName: "Supports Metrics",
    description: "A resource that supports metrics.",
    icon: "/icons/destinations/otlp.svg",
  },
  spec: {
    version: "0.0.0",
    parameters: [boolDefaultFalseDef, relevantIfDef],

    supportedPlatforms: ["linux", "macos", "windows"],
    telemetryTypes: [PipelineType.Metrics],
  },
};

export const SupportsBoth: SourceType = {
  apiVersion: APIVersion.V1_BETA,
  kind: "ResourceType",
  metadata: {
    id: "supports-logs-and-metrics",
    name: "supports-logs-and-metrics",
    displayName: "Supports Logs and Metrics",
    description: "A resource that supports logs and metrics.",
    icon: "/icons/destinations/otlp.svg",
  },
  spec: {
    version: "0.0.0",
    parameters: [boolDefaultFalseDef, relevantIfDef],

    supportedPlatforms: ["linux", "macos", "windows"],
    telemetryTypes: [PipelineType.Logs, PipelineType.Metrics],
  },
};

/* -------------------------------- Resources ------------------------------- */

// This destination is type resource-type-1
export const Destination1: Destination = {
  apiVersion: APIVersion.V1_BETA,
  kind: "Destination",
  metadata: {
    name: "destination-1-name",
    id: "destination-1-name",
  },
  spec: {
    parameters: [],
    type: "resource-type-1",
  },
};

// This destination is type resource-type-1
export const Destination2: Destination = {
  apiVersion: APIVersion.V1_BETA,
  kind: "Destination",
  metadata: {
    name: "destination-2-name",
    id: "destination-2-name",
  },
  spec: {
    parameters: [],
    type: "resource-type-1",
  },
};
