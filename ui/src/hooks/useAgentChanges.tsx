import { gql } from "@apollo/client";
import { useContext } from "react";
import { AgentsTableChange } from '../components/Tables/AgentsTable/AgentsDataGrid';
import { AgentChangesContext } from "../contexts/AgentChanges";

gql`
  subscription AgentChanges($selector: String, $query: String, $seed: Boolean) {
    agentChanges(selector: $selector, query: $query, seed: $seed) {
      agentChanges {
        agent {
          id
          name
          version
          operatingSystem
          labels
          platform

          status

          configurationResource {
            metadata {
              name
            }
          }
        }
        changeType
      }

      query

      suggestions {
        query
        label
      }
    }
  }
`;

export function useAgentChangesContext(): AgentsTableChange[] {
  const { agentChanges } = useContext(AgentChangesContext);
  return agentChanges;
}
