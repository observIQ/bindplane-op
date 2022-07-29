import { createContext } from "react";
import { AgentsTableChange } from '../components/Tables/AgentsTable/AgentsDataGrid';
import { useAgentChangesSubscription } from "../graphql/generated";

interface AgentChangesContextValue {
  agentChanges: AgentsTableChange[];
}

export const AgentChangesContext = createContext<AgentChangesContextValue>({
  agentChanges: [],
});

export const AgentChangesProvider: React.FC = ({ children }) => {
  const { data } = useAgentChangesSubscription();
  return (
    <AgentChangesContext.Provider
      value={{ agentChanges: data?.agentChanges.agentChanges || [] }}
    >
      {children}
    </AgentChangesContext.Provider>
  );
};
