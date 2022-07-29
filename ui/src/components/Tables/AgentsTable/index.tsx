import { gql } from "@apollo/client";
import {
  GridDensityTypes,
  GridRowParams,
  GridSelectionModel,
} from "@mui/x-data-grid";
import { debounce } from "lodash";
import { memo, useMemo, useState } from "react";
import {
  Agent,
  AgentChangeType,
  Suggestion,
  useAgentChangesSubscription,
} from "../../../graphql/generated";
import { SearchBar } from "../../SearchBar";
import {
  AgentsTableChange,
  AgentsDataGrid,
  AgentsTableField,
  AgentsTableRow,
} from "./AgentsDataGrid";

gql`
  query AgentsTable($selector: String, $query: String) {
    agents(selector: $selector, query: $query) {
      agents {
        id
        name
        status
        architecture
        labels
        platform

        operatingSystem

        connectedAt
        disconnectedAt

        configurationResource {
          metadata {
            name
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

function applyAgentChanges(
  changes: AgentsTableChange[],
  agents: AgentsTableRow[]
): AgentsTableRow[] {
  // make a map of id => agent
  const map: { [id: string]: AgentsTableRow } = {};
  for (const agent of agents) {
    map[agent.id] = agent;
  }

  // changes includes inserts, updates, and deletes
  for (const change of changes) {
    const agent = change.agent;
    switch (change.changeType) {
      case AgentChangeType.Remove:
        delete map[agent.id];
        break;
      default:
        // update and insert are the same
        map[agent.id] = agent;
        break;
    }
  }
  return Object.values(map);
}

let total = 0;

interface AgentsTableData {
  agents: AgentsTableRow[];
  suggestions?: Suggestion[];
  query: string;
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
  // const { data, loading, refetch, subscribeToMore } = useAgentsTableQuery({
  //   variables: { selector, query: initQuery },
  //   fetchPolicy: "network-only",
  //   nextFetchPolicy: "cache-only",
  // });

  const [data, setData] = useState<AgentsTableData>({
    agents: [],
    suggestions: [],
    query: "",
  });
  const [subQuery, setSubQuery] = useState<string>(initQuery);

  const { loading } = useAgentChangesSubscription({
    variables: { selector, query: subQuery, seed: true },
    fetchPolicy: "network-only",
    onSubscriptionData(options) {
      const agentChanges = options.subscriptionData.data?.agentChanges;
      if (agentChanges == null) {
        setData({
          agents: [],
          suggestions: [],
          query: "",
        });
        return;
      }

      // keep track of the amount of data loaded for now
      const size = JSON.stringify(agentChanges).length;
      total += size;
      console.log(`${size} bytes => ${total / 1000} k total`);

      const { query, agentChanges: changes, suggestions } = agentChanges;
      if (changes != null) {
        if (query === data.query) {
          // query is the same, accumulate results
          setData({
            agents: applyAgentChanges(changes, data.agents),
            suggestions: data.suggestions,
            query: data.query,
          });
        } else {
          // query changed, start over
          setData({
            agents: applyAgentChanges(changes, []),
            suggestions: suggestions || [],
            query: query || "",
          });
        }
      }
    },
  });

  // const debouncedRefetch = useMemo(() => debounce(refetch, 100), [refetch]);

  const filterOptions: Suggestion[] = [
    { label: "Disconnected agents", query: "status:disconnected" },
    { label: "Outdated agents", query: "-version:latest" },
    { label: "No managed configuration", query: "-configuration:" },
  ];

  // useEffect(() => {
  //   if (data?.agents.agents != null) {
  //     setAgents(data.agents.agents);
  //   }
  // }, [data?.agents.agents, setAgents]);

  // useEffect(() => {
  //   subscribeToMore({
  //     document: AgentChangesDocument,
  //     variables: { query: subQuery, selector },
  //     updateQuery: (prev, { subscriptionData, variables }) => {
  //       if (
  //         subscriptionData == null ||
  //         variables?.query !== subQuery ||
  //         variables.selector !== selector
  //       ) {
  //         return prev;
  //       }

  //       const { data } = subscriptionData as unknown as {
  //         data: AgentChangesSubscription;
  //       };

  //       return {
  //         agents: {
  //           __typename: "Agents",
  //           suggestions: prev.agents.suggestions,
  //           query: prev.agents.query,
  //           agents: mergeAgents(prev.agents.agents, data.agentChanges),
  //         },
  //       };
  //     },
  //   });
  // }, [selector, subQuery, subscribeToMore]);

  const debouncedSetSubQuery = useMemo(
    () => debounce(setSubQuery, 300),
    [setSubQuery]
  );

  function onQueryChange(query: string) {
    setData({
      agents: data.agents,
      suggestions: [],
      query: data.query,
    });
    debouncedSetSubQuery(query);
  }

  return (
    <>
      <SearchBar
        filterOptions={filterOptions}
        suggestions={data.suggestions}
        onQueryChange={onQueryChange}
        suggestionQuery={subQuery}
        initialQuery={initQuery}
      />

      <AgentsDataGrid
        isRowSelectable={isRowSelectable}
        onAgentsSelected={onAgentsSelected}
        density={density}
        minHeight={minHeight}
        loading={loading}
        agents={data.agents}
        columnFields={columnFields}
      />
    </>
  );
};

export const AgentsTable = memo(AgentsTableComponent);
