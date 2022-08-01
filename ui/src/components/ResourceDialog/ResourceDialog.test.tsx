import { render, screen } from "@testing-library/react";
import { ResourceDialog } from ".";
import {
  Destination1,
  ResourceType1,
  ResourceType2,
} from "../ResourceConfigForm/__test__/dummyResources";

describe("ResourceDialog", () => {
  it("renders without error", () => {
    render(
      <ResourceDialog
        resourceTypes={[ResourceType1, ResourceType2]}
        title={""}
        kind={"source"}
        open={true}
      />
    );
  });

  it("renders ResourceTypes", () => {
    render(
      <ResourceDialog
        resourceTypes={[ResourceType1, ResourceType2]}
        title={""}
        kind={"source"}
        open={true}
      />
    );

    screen.getByText(ResourceType1.metadata.displayName!);
    screen.getByText(ResourceType2.metadata.displayName!);
  });

  it("displays ResourceType form when clicking next", () => {
    render(
      <ResourceDialog
        resourceTypes={[ResourceType1, ResourceType2]}
        title={""}
        kind={"source"}
        open={true}
      />
    );

    screen.getByText("ResourceType One").click();
    screen.getByTestId("resource-form");
  });

  it("will offer to use an existing destination", () => {
    render(
      <ResourceDialog
        resourceTypes={[ResourceType1, ResourceType2]}
        resources={[Destination1]}
        title={""}
        kind={"destination"}
        open={true}
      />
    );

    screen.getByText(ResourceType1.metadata.displayName!).click();
    screen.getByText(Destination1.metadata.name);
    screen.getByText("Create New");
  });
});
