import { cloneDeep } from "@apollo/client/utilities";
import { applyAgentChanges } from '.';
import {
  AgentChangeType
} from "../../../graphql/generated";
import { AgentChangeAgent } from '../../../hooks/useAgentChanges';

const a1: AgentChangeAgent = {
  id: "1",
  name: "",
  status: 1,
};
const a2: AgentChangeAgent = {
  id: "2",
  name: "",
  status: 1,
};
const a3: AgentChangeAgent = {
  id: "3",
  name: "",
  status: 1,
};
const a4: AgentChangeAgent = {
  id: "4",
  name: "",
  status: 1,
};
const a5: AgentChangeAgent = {
  id: "5",
  name: "",
  status: 1,
};

describe("mergeAgents", () => {
  it("removes agents when it gets event type remove", () => {
    const updates = [
      {
        agent: a1,
        changeType: AgentChangeType.Remove,
      },
      {
        agent: a2,
        changeType: AgentChangeType.Remove,
      },
      {
        agent: a3,
        changeType: AgentChangeType.Remove,
      },
      {
        agent: a4,
        changeType: AgentChangeType.Remove,
      },
      {
        agent: a5,
        changeType: AgentChangeType.Remove,
      },
    ];

    const newAgents = applyAgentChanges([a1, a2, a3, a4, a5], updates);
    expect(newAgents).toEqual([]);
  });

  it("adds an agent with event type insert", () => {
    const current = [a1, a2, a3, a4];
    const updates = [{ agent: a5, changeType: AgentChangeType.Insert }];

    const merged = applyAgentChanges(current, updates);
    expect(merged).toEqual([a1, a2, a3, a4, a5]);
  });

  it("replaces an agent with event type update", () => {
    const a1Updated = cloneDeep(a1);
    a1Updated.status = 0;

    const current = [a1, a2];
    const updates = [
      {
        agent: a1Updated,
        changeType: AgentChangeType.Update,
      },
    ];

    const merged = applyAgentChanges(current, updates);
    expect(merged).toEqual([a1Updated, a2]);
  });

  it("will not re-insert an existing agent", () => {
    const current = [a1, a2];
    const updates = [
      {
        agent: a2,
        changeType: AgentChangeType.Insert,
      },
    ];

    const merged = applyAgentChanges(current, updates);
    expect(merged).toEqual([a1, a2]);
  });

  it("will insert an updated agent if not present", () => {
    const current = [a1, a2];
    const updates = [
      {
        agent: a3,
        changeType: AgentChangeType.Update,
      },
    ];

    const merged = applyAgentChanges(current, updates);
    expect(merged).toEqual([a1, a2, a3]);
  });
});
