import { createContext } from "react";
import { AgentChange, useAgentChangesSubscription } from "../graphql/generated";

interface AgentChangesContextValue {
  agentChanges: AgentChange[];
}

export const AgentChangesContext = createContext<AgentChangesContextValue>({
  agentChanges: [],
});

export const AgentChangesProvider: React.FC = ({ children }) => {
  const { data } = useAgentChangesSubscription();
  return (
    <AgentChangesContext.Provider
      value={{ agentChanges: data?.agentChanges || [] }}
    >
      {children}
    </AgentChangesContext.Provider>
  );
};
