import { ComponentStory, ComponentMeta } from "@storybook/react";
import { ResourceDialog } from ".";
import {
  Destination1,
  Destination2,
  ResourceType1,
  ResourceType2,
  SupportsBoth,
  SupportsLogs,
  SupportsMetrics,
} from "../ResourceConfigForm/__test__/dummyResources";

export default {
  title: "Resource Dialog",
  component: ResourceDialog,
} as ComponentMeta<typeof ResourceDialog>;

const Template: ComponentStory<typeof ResourceDialog> = (args) => (
  <ResourceDialog {...args} />
);

export const Destination = Template.bind({});
export const DestinationWithExistingResources = Template.bind({});
export const Source = Template.bind({});

Destination.args = {
  open: true,
  resourceTypes: [ResourceType1, ResourceType2],
  title: "Title",
  kind: "destination",
};

DestinationWithExistingResources.args = {
  open: true,
  resourceTypes: [ResourceType1, ResourceType2],
  resources: [Destination1, Destination2],
  title: "Title",
  kind: "destination",
};

Source.args = {
  open: true,
  resourceTypes: [SupportsLogs, SupportsMetrics, SupportsBoth],
  resources: [],
  title: "Title",
  kind: "source",
};
