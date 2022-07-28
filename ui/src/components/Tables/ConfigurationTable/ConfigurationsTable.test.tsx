import { render, screen } from "@testing-library/react";
import { ConfigurationsTable } from ".";
import { MemoryRouter } from "react-router-dom";
import {
  Configuration,
  ConfigurationChangesDocument,
  GetConfigurationTableDocument,
  GetConfigurationTableQuery,
} from "../../../graphql/generated";
import { MockedProvider, MockedResponse } from "@apollo/client/testing";

const TEST_CONFIGS: Pick<Configuration, "metadata">[] = [
  {
    metadata: {
      id: "1",
      name: "config-1",
      description: "description for config-1",
      labels: {
        env: "test",
        foo: "bar",
      },
    },
  },
  {
    metadata: {
      id: "2",
      name: "config-2",
      description: "description for config-2",
      labels: {
        env: "test",
        foo: "bar",
      },
    },
  },
];

const QUERY_RESULT: GetConfigurationTableQuery = {
  configurations: {
    configurations: TEST_CONFIGS,
    query: "",
    suggestions: [],
  },
};

const mocks: MockedResponse<Record<string, any>>[] = [
  {
    request: {
      query: GetConfigurationTableDocument,
      variables: {
        query: "",
      },
    },
    result: () => {
      return { data: QUERY_RESULT };
    },
  },
  {
    request: {
      query: ConfigurationChangesDocument,
      variables: {
        query: "",
      },
    },
    result: () => {
      return {
        data: { configurationChanges: [] },
      };
    },
  },
];

describe("ConfigurationsTable", () => {
  it("renders rows of configurations", async () => {
    render(
      <MemoryRouter>
        <MockedProvider mocks={mocks} addTypename={false}>
          <ConfigurationsTable />
        </MockedProvider>
      </MemoryRouter>
    );

    const rowOne = await screen.findByText("config-1");
    expect(rowOne).toBeInTheDocument();
    const rowTwo = await screen.findByText("config-2");
    expect(rowTwo).toBeInTheDocument();
  });

  it("shows delete button after selecting row", async () => {
    render(
      <MemoryRouter>
        <MockedProvider mocks={mocks} addTypename={false}>
          <ConfigurationsTable />
        </MockedProvider>
      </MemoryRouter>
    );

    // sanity check
    const row1 = await screen.findByText("config-1");
    expect(row1).toBeInTheDocument();

    const checkbox = await screen.findByLabelText("Select all rows");
    checkbox.click();

    const deleteButton = await screen.findByText("Delete 2 Configurations");
    expect(deleteButton).toBeInTheDocument();
  });

  it("opens the delete dialog after clicking delete", async () => {
    render(
      <MemoryRouter>
        <MockedProvider mocks={mocks} addTypename={false}>
          <ConfigurationsTable />
        </MockedProvider>
      </MemoryRouter>
    );

    const row1 = await screen.findByText("config-1");
    expect(row1).toBeInTheDocument();

    const checkbox = await screen.findByLabelText("Select all rows");
    checkbox.click();

    const deleteButton = await screen.findByText("Delete 2 Configurations");
    deleteButton.click();

    const dialog = await screen.findByTestId("delete-dialog");
    expect(dialog).toBeInTheDocument();
  });
});
