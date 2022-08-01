import { gql } from "@apollo/client";
import { Card, CardContent, Stack, Typography } from "@mui/material";
import { useSnackbar } from "notistack";
import { useState } from "react";
import { ShowPageConfig } from ".";
import { ConfirmDeleteResourceDialog } from "../../../components/ConfirmDeleteResourceDialog";
import { EditResourceDialog } from "../../../components/EditResourceDialog";
import {
  ResourceConfiguration,
  useDestinationTypeQuery,
} from "../../../graphql/generated";
import { UpdateStatus } from "../../../types/resources";
import { BPConfiguration } from "../../../utils/classes/configuration";
import { BPResourceConfiguration } from "../../../utils/classes/resource-configuration";

import styles from "./configuration-page.module.scss";

gql`
  query DestinationType($name: String!) {
    destinationType(name: $name) {
      metadata {
        displayName
        name
        icon
        displayName
        description
      }
      spec {
        parameters {
          label
          name
          description
          required
          type
          default
          relevantIf {
            name
            operator
            value
          }
          validValues
          options {
            creatable
          }
        }
      }
    }
  }
`;

export const InlineDestinationCard: React.FC<{
  // called when destination is successfully deleted.
  refetch: () => void;
  destination: ResourceConfiguration;
  destinationIndex: number;
  configuration: NonNullable<ShowPageConfig>;
}> = ({ destination, destinationIndex, configuration, refetch }) => {
  // We can count on the type existing for an inline Resource
  const name = destination.type!;
  const { data } = useDestinationTypeQuery({
    variables: { name },
  });
  const { enqueueSnackbar } = useSnackbar();

  const [editing, setEditing] = useState(false);
  const [confirmDeleteOpen, setDeleteOpen] = useState(false);

  if (data?.destinationType == null) {
    return null;
  }

  const icon = data.destinationType.metadata.icon;
  const displayName =
    data.destinationType.metadata.displayName ??
    data.destinationType.metadata.name;
  const description = data.destinationType.metadata.description ?? "";

  async function onDelete() {
    const updatedConfig = new BPConfiguration(configuration);
    updatedConfig.removeDestination(destinationIndex);

    try {
      const { status, reason } = await updatedConfig.apply();
      if (status === UpdateStatus.INVALID) {
        throw new Error(
          `failed to update configuration, configuration invalid, ${reason}`
        );
      }

      closeDeleteDialog();
      closeEditDialog();
      refetch();
    } catch (err) {
      enqueueSnackbar("Failed to update configuration.", { variant: "error" });
      console.error(err);
    }
  }

  async function onSave(formValues: Record<string, any>) {
    const resourceConfiguration = new BPResourceConfiguration();
    resourceConfiguration.setParamsFromMap(formValues);
    resourceConfiguration.type = data!.destinationType!.metadata.name;

    const updatedConfig = new BPConfiguration(configuration);
    updatedConfig.replaceDestination(resourceConfiguration, destinationIndex);

    try {
      const { status, reason } = await updatedConfig.apply();
      if (status === UpdateStatus.INVALID) {
        throw new Error(
          `failed to update configuration, configuration invalid, ${reason}`
        );
      }

      enqueueSnackbar("Saved updated configuration.", { variant: "success" });
      closeEditDialog();
      refetch();
    } catch (err) {}
  }

  function closeEditDialog() {
    setEditing(false);
  }

  function closeDeleteDialog() {
    setDeleteOpen(false);
  }

  return (
    <>
      <Card
        className={styles["resource-card"]}
        onClick={() => setEditing(true)}
      >
        <CardContent>
          <Stack alignItems="center">
            <span
              className={styles.icon}
              style={{ backgroundImage: `url(${icon})` }}
            />
            <Typography component="div" fontWeight={600}>
              {displayName}
            </Typography>
          </Stack>
        </CardContent>
      </Card>

      <EditResourceDialog
        kind="destination"
        fullWidth
        maxWidth="sm"
        title={displayName}
        description={description}
        parameters={destination.parameters ?? []}
        parameterDefinitions={data.destinationType.spec.parameters}
        open={editing}
        onClose={closeEditDialog}
        onCancel={closeEditDialog}
        onDelete={() => setDeleteOpen(true)}
        onSave={onSave}
      />

      <ConfirmDeleteResourceDialog
        open={confirmDeleteOpen}
        onClose={closeDeleteDialog}
        onCancel={closeDeleteDialog}
        onDelete={onDelete}
        action={"remove"}
      >
        <Typography>
          Are you sure you want to remove this destination?
        </Typography>
      </ConfirmDeleteResourceDialog>
    </>
  );
};
