import React from "react";
import APOLLO_CLIENT from "./apollo-client";
import { ApolloProvider } from "@apollo/client";
import { BrowserRouter, Route, Routes, Navigate } from "react-router-dom";
import {
  ConfigurationsPage,
  AgentsPage,
  NewConfigurationPage,
  InstallPage,
  AgentPage,
} from "./pages";
import { theme } from "./theme";
import { StyledEngineProvider, ThemeProvider } from "@mui/material";
import { ViewConfiguration } from "./pages/configurations/configuration";
import { NewRawConfigurationPage } from "./pages/configurations/new-raw";
import { SnackbarProvider } from "notistack";
import { ComponentsPage } from "./pages/components";
import { Version } from "./components/Version";
import { LoginPage } from "./pages/login";
import { LiveTailPage } from "./pages/agents/livetail";

export const App: React.FC = () => {
  return (
    <StyledEngineProvider injectFirst>
      <ThemeProvider theme={theme}>
        <ApolloProvider client={APOLLO_CLIENT}>
          <SnackbarProvider>
            <BrowserRouter>
              <Routes>
                <Route path="/login" element={<LoginPage />} />
                {/* --------------- The following routes require authentication -------------- */}
                {/* No path at "/", reroute to agents */}
                <Route path="/" element={<Navigate to="/agents" />} />
                <Route path="agents">
                  <Route index element={<AgentsPage />} />
                  <Route path="install" element={<InstallPage />} />
                  <Route path=":id">
                    <Route index element={<AgentPage />} />
                    <Route path="livetail" element={<LiveTailPage />} />
                  </Route>
                </Route>
                f
                <Route path="configurations">
                  <Route index element={<ConfigurationsPage />} />
                  <Route path="new-raw" element={<NewRawConfigurationPage />} />
                  <Route path="new" element={<NewConfigurationPage />} />
                  <Route path=":name" element={<ViewConfiguration />} />
                </Route>
                <Route path="components">
                  <Route index element={<ComponentsPage />} />
                </Route>
              </Routes>

              <footer>
                <Version />
              </footer>
            </BrowserRouter>
          </SnackbarProvider>
        </ApolloProvider>
      </ThemeProvider>
    </StyledEngineProvider>
  );
};
