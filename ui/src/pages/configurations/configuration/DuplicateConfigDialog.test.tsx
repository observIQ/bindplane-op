import { ApolloProvider } from "@apollo/client";
import { render, screen } from "@testing-library/react";
import { SnackbarProvider } from "notistack";
import { MemoryRouter } from "react-router-dom";
import APOLLO_CLIENT from "../../../apollo-client";
import { DuplicateConfigDialog } from "./DuplicateConfigDialog";

describe("DuplicateConfigDialog", () => {
  it("renders without error", () => {
    render(
      <SnackbarProvider>
        <ApolloProvider client={APOLLO_CLIENT}>
          <MemoryRouter>
            <DuplicateConfigDialog
              open={true}
              currentConfigName={"current-config-name"}
            />
          </MemoryRouter>
        </ApolloProvider>
      </SnackbarProvider>
    );
  });

  it("disables save button by default", () => {
    render(
      <SnackbarProvider>
        <ApolloProvider client={APOLLO_CLIENT}>
          <MemoryRouter>
            <DuplicateConfigDialog
              open={true}
              currentConfigName={"current-config-name"}
            />
          </MemoryRouter>
        </ApolloProvider>
      </SnackbarProvider>
    );

    const saveButton = screen.getByText("Save");
    expect(saveButton).toBeDisabled();
  });
});
