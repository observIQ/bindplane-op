import {
  GridDensityTypes,
  GridRowParams,
  GridSelectionModel,
} from "@mui/x-data-grid";
import { debounce } from "lodash";
import { memo, useMemo, useState } from "react";
import {
  AgentChangeType,
  Suggestion,
  useAgentChangesSubscription,
} from "../../../graphql/generated";
import { AgentChangeAgent, AgentChangeItem } from '../../../hooks/useAgentChanges';
import { SearchBar } from "../../SearchBar";
import {
  AgentsDataGrid,
  AgentsTableField,
} from "./AgentsDataGrid";

interface Props {
  onAgentsSelected?: (agentIds: GridSelectionModel) => void;
  isRowSelectable?: (params: GridRowParams<AgentChangeAgent>) => boolean;
  selector?: string;
  minHeight?: string;
  columnFields?: AgentsTableField[];
  density?: GridDensityTypes;
  initQuery?: string;
}

export function applyAgentChanges(
  agents: AgentChangeAgent[],
  changes: AgentChangeItem[],
): AgentChangeAgent[] {
  // make a map of id => agent
  const map: { [id: string]: AgentChangeAgent } = {};
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

interface AgentsTableData {
  agents: AgentChangeAgent[];
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

      const { query, agentChanges: changes, suggestions } = agentChanges;
      if (changes != null) {
        if (query === data.query) {
          // query is the same, accumulate results
          setData({
            agents: applyAgentChanges(data.agents, changes),
            suggestions: data.suggestions,
            query: data.query,
          });
        } else {
          // query changed, start over
          setData({
            agents: applyAgentChanges([], changes),
            suggestions: suggestions || [],
            query: query || "",
          });
        }
      }
    },
  });

  const filterOptions: Suggestion[] = [
    { label: "Disconnected agents", query: "status:disconnected" },
    { label: "Outdated agents", query: "-version:latest" },
    { label: "No managed configuration", query: "-configuration:" },
  ];

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
