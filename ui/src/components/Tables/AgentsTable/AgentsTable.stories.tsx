import { ComponentMeta, ComponentStory } from "@storybook/react";
import { AgentsTable } from ".";
import {
  AgentChangesDocument, AgentChangesSubscription
} from "../../../graphql/generated";
import { AgentTable } from "../AgentTable";
import { generateAgentChanges } from "./__testutil__/generate-agents";

export default {
  title: "Agents Table",
  component: AgentsTable,
  argTypes: {
    density: {
      options: ["standard", "comfortable", "compact"],
      control: "radio",
    },
    columnFields: {
      options: [
        "name",
        "status",
        "version",
        "configuration",
        "operatingSystem",
        "labels",
      ],
      control: "multi-select",
    },
  },
} as ComponentMeta<typeof AgentTable>;

const Template: ComponentStory<typeof AgentsTable> = (args) => (
  <div style={{ width: "80vw", height: "500px" }}>
    <AgentsTable {...args} />
  </div>
);

export const Default = Template.bind({});
export const Selectable = Template.bind({});

const resultData: AgentChangesSubscription = {
  agentChanges: {
    agentChanges: generateAgentChanges(50),
    suggestions: [],
    query: "",
  },
};

const mockParams = {
  apolloClient: {
    mocks: [
      {
        request: {
          query: AgentChangesDocument,
          variables: {
            query: "",
          },
        },
        result: {
          data: resultData,
        },
      },
      {
        request: {
          query: AgentChangesDocument,
          variables: {
            query: "",
          },
        },
        result: {
          data: {
            agentChanges: [],
          },
        },
      },
    ],
  },
};

Default.args = {};
Default.parameters = mockParams;

Selectable.args = {
  onAgentsSelected: (agentIds) => console.log({ agentIds }),
};
Selectable.parameters = mockParams;
