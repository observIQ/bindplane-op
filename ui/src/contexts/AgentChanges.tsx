import { createContext } from "react";
import { useAgentChangesSubscription } from "../graphql/generated";
import { AgentChangeItem } from '../hooks/useAgentChanges';

interface AgentChangesContextValue {
  agentChanges: AgentChangeItem[];
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
