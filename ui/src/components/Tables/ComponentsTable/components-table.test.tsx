import { GridSelectionModel } from "@mui/x-data-grid";
import { render, screen, waitFor } from "@testing-library/react";
import { resourcesFromSelected } from ".";
import {
  Destination1,
  Destination2,
} from "../../ResourceConfigForm/__test__/dummyResources";
import { ComponentsDataGrid } from "./ComponentsDataGrid";

describe("resourcesFromSelected", () => {
  it("Source|blah, Source|source-name, Destination|gcp", () => {
    const selected = ["Source|blah", "Source|source-name", "Destination|gcp"];

    const want = [
      {
        kind: "Source",
        metadata: {
          name: "blah",
        },
      },
      {
        kind: "Source",
        metadata: {
          name: "source-name",
        },
      },
      {
        kind: "Destination",
        metadata: {
          name: "gcp",
        },
      },
    ];

    const got = resourcesFromSelected(selected);

    expect(got).toEqual(want);
  });
});

describe("ComponentsDataGrid", () => {
  const destinationData = [Destination1, Destination2];

  it("renders without error", () => {
    render(
      <ComponentsDataGrid
        loading={false}
        queryData={{ destinations: destinationData, sources: [] }}
        onComponentsSelected={() => {}}
        disableSelectionOnClick
        checkboxSelection
        onEditDestination={() => {}}
      />
    );
  });

  it("displays destinations", () => {
    render(
      <ComponentsDataGrid
        loading={false}
        queryData={{ destinations: destinationData, sources: [] }}
        onComponentsSelected={() => {}}
        disableSelectionOnClick
        checkboxSelection
        onEditDestination={() => {}}
      />
    );

    screen.getByText(Destination1.metadata.name);
    screen.getByText(Destination2.metadata.name);
  });

  it("uses the expected GridSelectionModel", () => {
    function onComponentsSelected(m: GridSelectionModel) {
      expect(m).toEqual([
        `Destination|${Destination1.metadata.name}`,
        `Destination|${Destination2.metadata.name}`,
      ]);
    }
    render(
      <ComponentsDataGrid
        loading={false}
        queryData={{ destinations: destinationData, sources: [] }}
        onComponentsSelected={onComponentsSelected}
        disableSelectionOnClick
        checkboxSelection
        onEditDestination={() => {}}
      />
    );

    screen.getByLabelText("Select all rows").click();
  });

  it("calls onEditDestination when destinations are selected", async () => {
    let editCalled: boolean = false;
    function onEditDestination() {
      editCalled = true;
    }
    render(
      <ComponentsDataGrid
        loading={false}
        queryData={{ destinations: destinationData, sources: [] }}
        onComponentsSelected={() => {}}
        disableSelectionOnClick
        checkboxSelection
        onEditDestination={onEditDestination}
      />
    );

    screen.getByText(Destination1.metadata.name).click();

    await waitFor(() => expect(editCalled).toEqual(true));
  });
});
