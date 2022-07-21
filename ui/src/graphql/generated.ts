import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
const defaultOptions = {} as const;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string;
  String: string;
  Boolean: boolean;
  Int: number;
  Float: number;
  Any: any;
  Map: any;
  Time: any;
};

export type Agent = {
  __typename?: 'Agent';
  architecture?: Maybe<Scalars['String']>;
  configuration?: Maybe<AgentConfiguration>;
  configurationResource?: Maybe<Configuration>;
  connectedAt?: Maybe<Scalars['Time']>;
  disconnectedAt?: Maybe<Scalars['Time']>;
  errorMessage?: Maybe<Scalars['String']>;
  home?: Maybe<Scalars['String']>;
  hostName?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  labels?: Maybe<Scalars['Map']>;
  macAddress?: Maybe<Scalars['String']>;
  name: Scalars['String'];
  operatingSystem?: Maybe<Scalars['String']>;
  platform?: Maybe<Scalars['String']>;
  remoteAddress?: Maybe<Scalars['String']>;
  status: Scalars['Int'];
  type?: Maybe<Scalars['String']>;
  version?: Maybe<Scalars['String']>;
};

export type AgentChange = {
  __typename?: 'AgentChange';
  agent: Agent;
  changeType: AgentChangeType;
};

export enum AgentChangeType {
  Insert = 'INSERT',
  Remove = 'REMOVE',
  Update = 'UPDATE'
}

export type AgentConfiguration = {
  __typename?: 'AgentConfiguration';
  Collector?: Maybe<Scalars['String']>;
  Logging?: Maybe<Scalars['String']>;
  Manager?: Maybe<Scalars['Map']>;
};

export type AgentSelector = {
  __typename?: 'AgentSelector';
  matchLabels?: Maybe<Scalars['Map']>;
};

export type Agents = {
  __typename?: 'Agents';
  agents: Array<Agent>;
  query?: Maybe<Scalars['String']>;
  suggestions?: Maybe<Array<Suggestion>>;
};

export type Components = {
  __typename?: 'Components';
  destinations: Array<Destination>;
  sources: Array<Source>;
};

export type Configuration = {
  __typename?: 'Configuration';
  apiVersion: Scalars['String'];
  kind: Scalars['String'];
  metadata: Metadata;
  spec: ConfigurationSpec;
};

export type ConfigurationChange = {
  __typename?: 'ConfigurationChange';
  configuration: Configuration;
  eventType: EventType;
};

export type ConfigurationSpec = {
  __typename?: 'ConfigurationSpec';
  contentType?: Maybe<Scalars['String']>;
  destinations?: Maybe<Array<ResourceConfiguration>>;
  raw?: Maybe<Scalars['String']>;
  selector?: Maybe<AgentSelector>;
  sources?: Maybe<Array<ResourceConfiguration>>;
};

export type Configurations = {
  __typename?: 'Configurations';
  configurations: Array<Configuration>;
  query?: Maybe<Scalars['String']>;
  suggestions?: Maybe<Array<Suggestion>>;
};

export type Destination = {
  __typename?: 'Destination';
  apiVersion: Scalars['String'];
  kind: Scalars['String'];
  metadata: Metadata;
  spec: ParameterizedSpec;
};

export type DestinationType = {
  __typename?: 'DestinationType';
  apiVersion: Scalars['String'];
  kind: Scalars['String'];
  metadata: Metadata;
  spec: ResourceTypeSpec;
};

export type DestinationWithType = {
  __typename?: 'DestinationWithType';
  destination?: Maybe<Destination>;
  destinationType?: Maybe<DestinationType>;
};

export enum EventType {
  Insert = 'INSERT',
  Remove = 'REMOVE',
  Update = 'UPDATE'
}

export type LiveTailMessage = {
  __typename?: 'LiveTailMessage';
  records: Array<Scalars['Any']>;
  type?: Maybe<LiveTailRecordType>;
};

export enum LiveTailRecordType {
  Log = 'log',
  Metric = 'metric',
  Trace = 'trace'
}

export type Metadata = {
  __typename?: 'Metadata';
  description?: Maybe<Scalars['String']>;
  displayName?: Maybe<Scalars['String']>;
  icon?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  labels?: Maybe<Scalars['Map']>;
  name: Scalars['String'];
};

export type Parameter = {
  __typename?: 'Parameter';
  name: Scalars['String'];
  value: Scalars['Any'];
};

export type ParameterDefinition = {
  __typename?: 'ParameterDefinition';
  default?: Maybe<Scalars['Any']>;
  description: Scalars['String'];
  label: Scalars['String'];
  name: Scalars['String'];
  relevantIf?: Maybe<Array<RelevantIfCondition>>;
  required: Scalars['Boolean'];
  type: ParameterType;
  validValues?: Maybe<Array<Scalars['String']>>;
};

export enum ParameterType {
  Bool = 'bool',
  Enum = 'enum',
  Int = 'int',
  String = 'string',
  Strings = 'strings'
}

export type ParameterizedSpec = {
  __typename?: 'ParameterizedSpec';
  parameters?: Maybe<Array<Parameter>>;
  type: Scalars['String'];
};

export enum PipelineType {
  Logs = 'logs',
  Metrics = 'metrics',
  Traces = 'traces'
}

export type Processor = {
  __typename?: 'Processor';
  apiVersion: Scalars['String'];
  kind: Scalars['String'];
  metadata: Metadata;
  spec: ParameterizedSpec;
};

export type ProcessorType = {
  __typename?: 'ProcessorType';
  apiVersion: Scalars['String'];
  kind: Scalars['String'];
  metadata: Metadata;
  spec: ResourceTypeSpec;
};

export type Query = {
  __typename?: 'Query';
  agent?: Maybe<Agent>;
  agents: Agents;
  components: Components;
  configuration?: Maybe<Configuration>;
  configurations: Configurations;
  destination?: Maybe<Destination>;
  destinationType?: Maybe<DestinationType>;
  destinationTypes: Array<DestinationType>;
  destinationWithType: DestinationWithType;
  destinations: Array<Destination>;
  processor?: Maybe<Processor>;
  processorType?: Maybe<ProcessorType>;
  processorTypes: Array<ProcessorType>;
  processors: Array<Processor>;
  source?: Maybe<Source>;
  sourceType?: Maybe<SourceType>;
  sourceTypes: Array<SourceType>;
  sources: Array<Source>;
};


export type QueryAgentArgs = {
  id: Scalars['ID'];
};


export type QueryAgentsArgs = {
  query?: InputMaybe<Scalars['String']>;
  selector?: InputMaybe<Scalars['String']>;
};


export type QueryConfigurationArgs = {
  name: Scalars['String'];
};


export type QueryConfigurationsArgs = {
  query?: InputMaybe<Scalars['String']>;
  selector?: InputMaybe<Scalars['String']>;
};


export type QueryDestinationArgs = {
  name: Scalars['String'];
};


export type QueryDestinationTypeArgs = {
  name: Scalars['String'];
};


export type QueryDestinationWithTypeArgs = {
  name: Scalars['String'];
};


export type QueryProcessorArgs = {
  name: Scalars['String'];
};


export type QueryProcessorTypeArgs = {
  name: Scalars['String'];
};


export type QuerySourceArgs = {
  name: Scalars['String'];
};


export type QuerySourceTypeArgs = {
  name: Scalars['String'];
};

export type RelevantIfCondition = {
  __typename?: 'RelevantIfCondition';
  name: Scalars['String'];
  operator: RelevantIfOperatorType;
  value: Scalars['Any'];
};

export enum RelevantIfOperatorType {
  Equals = 'equals'
}

export type ResourceConfiguration = {
  __typename?: 'ResourceConfiguration';
  name?: Maybe<Scalars['String']>;
  parameters?: Maybe<Array<Parameter>>;
  processors?: Maybe<Array<ResourceConfiguration>>;
  type?: Maybe<Scalars['String']>;
};

export type ResourceTypeSpec = {
  __typename?: 'ResourceTypeSpec';
  parameters: Array<ParameterDefinition>;
  supportedPlatforms: Array<Scalars['String']>;
  telemetryTypes: Array<PipelineType>;
  version: Scalars['String'];
};

export type Source = {
  __typename?: 'Source';
  apiVersion: Scalars['String'];
  kind: Scalars['String'];
  metadata: Metadata;
  spec: ParameterizedSpec;
};

export type SourceType = {
  __typename?: 'SourceType';
  apiVersion: Scalars['String'];
  kind: Scalars['String'];
  metadata: Metadata;
  spec: ResourceTypeSpec;
};

export type Subscription = {
  __typename?: 'Subscription';
  agentChanges: Array<AgentChange>;
  configurationChanges: Array<ConfigurationChange>;
  livetail: Array<LiveTailMessage>;
};


export type SubscriptionAgentChangesArgs = {
  query?: InputMaybe<Scalars['String']>;
  selector?: InputMaybe<Scalars['String']>;
};


export type SubscriptionConfigurationChangesArgs = {
  query?: InputMaybe<Scalars['String']>;
  selector?: InputMaybe<Scalars['String']>;
};


export type SubscriptionLivetailArgs = {
  agentIds: Array<Scalars['String']>;
  filters: Array<Scalars['String']>;
};

export type Suggestion = {
  __typename?: 'Suggestion';
  label: Scalars['String'];
  query: Scalars['String'];
};

export type AgentsTableQueryVariables = Exact<{
  selector?: InputMaybe<Scalars['String']>;
  query?: InputMaybe<Scalars['String']>;
}>;


export type AgentsTableQuery = { __typename?: 'Query', agents: { __typename?: 'Agents', query?: string | null, agents: Array<{ __typename?: 'Agent', id: string, architecture?: string | null, hostName?: string | null, labels?: any | null, platform?: string | null, version?: string | null, name: string, home?: string | null, operatingSystem?: string | null, macAddress?: string | null, type?: string | null, status: number, connectedAt?: any | null, disconnectedAt?: any | null, configurationResource?: { __typename?: 'Configuration', apiVersion: string, kind: string, metadata: { __typename?: 'Metadata', id: string, name: string }, spec: { __typename?: 'ConfigurationSpec', contentType?: string | null } } | null }>, suggestions?: Array<{ __typename?: 'Suggestion', query: string, label: string }> | null } };

export type GetDestinationTypeDisplayInfoQueryVariables = Exact<{
  name: Scalars['String'];
}>;


export type GetDestinationTypeDisplayInfoQuery = { __typename?: 'Query', destinationType?: { __typename?: 'DestinationType', metadata: { __typename?: 'Metadata', displayName?: string | null, icon?: string | null, name: string } } | null };

export type GetSourceTypeDisplayInfoQueryVariables = Exact<{
  name: Scalars['String'];
}>;


export type GetSourceTypeDisplayInfoQuery = { __typename?: 'Query', sourceType?: { __typename?: 'SourceType', metadata: { __typename?: 'Metadata', displayName?: string | null, icon?: string | null, name: string } } | null };

export type ComponentsQueryVariables = Exact<{ [key: string]: never; }>;


export type ComponentsQuery = { __typename?: 'Query', sources: Array<{ __typename?: 'Source', kind: string, metadata: { __typename?: 'Metadata', name: string }, spec: { __typename?: 'ParameterizedSpec', type: string } }>, destinations: Array<{ __typename?: 'Destination', kind: string, metadata: { __typename?: 'Metadata', name: string }, spec: { __typename?: 'ParameterizedSpec', type: string } }> };

export type GetConfigurationTableQueryVariables = Exact<{
  selector?: InputMaybe<Scalars['String']>;
  query?: InputMaybe<Scalars['String']>;
}>;


export type GetConfigurationTableQuery = { __typename?: 'Query', configurations: { __typename?: 'Configurations', query?: string | null, configurations: Array<{ __typename?: 'Configuration', metadata: { __typename?: 'Metadata', name: string, labels?: any | null, description?: string | null } }>, suggestions?: Array<{ __typename?: 'Suggestion', query: string, label: string }> | null } };

export type ConfigurationChangesSubscriptionVariables = Exact<{
  selector?: InputMaybe<Scalars['String']>;
  query?: InputMaybe<Scalars['String']>;
}>;


export type ConfigurationChangesSubscription = { __typename?: 'Subscription', configurationChanges: Array<{ __typename?: 'ConfigurationChange', eventType: EventType, configuration: { __typename?: 'Configuration', metadata: { __typename?: 'Metadata', name: string, description?: string | null, labels?: any | null } } }> };

export type AgentChangesSubscriptionVariables = Exact<{
  selector?: InputMaybe<Scalars['String']>;
  query?: InputMaybe<Scalars['String']>;
}>;


export type AgentChangesSubscription = { __typename?: 'Subscription', agentChanges: Array<{ __typename?: 'AgentChange', changeType: AgentChangeType, agent: { __typename?: 'Agent', id: string, name: string, architecture?: string | null, operatingSystem?: string | null, labels?: any | null, hostName?: string | null, platform?: string | null, version?: string | null, macAddress?: string | null, home?: string | null, type?: string | null, status: number, connectedAt?: any | null, disconnectedAt?: any | null, configuration?: { __typename?: 'AgentConfiguration', Collector?: string | null } | null, configurationResource?: { __typename?: 'Configuration', apiVersion: string, kind: string, metadata: { __typename?: 'Metadata', id: string, name: string }, spec: { __typename?: 'ConfigurationSpec', contentType?: string | null } } | null } }> };

export type GetAgentAndConfigurationsQueryVariables = Exact<{
  agentId: Scalars['ID'];
}>;


export type GetAgentAndConfigurationsQuery = { __typename?: 'Query', agent?: { __typename?: 'Agent', id: string, name: string, architecture?: string | null, operatingSystem?: string | null, labels?: any | null, hostName?: string | null, platform?: string | null, version?: string | null, macAddress?: string | null, remoteAddress?: string | null, home?: string | null, status: number, connectedAt?: any | null, disconnectedAt?: any | null, errorMessage?: string | null, configuration?: { __typename?: 'AgentConfiguration', Collector?: string | null } | null, configurationResource?: { __typename?: 'Configuration', metadata: { __typename?: 'Metadata', name: string } } | null } | null, configurations: { __typename?: 'Configurations', configurations: Array<{ __typename?: 'Configuration', metadata: { __typename?: 'Metadata', name: string, labels?: any | null }, spec: { __typename?: 'ConfigurationSpec', raw?: string | null } }> } };

export type GetConfigurationNamesQueryVariables = Exact<{ [key: string]: never; }>;


export type GetConfigurationNamesQuery = { __typename?: 'Query', configurations: { __typename?: 'Configurations', configurations: Array<{ __typename?: 'Configuration', metadata: { __typename?: 'Metadata', name: string, labels?: any | null } }> } };

export type DestinationTypeQueryVariables = Exact<{
  name: Scalars['String'];
}>;


export type DestinationTypeQuery = { __typename?: 'Query', destinationType?: { __typename?: 'DestinationType', metadata: { __typename?: 'Metadata', displayName?: string | null, name: string, icon?: string | null, description?: string | null }, spec: { __typename?: 'ResourceTypeSpec', parameters: Array<{ __typename?: 'ParameterDefinition', label: string, name: string, description: string, required: boolean, type: ParameterType, default?: any | null, validValues?: Array<string> | null, relevantIf?: Array<{ __typename?: 'RelevantIfCondition', name: string, operator: RelevantIfOperatorType, value: any }> | null }> } } | null };

export type GetDestinationWithTypeQueryVariables = Exact<{
  name: Scalars['String'];
}>;


export type GetDestinationWithTypeQuery = { __typename?: 'Query', destinationWithType: { __typename?: 'DestinationWithType', destination?: { __typename?: 'Destination', metadata: { __typename?: 'Metadata', name: string, id: string, labels?: any | null }, spec: { __typename?: 'ParameterizedSpec', type: string, parameters?: Array<{ __typename?: 'Parameter', name: string, value: any }> | null } } | null, destinationType?: { __typename?: 'DestinationType', metadata: { __typename?: 'Metadata', name: string, icon?: string | null }, spec: { __typename?: 'ResourceTypeSpec', parameters: Array<{ __typename?: 'ParameterDefinition', label: string, name: string, description: string, required: boolean, type: ParameterType, default?: any | null, validValues?: Array<string> | null, relevantIf?: Array<{ __typename?: 'RelevantIfCondition', name: string, operator: RelevantIfOperatorType, value: any }> | null }> } } | null } };

export type SourceTypeQueryVariables = Exact<{
  name: Scalars['String'];
}>;


export type SourceTypeQuery = { __typename?: 'Query', sourceType?: { __typename?: 'SourceType', metadata: { __typename?: 'Metadata', name: string, icon?: string | null, displayName?: string | null } } | null };

export type GetConfigurationQueryVariables = Exact<{
  name: Scalars['String'];
}>;


export type GetConfigurationQuery = { __typename?: 'Query', configuration?: { __typename?: 'Configuration', metadata: { __typename?: 'Metadata', id: string, name: string, description?: string | null, labels?: any | null }, spec: { __typename?: 'ConfigurationSpec', raw?: string | null, sources?: Array<{ __typename?: 'ResourceConfiguration', type?: string | null, name?: string | null, parameters?: Array<{ __typename?: 'Parameter', name: string, value: any }> | null }> | null, destinations?: Array<{ __typename?: 'ResourceConfiguration', type?: string | null, name?: string | null, parameters?: Array<{ __typename?: 'Parameter', name: string, value: any }> | null }> | null, selector?: { __typename?: 'AgentSelector', matchLabels?: any | null } | null } } | null };

export type DestinationsAndTypesQueryVariables = Exact<{ [key: string]: never; }>;


export type DestinationsAndTypesQuery = { __typename?: 'Query', destinationTypes: Array<{ __typename?: 'DestinationType', kind: string, apiVersion: string, metadata: { __typename?: 'Metadata', id: string, name: string, displayName?: string | null, description?: string | null, icon?: string | null }, spec: { __typename?: 'ResourceTypeSpec', version: string, supportedPlatforms: Array<string>, telemetryTypes: Array<PipelineType>, parameters: Array<{ __typename?: 'ParameterDefinition', label: string, type: ParameterType, name: string, description: string, default?: any | null, validValues?: Array<string> | null, required: boolean, relevantIf?: Array<{ __typename?: 'RelevantIfCondition', name: string, value: any, operator: RelevantIfOperatorType }> | null }> } }>, destinations: Array<{ __typename?: 'Destination', metadata: { __typename?: 'Metadata', name: string }, spec: { __typename?: 'ParameterizedSpec', type: string, parameters?: Array<{ __typename?: 'Parameter', name: string, value: any }> | null } }> };

export type SourceTypesQueryVariables = Exact<{ [key: string]: never; }>;


export type SourceTypesQuery = { __typename?: 'Query', sourceTypes: Array<{ __typename?: 'SourceType', apiVersion: string, kind: string, metadata: { __typename?: 'Metadata', id: string, name: string, displayName?: string | null, description?: string | null, icon?: string | null }, spec: { __typename?: 'ResourceTypeSpec', supportedPlatforms: Array<string>, version: string, telemetryTypes: Array<PipelineType>, parameters: Array<{ __typename?: 'ParameterDefinition', name: string, label: string, description: string, required: boolean, type: ParameterType, validValues?: Array<string> | null, default?: any | null, relevantIf?: Array<{ __typename?: 'RelevantIfCondition', name: string, operator: RelevantIfOperatorType, value: any }> | null }> } }> };

export type GetConfigNamesQueryVariables = Exact<{ [key: string]: never; }>;


export type GetConfigNamesQuery = { __typename?: 'Query', configurations: { __typename?: 'Configurations', configurations: Array<{ __typename?: 'Configuration', metadata: { __typename?: 'Metadata', name: string } }> } };


export const AgentsTableDocument = gql`
    query AgentsTable($selector: String, $query: String) {
  agents(selector: $selector, query: $query) {
    agents {
      id
      architecture
      hostName
      labels
      platform
      version
      name
      home
      operatingSystem
      macAddress
      type
      status
      connectedAt
      disconnectedAt
      configurationResource {
        apiVersion
        kind
        metadata {
          id
          name
        }
        spec {
          contentType
        }
      }
    }
    query
    suggestions {
      query
      label
    }
  }
}
    `;

/**
 * __useAgentsTableQuery__
 *
 * To run a query within a React component, call `useAgentsTableQuery` and pass it any options that fit your needs.
 * When your component renders, `useAgentsTableQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAgentsTableQuery({
 *   variables: {
 *      selector: // value for 'selector'
 *      query: // value for 'query'
 *   },
 * });
 */
export function useAgentsTableQuery(baseOptions?: Apollo.QueryHookOptions<AgentsTableQuery, AgentsTableQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AgentsTableQuery, AgentsTableQueryVariables>(AgentsTableDocument, options);
      }
export function useAgentsTableLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AgentsTableQuery, AgentsTableQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AgentsTableQuery, AgentsTableQueryVariables>(AgentsTableDocument, options);
        }
export type AgentsTableQueryHookResult = ReturnType<typeof useAgentsTableQuery>;
export type AgentsTableLazyQueryHookResult = ReturnType<typeof useAgentsTableLazyQuery>;
export type AgentsTableQueryResult = Apollo.QueryResult<AgentsTableQuery, AgentsTableQueryVariables>;
export const GetDestinationTypeDisplayInfoDocument = gql`
    query getDestinationTypeDisplayInfo($name: String!) {
  destinationType(name: $name) {
    metadata {
      displayName
      icon
      name
    }
  }
}
    `;

/**
 * __useGetDestinationTypeDisplayInfoQuery__
 *
 * To run a query within a React component, call `useGetDestinationTypeDisplayInfoQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetDestinationTypeDisplayInfoQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetDestinationTypeDisplayInfoQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetDestinationTypeDisplayInfoQuery(baseOptions: Apollo.QueryHookOptions<GetDestinationTypeDisplayInfoQuery, GetDestinationTypeDisplayInfoQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetDestinationTypeDisplayInfoQuery, GetDestinationTypeDisplayInfoQueryVariables>(GetDestinationTypeDisplayInfoDocument, options);
      }
export function useGetDestinationTypeDisplayInfoLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetDestinationTypeDisplayInfoQuery, GetDestinationTypeDisplayInfoQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetDestinationTypeDisplayInfoQuery, GetDestinationTypeDisplayInfoQueryVariables>(GetDestinationTypeDisplayInfoDocument, options);
        }
export type GetDestinationTypeDisplayInfoQueryHookResult = ReturnType<typeof useGetDestinationTypeDisplayInfoQuery>;
export type GetDestinationTypeDisplayInfoLazyQueryHookResult = ReturnType<typeof useGetDestinationTypeDisplayInfoLazyQuery>;
export type GetDestinationTypeDisplayInfoQueryResult = Apollo.QueryResult<GetDestinationTypeDisplayInfoQuery, GetDestinationTypeDisplayInfoQueryVariables>;
export const GetSourceTypeDisplayInfoDocument = gql`
    query getSourceTypeDisplayInfo($name: String!) {
  sourceType(name: $name) {
    metadata {
      displayName
      icon
      name
    }
  }
}
    `;

/**
 * __useGetSourceTypeDisplayInfoQuery__
 *
 * To run a query within a React component, call `useGetSourceTypeDisplayInfoQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetSourceTypeDisplayInfoQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetSourceTypeDisplayInfoQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetSourceTypeDisplayInfoQuery(baseOptions: Apollo.QueryHookOptions<GetSourceTypeDisplayInfoQuery, GetSourceTypeDisplayInfoQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetSourceTypeDisplayInfoQuery, GetSourceTypeDisplayInfoQueryVariables>(GetSourceTypeDisplayInfoDocument, options);
      }
export function useGetSourceTypeDisplayInfoLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetSourceTypeDisplayInfoQuery, GetSourceTypeDisplayInfoQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetSourceTypeDisplayInfoQuery, GetSourceTypeDisplayInfoQueryVariables>(GetSourceTypeDisplayInfoDocument, options);
        }
export type GetSourceTypeDisplayInfoQueryHookResult = ReturnType<typeof useGetSourceTypeDisplayInfoQuery>;
export type GetSourceTypeDisplayInfoLazyQueryHookResult = ReturnType<typeof useGetSourceTypeDisplayInfoLazyQuery>;
export type GetSourceTypeDisplayInfoQueryResult = Apollo.QueryResult<GetSourceTypeDisplayInfoQuery, GetSourceTypeDisplayInfoQueryVariables>;
export const ComponentsDocument = gql`
    query Components {
  sources {
    kind
    metadata {
      name
    }
    spec {
      type
    }
  }
  destinations {
    kind
    metadata {
      name
    }
    spec {
      type
    }
  }
}
    `;

/**
 * __useComponentsQuery__
 *
 * To run a query within a React component, call `useComponentsQuery` and pass it any options that fit your needs.
 * When your component renders, `useComponentsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useComponentsQuery({
 *   variables: {
 *   },
 * });
 */
export function useComponentsQuery(baseOptions?: Apollo.QueryHookOptions<ComponentsQuery, ComponentsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ComponentsQuery, ComponentsQueryVariables>(ComponentsDocument, options);
      }
export function useComponentsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ComponentsQuery, ComponentsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ComponentsQuery, ComponentsQueryVariables>(ComponentsDocument, options);
        }
export type ComponentsQueryHookResult = ReturnType<typeof useComponentsQuery>;
export type ComponentsLazyQueryHookResult = ReturnType<typeof useComponentsLazyQuery>;
export type ComponentsQueryResult = Apollo.QueryResult<ComponentsQuery, ComponentsQueryVariables>;
export const GetConfigurationTableDocument = gql`
    query GetConfigurationTable($selector: String, $query: String) {
  configurations(selector: $selector, query: $query) {
    configurations {
      metadata {
        name
        labels
        description
      }
    }
    query
    suggestions {
      query
      label
    }
  }
}
    `;

/**
 * __useGetConfigurationTableQuery__
 *
 * To run a query within a React component, call `useGetConfigurationTableQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetConfigurationTableQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetConfigurationTableQuery({
 *   variables: {
 *      selector: // value for 'selector'
 *      query: // value for 'query'
 *   },
 * });
 */
export function useGetConfigurationTableQuery(baseOptions?: Apollo.QueryHookOptions<GetConfigurationTableQuery, GetConfigurationTableQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetConfigurationTableQuery, GetConfigurationTableQueryVariables>(GetConfigurationTableDocument, options);
      }
export function useGetConfigurationTableLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetConfigurationTableQuery, GetConfigurationTableQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetConfigurationTableQuery, GetConfigurationTableQueryVariables>(GetConfigurationTableDocument, options);
        }
export type GetConfigurationTableQueryHookResult = ReturnType<typeof useGetConfigurationTableQuery>;
export type GetConfigurationTableLazyQueryHookResult = ReturnType<typeof useGetConfigurationTableLazyQuery>;
export type GetConfigurationTableQueryResult = Apollo.QueryResult<GetConfigurationTableQuery, GetConfigurationTableQueryVariables>;
export const ConfigurationChangesDocument = gql`
    subscription ConfigurationChanges($selector: String, $query: String) {
  configurationChanges(selector: $selector, query: $query) {
    configuration {
      metadata {
        name
        description
        labels
      }
    }
    eventType
  }
}
    `;

/**
 * __useConfigurationChangesSubscription__
 *
 * To run a query within a React component, call `useConfigurationChangesSubscription` and pass it any options that fit your needs.
 * When your component renders, `useConfigurationChangesSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useConfigurationChangesSubscription({
 *   variables: {
 *      selector: // value for 'selector'
 *      query: // value for 'query'
 *   },
 * });
 */
export function useConfigurationChangesSubscription(baseOptions?: Apollo.SubscriptionHookOptions<ConfigurationChangesSubscription, ConfigurationChangesSubscriptionVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<ConfigurationChangesSubscription, ConfigurationChangesSubscriptionVariables>(ConfigurationChangesDocument, options);
      }
export type ConfigurationChangesSubscriptionHookResult = ReturnType<typeof useConfigurationChangesSubscription>;
export type ConfigurationChangesSubscriptionResult = Apollo.SubscriptionResult<ConfigurationChangesSubscription>;
export const AgentChangesDocument = gql`
    subscription AgentChanges($selector: String, $query: String) {
  agentChanges(selector: $selector, query: $query) {
    agent {
      id
      name
      architecture
      operatingSystem
      labels
      hostName
      platform
      version
      macAddress
      home
      type
      status
      connectedAt
      disconnectedAt
      configuration {
        Collector
      }
      configurationResource {
        apiVersion
        kind
        metadata {
          id
          name
        }
        spec {
          contentType
        }
      }
    }
    changeType
  }
}
    `;

/**
 * __useAgentChangesSubscription__
 *
 * To run a query within a React component, call `useAgentChangesSubscription` and pass it any options that fit your needs.
 * When your component renders, `useAgentChangesSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAgentChangesSubscription({
 *   variables: {
 *      selector: // value for 'selector'
 *      query: // value for 'query'
 *   },
 * });
 */
export function useAgentChangesSubscription(baseOptions?: Apollo.SubscriptionHookOptions<AgentChangesSubscription, AgentChangesSubscriptionVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<AgentChangesSubscription, AgentChangesSubscriptionVariables>(AgentChangesDocument, options);
      }
export type AgentChangesSubscriptionHookResult = ReturnType<typeof useAgentChangesSubscription>;
export type AgentChangesSubscriptionResult = Apollo.SubscriptionResult<AgentChangesSubscription>;
export const GetAgentAndConfigurationsDocument = gql`
    query GetAgentAndConfigurations($agentId: ID!) {
  agent(id: $agentId) {
    id
    name
    architecture
    operatingSystem
    labels
    hostName
    platform
    version
    macAddress
    remoteAddress
    home
    status
    connectedAt
    disconnectedAt
    errorMessage
    configuration {
      Collector
    }
    configurationResource {
      metadata {
        name
      }
    }
  }
  configurations {
    configurations {
      metadata {
        name
        labels
      }
      spec {
        raw
      }
    }
  }
}
    `;

/**
 * __useGetAgentAndConfigurationsQuery__
 *
 * To run a query within a React component, call `useGetAgentAndConfigurationsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetAgentAndConfigurationsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetAgentAndConfigurationsQuery({
 *   variables: {
 *      agentId: // value for 'agentId'
 *   },
 * });
 */
export function useGetAgentAndConfigurationsQuery(baseOptions: Apollo.QueryHookOptions<GetAgentAndConfigurationsQuery, GetAgentAndConfigurationsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetAgentAndConfigurationsQuery, GetAgentAndConfigurationsQueryVariables>(GetAgentAndConfigurationsDocument, options);
      }
export function useGetAgentAndConfigurationsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetAgentAndConfigurationsQuery, GetAgentAndConfigurationsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetAgentAndConfigurationsQuery, GetAgentAndConfigurationsQueryVariables>(GetAgentAndConfigurationsDocument, options);
        }
export type GetAgentAndConfigurationsQueryHookResult = ReturnType<typeof useGetAgentAndConfigurationsQuery>;
export type GetAgentAndConfigurationsLazyQueryHookResult = ReturnType<typeof useGetAgentAndConfigurationsLazyQuery>;
export type GetAgentAndConfigurationsQueryResult = Apollo.QueryResult<GetAgentAndConfigurationsQuery, GetAgentAndConfigurationsQueryVariables>;
export const GetConfigurationNamesDocument = gql`
    query GetConfigurationNames {
  configurations {
    configurations {
      metadata {
        name
        labels
      }
    }
  }
}
    `;

/**
 * __useGetConfigurationNamesQuery__
 *
 * To run a query within a React component, call `useGetConfigurationNamesQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetConfigurationNamesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetConfigurationNamesQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetConfigurationNamesQuery(baseOptions?: Apollo.QueryHookOptions<GetConfigurationNamesQuery, GetConfigurationNamesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetConfigurationNamesQuery, GetConfigurationNamesQueryVariables>(GetConfigurationNamesDocument, options);
      }
export function useGetConfigurationNamesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetConfigurationNamesQuery, GetConfigurationNamesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetConfigurationNamesQuery, GetConfigurationNamesQueryVariables>(GetConfigurationNamesDocument, options);
        }
export type GetConfigurationNamesQueryHookResult = ReturnType<typeof useGetConfigurationNamesQuery>;
export type GetConfigurationNamesLazyQueryHookResult = ReturnType<typeof useGetConfigurationNamesLazyQuery>;
export type GetConfigurationNamesQueryResult = Apollo.QueryResult<GetConfigurationNamesQuery, GetConfigurationNamesQueryVariables>;
export const DestinationTypeDocument = gql`
    query DestinationType($name: String!) {
  destinationType(name: $name) {
    metadata {
      displayName
      name
      icon
      displayName
      description
    }
    spec {
      parameters {
        label
        name
        description
        required
        type
        default
        relevantIf {
          name
          operator
          value
        }
        validValues
      }
    }
  }
}
    `;

/**
 * __useDestinationTypeQuery__
 *
 * To run a query within a React component, call `useDestinationTypeQuery` and pass it any options that fit your needs.
 * When your component renders, `useDestinationTypeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useDestinationTypeQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useDestinationTypeQuery(baseOptions: Apollo.QueryHookOptions<DestinationTypeQuery, DestinationTypeQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<DestinationTypeQuery, DestinationTypeQueryVariables>(DestinationTypeDocument, options);
      }
export function useDestinationTypeLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<DestinationTypeQuery, DestinationTypeQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<DestinationTypeQuery, DestinationTypeQueryVariables>(DestinationTypeDocument, options);
        }
export type DestinationTypeQueryHookResult = ReturnType<typeof useDestinationTypeQuery>;
export type DestinationTypeLazyQueryHookResult = ReturnType<typeof useDestinationTypeLazyQuery>;
export type DestinationTypeQueryResult = Apollo.QueryResult<DestinationTypeQuery, DestinationTypeQueryVariables>;
export const GetDestinationWithTypeDocument = gql`
    query getDestinationWithType($name: String!) {
  destinationWithType(name: $name) {
    destination {
      metadata {
        name
        id
        labels
      }
      spec {
        type
        parameters {
          name
          value
        }
      }
    }
    destinationType {
      metadata {
        name
        icon
      }
      spec {
        parameters {
          label
          name
          description
          required
          type
          default
          relevantIf {
            name
            operator
            value
          }
          validValues
        }
      }
    }
  }
}
    `;

/**
 * __useGetDestinationWithTypeQuery__
 *
 * To run a query within a React component, call `useGetDestinationWithTypeQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetDestinationWithTypeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetDestinationWithTypeQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetDestinationWithTypeQuery(baseOptions: Apollo.QueryHookOptions<GetDestinationWithTypeQuery, GetDestinationWithTypeQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetDestinationWithTypeQuery, GetDestinationWithTypeQueryVariables>(GetDestinationWithTypeDocument, options);
      }
export function useGetDestinationWithTypeLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetDestinationWithTypeQuery, GetDestinationWithTypeQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetDestinationWithTypeQuery, GetDestinationWithTypeQueryVariables>(GetDestinationWithTypeDocument, options);
        }
export type GetDestinationWithTypeQueryHookResult = ReturnType<typeof useGetDestinationWithTypeQuery>;
export type GetDestinationWithTypeLazyQueryHookResult = ReturnType<typeof useGetDestinationWithTypeLazyQuery>;
export type GetDestinationWithTypeQueryResult = Apollo.QueryResult<GetDestinationWithTypeQuery, GetDestinationWithTypeQueryVariables>;
export const SourceTypeDocument = gql`
    query SourceType($name: String!) {
  sourceType(name: $name) {
    metadata {
      name
      icon
      displayName
    }
  }
}
    `;

/**
 * __useSourceTypeQuery__
 *
 * To run a query within a React component, call `useSourceTypeQuery` and pass it any options that fit your needs.
 * When your component renders, `useSourceTypeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSourceTypeQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useSourceTypeQuery(baseOptions: Apollo.QueryHookOptions<SourceTypeQuery, SourceTypeQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<SourceTypeQuery, SourceTypeQueryVariables>(SourceTypeDocument, options);
      }
export function useSourceTypeLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<SourceTypeQuery, SourceTypeQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<SourceTypeQuery, SourceTypeQueryVariables>(SourceTypeDocument, options);
        }
export type SourceTypeQueryHookResult = ReturnType<typeof useSourceTypeQuery>;
export type SourceTypeLazyQueryHookResult = ReturnType<typeof useSourceTypeLazyQuery>;
export type SourceTypeQueryResult = Apollo.QueryResult<SourceTypeQuery, SourceTypeQueryVariables>;
export const GetConfigurationDocument = gql`
    query GetConfiguration($name: String!) {
  configuration(name: $name) {
    metadata {
      id
      name
      description
      labels
    }
    spec {
      raw
      sources {
        type
        name
        parameters {
          name
          value
        }
      }
      destinations {
        type
        name
        parameters {
          name
          value
        }
      }
      selector {
        matchLabels
      }
    }
  }
}
    `;

/**
 * __useGetConfigurationQuery__
 *
 * To run a query within a React component, call `useGetConfigurationQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetConfigurationQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetConfigurationQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetConfigurationQuery(baseOptions: Apollo.QueryHookOptions<GetConfigurationQuery, GetConfigurationQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetConfigurationQuery, GetConfigurationQueryVariables>(GetConfigurationDocument, options);
      }
export function useGetConfigurationLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetConfigurationQuery, GetConfigurationQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetConfigurationQuery, GetConfigurationQueryVariables>(GetConfigurationDocument, options);
        }
export type GetConfigurationQueryHookResult = ReturnType<typeof useGetConfigurationQuery>;
export type GetConfigurationLazyQueryHookResult = ReturnType<typeof useGetConfigurationLazyQuery>;
export type GetConfigurationQueryResult = Apollo.QueryResult<GetConfigurationQuery, GetConfigurationQueryVariables>;
export const DestinationsAndTypesDocument = gql`
    query DestinationsAndTypes {
  destinationTypes {
    kind
    apiVersion
    metadata {
      id
      name
      displayName
      description
      icon
    }
    spec {
      version
      parameters {
        label
        type
        name
        description
        default
        validValues
        relevantIf {
          name
          value
          operator
        }
        required
      }
      supportedPlatforms
      telemetryTypes
    }
  }
  destinations {
    metadata {
      name
    }
    spec {
      type
      parameters {
        name
        value
      }
    }
  }
}
    `;

/**
 * __useDestinationsAndTypesQuery__
 *
 * To run a query within a React component, call `useDestinationsAndTypesQuery` and pass it any options that fit your needs.
 * When your component renders, `useDestinationsAndTypesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useDestinationsAndTypesQuery({
 *   variables: {
 *   },
 * });
 */
export function useDestinationsAndTypesQuery(baseOptions?: Apollo.QueryHookOptions<DestinationsAndTypesQuery, DestinationsAndTypesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<DestinationsAndTypesQuery, DestinationsAndTypesQueryVariables>(DestinationsAndTypesDocument, options);
      }
export function useDestinationsAndTypesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<DestinationsAndTypesQuery, DestinationsAndTypesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<DestinationsAndTypesQuery, DestinationsAndTypesQueryVariables>(DestinationsAndTypesDocument, options);
        }
export type DestinationsAndTypesQueryHookResult = ReturnType<typeof useDestinationsAndTypesQuery>;
export type DestinationsAndTypesLazyQueryHookResult = ReturnType<typeof useDestinationsAndTypesLazyQuery>;
export type DestinationsAndTypesQueryResult = Apollo.QueryResult<DestinationsAndTypesQuery, DestinationsAndTypesQueryVariables>;
export const SourceTypesDocument = gql`
    query sourceTypes {
  sourceTypes {
    apiVersion
    kind
    metadata {
      id
      name
      displayName
      description
      icon
    }
    spec {
      parameters {
        name
        label
        description
        relevantIf {
          name
          operator
          value
        }
        required
        type
        validValues
        default
      }
      supportedPlatforms
      version
      telemetryTypes
    }
  }
}
    `;

/**
 * __useSourceTypesQuery__
 *
 * To run a query within a React component, call `useSourceTypesQuery` and pass it any options that fit your needs.
 * When your component renders, `useSourceTypesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSourceTypesQuery({
 *   variables: {
 *   },
 * });
 */
export function useSourceTypesQuery(baseOptions?: Apollo.QueryHookOptions<SourceTypesQuery, SourceTypesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<SourceTypesQuery, SourceTypesQueryVariables>(SourceTypesDocument, options);
      }
export function useSourceTypesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<SourceTypesQuery, SourceTypesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<SourceTypesQuery, SourceTypesQueryVariables>(SourceTypesDocument, options);
        }
export type SourceTypesQueryHookResult = ReturnType<typeof useSourceTypesQuery>;
export type SourceTypesLazyQueryHookResult = ReturnType<typeof useSourceTypesLazyQuery>;
export type SourceTypesQueryResult = Apollo.QueryResult<SourceTypesQuery, SourceTypesQueryVariables>;
export const GetConfigNamesDocument = gql`
    query getConfigNames {
  configurations {
    configurations {
      metadata {
        name
      }
    }
  }
}
    `;

/**
 * __useGetConfigNamesQuery__
 *
 * To run a query within a React component, call `useGetConfigNamesQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetConfigNamesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetConfigNamesQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetConfigNamesQuery(baseOptions?: Apollo.QueryHookOptions<GetConfigNamesQuery, GetConfigNamesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetConfigNamesQuery, GetConfigNamesQueryVariables>(GetConfigNamesDocument, options);
      }
export function useGetConfigNamesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetConfigNamesQuery, GetConfigNamesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetConfigNamesQuery, GetConfigNamesQueryVariables>(GetConfigNamesDocument, options);
        }
export type GetConfigNamesQueryHookResult = ReturnType<typeof useGetConfigNamesQuery>;
export type GetConfigNamesLazyQueryHookResult = ReturnType<typeof useGetConfigNamesLazyQuery>;
export type GetConfigNamesQueryResult = Apollo.QueryResult<GetConfigNamesQuery, GetConfigNamesQueryVariables>;