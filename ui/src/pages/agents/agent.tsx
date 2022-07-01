import { gql } from "@apollo/client";
import {
  Dialog,
  DialogContent,
  Grid,
  Stack,
  Typography,
  Alert,
  AlertTitle,
} from "@mui/material";
import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { CardContainer } from "../../components/CardContainer";
import { ManageConfigForm } from "../../components/ManageConfigForm";
import { AgentTable } from "../../components/Tables/AgentTable";
import { useGetAgentAndConfigurationsQuery } from "../../graphql/generated";
import { useAgentChangesContext } from "../../hooks/useAgentChanges";
import { RawConfigWizard } from "../configurations/wizards/RawConfigWizard";
import { useSnackbar } from "notistack";
import { labelAgents } from "../../utils/rest/label-agents";
import { RawConfigFormValues } from "../../types/forms";
import { withRequireLogin } from "../../contexts/RequireLogin";
import { withNavBar } from "../../components/NavBar";
import { AgentChangesProvider } from "../../contexts/AgentChanges";

import mixins from "../../styles/mixins.module.scss";

gql`
  query GetAgentAndConfigurations($agentId: ID!) {
    agent(id: $agentId) {
      id
      name
      architecture
      operatingSystem
      labels
      hostName
      platform
      version
      macAddress
      remoteAddress
      home
      status
      connectedAt
      disconnectedAt
      errorMessage
      configuration {
        Collector
      }
      configurationResource {
        metadata {
          name
        }
      }
    }
    configurations {
      configurations {
        metadata {
          name
          labels
        }
        spec {
          raw
        }
      }
    }
  }
`;

const AgentPageContent: React.FC = () => {
  const { id } = useParams();
  const snackbar = useSnackbar();
  const [importOpen, setImportOpen] = useState(false);

  // AgentChanges subscription to trigger a refetch.
  const agentChanges = useAgentChangesContext();

  const { data, refetch } = useGetAgentAndConfigurationsQuery({
    variables: { agentId: id ?? "" },
    fetchPolicy: "network-only",
  });

  const navigate = useNavigate();

  async function onImportSuccess(values: RawConfigFormValues) {
    if (data?.agent != null) {
      try {
        await labelAgents(
          [data.agent.id],
          { configuration: values.name },
          true
        );
      } catch (err) {
        snackbar.enqueueSnackbar("Failed to apply label to agent.", {
          variant: "error",
        });
      }
    }

    setImportOpen(false);
  }

  useEffect(() => {
    if (agentChanges.length > 0) {
      const thisAgent = agentChanges
        .map((c) => c.agent)
        .find((a) => a.id === id);
      if (thisAgent != null) {
        refetch();
      }
    }
  }, [agentChanges, id, refetch]);

  // Here we use the distinction between graphql returning null vs undefined.
  // If the agent is null then this agent doesn't exist, redirect to agents.
  if (data?.agent === null) {
    navigate("/agents");
    return null;
  }

  // Data is loading, return null for now.
  if (data === undefined || data.agent == null) {
    return null;
  }

  return (
    <>
      <CardContainer>
        <Typography variant="h5" classes={{ root: mixins["mb-5"] }}>
          Agent - {data.agent.name}
        </Typography>
        <Grid container spacing={5}>
          <Grid item xs={12} lg={6}>
            <Typography variant="h6" classes={{ root: mixins["mb-2"] }}>
              Details
            </Typography>
            <AgentTable agent={data.agent} />
            {data.agent.errorMessage && (
              <Alert severity="error" classes={{ root: mixins["mt-3"] }}>
                <AlertTitle>Error</AlertTitle>
                {data.agent.errorMessage}
              </Alert>
            )}
          </Grid>
          <Grid item xs={12} lg={6}>
            <ManageConfigForm
              agent={data.agent}
              configurations={data.configurations.configurations ?? []}
              onImport={() => setImportOpen(true)}
            />
          </Grid>
        </Grid>
      </CardContainer>

      {/** Raw Config wizard for importing an agents config */}
      <Dialog
        open={importOpen}
        onClose={() => setImportOpen(false)}
        PaperComponent={EmptyComponent}
      >
        <DialogContent>
          <Stack justifyContent="center" alignItems="center" height="100%">
            <RawConfigWizard
              onClose={() => setImportOpen(false)}
              initialValues={{
                name: data.agent.name,
                description: `Imported config from agent ${data.agent.name}.`,
                fileName: "",
                rawConfig: data.agent.configuration?.Collector ?? "",
                platform: configPlatformFromAgentPlatform(data.agent.platform),
              }}
              onSuccess={onImportSuccess}
              fromImport
            />
          </Stack>
        </DialogContent>
      </Dialog>
    </>
  );
};

const EmptyComponent: React.FC = ({ children }) => {
  return <>{children}</>;
};

function configPlatformFromAgentPlatform(platform: string | null | undefined) {
  if (platform == null) return "linux";
  if (platform === "darwin") return "macos";
  return platform;
}

export const AgentPage = withRequireLogin(
  withNavBar(() => (
    <AgentChangesProvider>
      <AgentPageContent />
    </AgentChangesProvider>
  ))
);
