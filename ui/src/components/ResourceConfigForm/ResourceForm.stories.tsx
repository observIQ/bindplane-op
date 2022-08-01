import { ComponentStory, ComponentMeta } from "@storybook/react";
import { ResourceConfigForm } from ".";
import { ResourceType2, ResourceType1 } from "./__test__/dummyResources";

export default {
  title: "Resource Form",
  component: ResourceConfigForm,
} as ComponentMeta<typeof ResourceConfigForm>;

const Template: ComponentStory<typeof ResourceConfigForm> = (args) => (
  <div style={{ width: 400 }}>
    <ResourceConfigForm {...args} />
  </div>
);

export const Default = Template.bind({});
export const RelevantIf = Template.bind({});

Default.args = {
  title: ResourceType1.metadata.displayName!,
  description: ResourceType1.metadata.description!,
  parameterDefinitions: ResourceType1.spec.parameters,
};
RelevantIf.args = {
  title: ResourceType2.metadata.displayName!,
  description: ResourceType2.metadata.description!,
  parameterDefinitions: ResourceType2.spec.parameters,
};
