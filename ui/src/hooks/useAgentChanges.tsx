import { gql } from "@apollo/client";
import { useContext } from "react";
import { AgentChangesContext } from "../contexts/AgentChanges";
import { AgentChangesSubscription } from '../graphql/generated';

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

export type AgentChangeItem =  AgentChangesSubscription["agentChanges"]["agentChanges"][0];
export type AgentChangeAgent =
  AgentChangesSubscription["agentChanges"]["agentChanges"][0]["agent"];

export function useAgentChangesContext(): AgentChangeItem[] {
  const { agentChanges } = useContext(AgentChangesContext);
  return agentChanges;
}
