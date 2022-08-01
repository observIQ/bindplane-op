import { Agent, AgentChange, AgentChangeType, AgentConfiguration } from "../../../../graphql/generated";

function createAgent(): Agent {
  const now = new Date();
  const connectedAt = new Date(now.getTime() - 5 * 60 * 1000);
  const id = makeId();
  const configuration: AgentConfiguration = {};
  return {
    id: id,
    name: `test-agent-${id}`,
    status: 1,
    __typename: "Agent",
    architecture: "amd64",
    connectedAt: connectedAt.toString(),
    hostName: "host-name",
    configuration,
    labels: {},
    macAddress: "",
    operatingSystem: "Ubuntu",
    home: "/path/to/home",
    disconnectedAt: null,
    type: "otel",
    version: "3.0.0",
    platform: "linux",
  };
}

export function generateAgents(length: number): Agent[] {
  const agents: Agent[] = [];
  for (let i = 0; i < length; i++) {
    agents.push(createAgent());
  }

  return agents;
}

export function generateAgentChanges(length: number): AgentChange[] {
  const agents: AgentChange[] = [];
  for (let i = 0; i < length; i++) {
    agents.push({
      agent: createAgent(),
      changeType: AgentChangeType.Insert,
    });
  }

  return agents;
}

function makeId() {
  var result = "";
  var characters =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  var charactersLength = characters.length;
  for (var i = 0; i < 5; i++) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength));
  }
  return result;
}
