import { gql } from "@apollo/client";
import { Alert, IconButton, Snackbar, Typography } from "@mui/material";
import React, { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { CardContainer } from "../../../components/CardContainer";
import {
  GetConfigurationQuery,
  useGetConfigurationQuery,
} from "../../../graphql/generated";
import { GridDensityTypes } from "@mui/x-data-grid";
import { AgentsTable } from "../../../components/Tables/AgentsTable";
import { AgentsTableField } from "../../../components/Tables/AgentsTable/AgentsDataGrid";
import { PlusCircleIcon } from "../../../components/Icons";
import { selectorString } from "../../../types/configuration";
import { ApplyConfigDialog } from "./ApplyConfigDialog";
import { DetailsSection } from "./DetailsSection";
import { ConfigurationSection } from "./ConfigurationSection";
import { SourcesSection } from "./SourcesSection";
import { DestinationsSection } from "./DestinationsSection";
import { useSnackbar } from "notistack";

import styles from "./configuration-page.module.scss";
import { withRequireLogin } from "../../../contexts/RequireLogin";
import { withNavBar } from "../../../components/NavBar";

gql`
  query GetConfiguration($name: String!) {
    configuration(name: $name) {
      metadata {
        id
        name
        description
        labels
      }
      spec {
        raw
        sources {
          type
          name
          parameters {
            name
            value
          }
        }
        destinations {
          type
          name
          parameters {
            name
            value
          }
        }
        selector {
          matchLabels
        }
      }
    }
  }
`;

export type ShowPageConfig = GetConfigurationQuery["configuration"];

const ConfigPageContent: React.FC = () => {
  const { name } = useParams();

  // Get Configuration Data
  const { data, refetch } = useGetConfigurationQuery({
    variables: { name: name ?? "" },
    fetchPolicy: "cache-and-network",
  });

  const [showApplyDialog, setShowApply] = useState(false);
  const [applySuccess, setApplySuccess] = useState(false);
  const [saveDescriptionSuccess, setSaveDescriptionSuccess] = useState(false);
  const [saveConfigSuccess, setSaveConfigSuccess] = useState(false);

  const [applyError, setApplyError] = useState(false);
  const [saveDescriptionError, setSaveDescriptionError] = useState(false);
  const [saveConfigError, setSaveConfigError] = useState(false);

  const navigate = useNavigate();
  const { enqueueSnackbar } = useSnackbar();

  const isRaw = (data?.configuration?.spec?.raw?.length || 0) > 0;
  function openApplyDialog() {
    setShowApply(true);
  }

  function closeApplyDialog() {
    setShowApply(false);
  }

  function onApplySuccess() {
    setApplySuccess(true);
    closeApplyDialog();
  }

  if (data?.configuration === undefined) {
    return null;
  }

  if (data.configuration === null) {
    enqueueSnackbar(`No configuration with name ${name} found.`, {
      variant: "error",
    });
    navigate("/configurations");
    return null;
  }

  return (
    <>
      <section>
        <DetailsSection
          configuration={data.configuration}
          refetch={refetch}
          onSaveDescriptionError={() => setSaveDescriptionError(true)}
          onSaveDescriptionSuccess={() => setSaveConfigSuccess(true)}
        />
      </section>

      {isRaw && (
        <section>
          <ConfigurationSection
            configuration={data.configuration}
            refetch={refetch}
            onSaveSuccess={() => setSaveConfigSuccess(true)}
            onSaveError={() => setSaveConfigError(true)}
          />
        </section>
      )}

      {!isRaw && (
        <section>
          <SourcesSection
            configuration={data.configuration}
            refetch={refetch}
          />
        </section>
      )}

      {!isRaw && (
        <section>
          <DestinationsSection
            configuration={data.configuration}
            destinations={data.configuration.spec.destinations ?? []}
            refetch={refetch}
          />
        </section>
      )}

      <section>
        <CardContainer>
          <div className={styles["title-button-row"]}>
            <Typography variant="h5">Agents</Typography>
            <IconButton onClick={openApplyDialog} color="primary">
              <PlusCircleIcon />
            </IconButton>
          </div>

          <AgentsTable
            selector={selectorString(data.configuration.spec.selector)}
            columnFields={[
              AgentsTableField.NAME,
              AgentsTableField.STATUS,
              AgentsTableField.OPERATING_SYSTEM,
            ]}
            density={GridDensityTypes.Compact}
            minHeight="300px"
          />
        </CardContainer>
      </section>

      {showApplyDialog && (
        <ApplyConfigDialog
          configuration={data.configuration}
          maxWidth="lg"
          fullWidth
          open={showApplyDialog}
          onError={() => setApplyError(true)}
          onSuccess={onApplySuccess}
          onClose={closeApplyDialog}
          onCancel={closeApplyDialog}
        />
      )}

      <Snackbar
        open={applySuccess}
        onClose={() => setApplySuccess(false)}
        autoHideDuration={6000}
      >
        <Alert onClose={() => setApplySuccess(false)} severity="success">
          Successfully applied configuration!
        </Alert>
      </Snackbar>

      <Snackbar
        open={saveDescriptionSuccess}
        onClose={() => setSaveDescriptionSuccess(false)}
        autoHideDuration={6000}
      >
        <Alert
          onClose={() => setSaveDescriptionSuccess(false)}
          severity="success"
        >
          Successfully saved description!
        </Alert>
      </Snackbar>

      <Snackbar
        open={saveConfigSuccess}
        autoHideDuration={6000}
        onClose={() => setSaveConfigSuccess(false)}
      >
        <Alert onClose={() => setSaveConfigSuccess(false)} severity="success">
          Successfully saved configuration!
        </Alert>
      </Snackbar>

      <Snackbar
        open={applyError}
        autoHideDuration={6000}
        onClose={() => setApplyError(false)}
      >
        <Alert onClose={() => setApplyError(false)} severity="error">
          Failed to apply configuration.
        </Alert>
      </Snackbar>

      <Snackbar
        open={saveDescriptionError}
        autoHideDuration={6000}
        onClose={() => setSaveDescriptionError(false)}
      >
        <Alert onClose={() => setSaveDescriptionError(false)} severity="error">
          Failed to save description.
        </Alert>
      </Snackbar>

      <Snackbar
        open={saveConfigError}
        autoHideDuration={6000}
        onClose={() => setSaveConfigError(false)}
      >
        <Alert onClose={() => setSaveConfigError(false)} severity="error">
          Failed to save configuration!
        </Alert>
      </Snackbar>
    </>
  );
};

export const ViewConfiguration = withRequireLogin(
  withNavBar(ConfigPageContent)
);
