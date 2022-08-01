import { gql } from "@apollo/client";
import { debounce } from "lodash";
import { memo, useEffect, useMemo, useState } from "react";
import {
  Agent,
  AgentChangesDocument,
  AgentChangesSubscription,
  Suggestion,
  useAgentsTableQuery,
} from "../../../graphql/generated";
import { SearchBar } from "../../SearchBar";
import { AgentsDataGrid, AgentsTableField } from "./AgentsDataGrid";
import {
  GridDensityTypes,
  GridRowParams,
  GridSelectionModel,
} from "@mui/x-data-grid";
import { mergeAgents } from "./merge-agents";

gql`
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

interface Props {
  onAgentsSelected?: (agentIds: GridSelectionModel) => void;
  isRowSelectable?: (params: GridRowParams<Agent>) => boolean;
  selector?: string;
  minHeight?: string;
  columnFields?: AgentsTableField[];
  density?: GridDensityTypes;
  initQuery?: string;
}

const AgentsTableComponent: React.FC<Props> = ({
  onAgentsSelected,
  isRowSelectable,
  selector,
  minHeight,
  columnFields,
  density = GridDensityTypes.Standard,
  initQuery = "",
}) => {
  const { data, loading, refetch, subscribeToMore } = useAgentsTableQuery({
    variables: { selector, query: initQuery },
    fetchPolicy: "network-only",
    nextFetchPolicy: "cache-only",
  });

  const [agents, setAgents] = useState<Agent[]>([]);
  const [subQuery, setSubQuery] = useState<string>(initQuery);

  const debouncedRefetch = useMemo(() => debounce(refetch, 100), [refetch]);

  const filterOptions: Suggestion[] = [
    { label: "Disconnected agents", query: "status:disconnected" },
    { label: "Outdated agents", query: "-version:latest" },
    { label: "No managed configuration", query: "-configuration:" },
  ];

  useEffect(() => {
    if (data?.agents.agents != null) {
      setAgents(data.agents.agents);
    }
  }, [data?.agents.agents, setAgents]);

  useEffect(() => {
    subscribeToMore({
      document: AgentChangesDocument,
      variables: { query: subQuery, selector },
      updateQuery: (prev, { subscriptionData, variables }) => {
        if (
          subscriptionData == null ||
          variables?.query !== subQuery ||
          variables.selector !== selector
        ) {
          return prev;
        }

        const { data } = subscriptionData as unknown as {
          data: AgentChangesSubscription;
        };

        return {
          agents: {
            __typename: "Agents",
            suggestions: prev.agents.suggestions,
            query: prev.agents.query,
            agents: mergeAgents(prev.agents.agents, data.agentChanges),
          },
        };
      },
    });
  }, [selector, subQuery, subscribeToMore]);

  function onQueryChange(query: string) {
    debouncedRefetch({ selector, query });
    setSubQuery(query);
  }

  return (
    <>
      <SearchBar
        filterOptions={filterOptions}
        suggestions={data?.agents.suggestions}
        onQueryChange={onQueryChange}
        suggestionQuery={data?.agents.query}
        initialQuery={initQuery}
      />

      <AgentsDataGrid
        isRowSelectable={isRowSelectable}
        onAgentsSelected={onAgentsSelected}
        density={density}
        minHeight={minHeight}
        loading={loading}
        agents={agents}
        columnFields={columnFields}
      />
    </>
  );
};

export const AgentsTable = memo(AgentsTableComponent);
