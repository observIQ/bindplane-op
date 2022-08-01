import { ApolloProvider } from "@apollo/client";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import nock from "nock";
import { SnackbarProvider } from "notistack";
import { MemoryRouter } from "react-router-dom";
import { AssistedConfigWizard } from ".";
import APOLLO_CLIENT from "../../../../apollo-client";
import {
  SourceTypesQuery,
  ParameterType,
  DestinationsAndTypesQuery,
} from "../../../../graphql/generated";
import {
  APIVersion,
  Resource,
  UpdateStatus,
} from "../../../../types/resources";

const dummySourceType: SourceTypesQuery["sourceTypes"][0] = {
  __typename: "SourceType",
  apiVersion: APIVersion.V1_BETA,
  kind: "SourceType",
  metadata: {
    __typename: "Metadata",
    name: "test-source-type",
    id: "test-source-type",
    displayName: "Test Source",
    icon: "/path/to/icon",
    description: "",
  },
  spec: {
    __typename: "ResourceTypeSpec",
    supportedPlatforms: ["linux", "macos", "windows"],
    version: "0.0.0",
    parameters: [
      {
        __typename: "ParameterDefinition",
        name: "parameter1",
        label: "First Parameter",
        description: "A description for the first parameter",
        type: ParameterType.Bool,
        default: true,
        required: false,
        validValues: null,
        relevantIf: null,
        options: {},
      },
      {
        __typename: "ParameterDefinition",
        name: "parameter2",
        label: "Second Parameter",
        description: "A description for the second parameter",
        type: ParameterType.String,
        default: "default-value",
        required: false,
        validValues: null,
        relevantIf: null,
        options: {},
      },
    ],
    telemetryTypes: [],
  },
};

const dummyDestinationType: DestinationsAndTypesQuery["destinationTypes"][0] = {
  __typename: "DestinationType",
  apiVersion: APIVersion.V1_BETA,
  kind: "SourceType",
  metadata: {
    __typename: "Metadata",
    name: "test-destination-type",
    id: "test-destination-type",
    displayName: "Test Destination",
    icon: "/path/to/icon",
    description: "",
  },
  spec: {
    __typename: "ResourceTypeSpec",
    supportedPlatforms: ["linux", "macos", "windows"],
    version: "0.0.0",
    parameters: [
      {
        __typename: "ParameterDefinition",
        name: "parameter1",
        label: "First Parameter",
        description: "A description for the first parameter",
        type: ParameterType.Bool,
        default: true,
        required: false,
        validValues: null,
        relevantIf: null,
        options: {},
      },
      {
        __typename: "ParameterDefinition",
        name: "parameter2",
        label: "Second Parameter",
        description: "A description for the second parameter",
        type: ParameterType.String,
        default: "default-value",
        required: false,
        validValues: null,
        relevantIf: null,
        options: {},
      },
    ],
    telemetryTypes: [],
  },
};

const sourceTypesQuery: SourceTypesQuery = {
  sourceTypes: [dummySourceType],
};

const destinationTypesQuery: DestinationsAndTypesQuery = {
  destinationTypes: [dummyDestinationType],
  destinations: [],
};

beforeEach(() => {
  nock("http://localhost:80")
    .post("/v1/graphql", (body) => {
      return body.operationName === "sourceTypes";
    })
    .once()
    .reply(200, {
      data: sourceTypesQuery,
    });

  const gqlscope = nock("http://localhost")
    .post("/v1/graphql", (body) => {
      return body.operationName === "DestinationsAndTypes";
    })
    .once()
    .reply(200, {
      data: destinationTypesQuery,
    });

  gqlscope
    .post("/v1/graphql", (body) => {
      return body.operationName === "getConfigNames";
    })
    .once()
    .reply(200, {
      data: {
        configurations: [],
      },
    });
});

describe("AssistedConfigWizard", () => {
  it("requires name and platform to go to step 2", async () => {
    render(
      <ApolloProvider client={APOLLO_CLIENT}>
        <MemoryRouter>
          <AssistedConfigWizard />
        </MemoryRouter>
      </ApolloProvider>
    );

    // Hit next with no form values
    screen.getByText("Next").click();
    expect(screen.getByTestId("step-one")).toBeInTheDocument();

    // Expect to see Required for Name and Platform fields.
    const requiredErrors = screen.getAllByText("Required.");
    expect(requiredErrors.length).toEqual(2);
  });

  it("can navigate to step two", () => {
    render(
      <ApolloProvider client={APOLLO_CLIENT}>
        <MemoryRouter>
          <AssistedConfigWizard />
        </MemoryRouter>
      </ApolloProvider>
    );

    goToStepTwo("test");
    expect(screen.getByTestId("step-two")).toBeInTheDocument();
  });

  it("can add a source via the ResourceDialog", async () => {
    render(
      <ApolloProvider client={APOLLO_CLIENT}>
        <MemoryRouter>
          <AssistedConfigWizard />
        </MemoryRouter>
      </ApolloProvider>
    );

    goToStepTwo("test");

    // Open dialog
    screen.getByText("Add Source").click();
    expect(screen.getByTestId("resource-dialog")).toBeInTheDocument();

    // Select Test Source
    const button = await screen.findByText("Test Source");
    button.click();

    // Save it
    screen.getByText("Save").click();

    // Verify it has an accordion
    const sourceAccordion = screen.getByTestId("source-accordion");
    sourceAccordion.click();

    // Verify it renders parameter table
    expect(screen.getByText("First Parameter")).toBeInTheDocument();
  });

  it("can delete a source", async () => {
    render(
      <ApolloProvider client={APOLLO_CLIENT}>
        <MemoryRouter>
          <AssistedConfigWizard />
        </MemoryRouter>
      </ApolloProvider>
    );

    goToStepTwo("test");

    // Open dialog
    screen.getByText("Add Source").click();
    expect(screen.getByTestId("resource-dialog")).toBeInTheDocument();

    // Select Test Source
    const button = await screen.findByText("Test Source");
    button.click();

    // Save it
    screen.getByText("Save").click();

    // Open accordion
    const sourceAccordion = screen.getByTestId("source-accordion");
    sourceAccordion.click();

    // Hit Delete
    screen.getByText("Delete").click();
    expect(screen.getByTestId("confirm-delete-dialog")).toBeInTheDocument();

    // Check for dialog
    screen.getByTestId("confirm-delete-dialog-delete-button").click();

    // Confirm delete via dialog
    const check = screen.queryByTestId("source-accordion");

    // Verify source is gone
    expect(check).not.toBeInTheDocument();
  });

  it("can add a destination via the ResourceDialog", async () => {
    render(
      <ApolloProvider client={APOLLO_CLIENT}>
        <MemoryRouter>
          <AssistedConfigWizard />
        </MemoryRouter>
      </ApolloProvider>
    );

    goToStepThree("test");

    // Open the dialog
    screen.getByTestId("add-destination-button").click();
    expect(screen.getByTestId("resource-dialog")).toBeInTheDocument();

    // Select Destination
    const destinationButton = await screen.findByText("Test Destination");
    expect(destinationButton).toBeInTheDocument();
    destinationButton.click();

    // Save it
    screen.getByTestId("resource-form-save").click();

    // Verify accordion is present
    expect(screen.getByTestId("destination-accordion")).toBeInTheDocument();
  });

  it("can remove a destination", async () => {
    render(
      <ApolloProvider client={APOLLO_CLIENT}>
        <MemoryRouter>
          <AssistedConfigWizard />
        </MemoryRouter>
      </ApolloProvider>
    );

    goToStepThree("test");

    // Open the dialog
    screen.getByTestId("add-destination-button").click();
    expect(screen.getByTestId("resource-dialog")).toBeInTheDocument();

    // Select Destination
    const destinationButton = await screen.findByText("Test Destination");
    expect(destinationButton).toBeInTheDocument();
    destinationButton.click();

    // Save it
    screen.getByTestId("resource-form-save").click();

    // Verify accordion is present
    const destAccordion = screen.getByTestId("destination-accordion");
    expect(destAccordion).toBeInTheDocument();
    destAccordion.click();

    // Hit delete
    screen.getByText("Remove").click();

    // Verify modal pops up
    screen.getByTestId("confirm-delete-dialog-delete-button").click();

    // Confirm delete via dialog
    const check = screen.queryByTestId("destination-accordion");

    // Verify destination is gone
    expect(check).not.toBeInTheDocument();
  });

  it("can edit a destination", async () => {
    render(
      <ApolloProvider client={APOLLO_CLIENT}>
        <MemoryRouter>
          <SnackbarProvider>
            <AssistedConfigWizard />
          </SnackbarProvider>
        </MemoryRouter>
      </ApolloProvider>
    );
    const configName = "this-is-the-config-name";

    let postData: any;
    let applyDone = false;
    // Track the save payload
    nock("http://localhost")
      .post("/v1/apply", (body) => {
        applyDone = true;
        postData = body;
        return true;
      })
      .once()
      .reply(202, {
        updates: [
          {
            resource: {
              metadata: {
                name: configName,
              },
            },
            status: UpdateStatus.CREATED,
          },
          {
            resource: {
              metadata: {
                name: "dest-name",
              },
            },
            status: UpdateStatus.CREATED,
          },
        ],
      });

    goToStepThree(configName);

    // Open the dialog
    screen.getByTestId("add-destination-button").click();
    expect(screen.getByTestId("resource-dialog")).toBeInTheDocument();

    // Select Destination
    const destinationButton = await screen.findByText("Test Destination");
    expect(destinationButton).toBeInTheDocument();
    destinationButton.click();

    // set the name
    const nameField = await screen.findByTestId("name-field");
    fireEvent.change(nameField, { target: { value: "dest-name" } });

    // Save it
    screen.getByTestId("resource-form-save").click();

    // We should see the destination name
    await screen.findByText("dest-name");

    // Verify accordion is present
    let destAccordion = screen.getByTestId("destination-accordion");
    expect(destAccordion).toBeInTheDocument();
    destAccordion.click();

    // hit edit
    screen.getByText("Edit").click();
    await screen.findByTestId("resource-form");

    // should not be a name field
    expect(screen.queryByTestId("name-field")).not.toBeInTheDocument();

    // edit the field
    const newValue = "!!!!!!!!!";
    const field = await screen.findByLabelText("Second Parameter");
    fireEvent.change(field, { target: { value: newValue } });

    // save it
    screen.getByTestId("resource-form-save").click();

    // verify the new value is present
    await screen.findByText(newValue);

    // we should still see the correct destination name
    await screen.findByText("dest-name");

    // hit save and make sure the values are what we expect
    screen.getByTestId("save-button").click();

    await waitFor(() => expect(applyDone).toEqual(true));

    const sentDestination = postData.resources.find(
      (r: Resource) => r.kind === "Destination"
    );

    const expectDestination = {
      apiVersion: "bindplane.observiq.com/v1beta",
      kind: "Destination",
      metadata: {
        id: "dest-name",
        name: "dest-name",
      },
      spec: {
        parameters: [
          {
            name: "parameter1",
            value: true,
          },
          {
            name: "parameter2",
            value: "!!!!!!!!!",
          },
        ],
        type: "test-destination-type",
      },
    };

    expect(sentDestination).toEqual(expectDestination);
  });
});

function goToStepTwo(name: string) {
  fireEvent.change(screen.getByLabelText("Name"), {
    target: { value: name },
  });

  fireEvent.mouseDown(screen.getByLabelText("Platform"));
  screen.getByText("Linux").click();

  screen.getByText("Next").click();
  expect(screen.getByTestId("step-two")).toBeInTheDocument();
}

function goToStepThree(name: string) {
  goToStepTwo(name);

  screen.getByText("Next").click();

  expect(screen.getByTestId("step-three")).toBeInTheDocument();
}
