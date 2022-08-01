import { gql } from "@apollo/client";
import { Card, CardContent, Stack, Typography } from "@mui/material";
import { useSnackbar } from "notistack";
import { memo, useState } from "react";
import { ShowPageConfig } from ".";
import { ConfirmDeleteResourceDialog } from "../../../components/ConfirmDeleteResourceDialog";
import { EditResourceDialog } from "../../../components/EditResourceDialog";
import {
  ResourceConfiguration,
  useGetDestinationWithTypeQuery,
} from "../../../graphql/generated";
import { UpdateStatus } from "../../../types/resources";
import { BPConfiguration, BPDestination } from "../../../utils/classes";

import styles from "./configuration-page.module.scss";

gql`
  query getDestinationWithType($name: String!) {
    destinationWithType(name: $name) {
      destination {
        metadata {
          name
          id
          labels
        }
        spec {
          type
          parameters {
            name
            value
          }
        }
      }
      destinationType {
        metadata {
          name
          icon
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
  }
`;

const ResourceDestinationCardComponent: React.FC<{
  configuration: NonNullable<ShowPageConfig>;
  destination: ResourceConfiguration;
  destinationIndex: number;
  refetch: () => void;
}> = ({ destination, destinationIndex, configuration, refetch }) => {
  const { data, refetch: refetchDestination } = useGetDestinationWithTypeQuery({
    variables: { name: destination.name ?? "" },
    fetchPolicy: "cache-and-network",
  });

  const { enqueueSnackbar } = useSnackbar();

  const [editing, setEditing] = useState(false);
  const [confirmDeleteOpen, setDeleteOpen] = useState(false);

  function closeEditDialog() {
    setEditing(false);
  }

  async function onSave(formValues: Record<string, any>) {
    const updatedDestination = new BPDestination(
      data!.destinationWithType!.destination!
    );

    updatedDestination.setParamsFromMap(formValues);

    try {
      const update = await updatedDestination.apply();
      if (update.status === UpdateStatus.INVALID) {
        console.error("Update: ", update);
        throw new Error(
          `failed to apply destination, got status ${update.status}`
        );
      }

      enqueueSnackbar("Successfully saved destination.", {
        variant: "success",
      });
      setEditing(false);
      refetch();
      refetchDestination();
    } catch (err) {
      console.error(err);
      enqueueSnackbar("Failed to update destination.", { variant: "error" });
    }
  }

  async function onDelete() {
    const updatedConfig = new BPConfiguration(configuration);
    updatedConfig.removeDestination(destinationIndex);

    try {
      const update = await updatedConfig.apply();
      if (update.status === UpdateStatus.INVALID) {
        console.error("Update: ", update);
        throw new Error(
          `failed to remove destination from configuration, configuration invalid`
        );
      }

      closeEditDialog();
      closeDeleteDialog();
      refetch();
      refetchDestination();
    } catch (err) {
      enqueueSnackbar("Failed to remove destination.", { variant: "error" });
    }
  }

  function closeDeleteDialog() {
    setDeleteOpen(false);
  }

  // Loading
  if (data === undefined) {
    return null;
  }

  if (data.destinationWithType.destination == null) {
    enqueueSnackbar(`Could not retrieve destination ${destination.name!}.`, {
      variant: "error",
    });
    return null;
  }

  if (data.destinationWithType.destinationType == null) {
    enqueueSnackbar(
      `Could not retrieve destination type for destination ${destination.name!}.`,
      { variant: "error" }
    );
    return null;
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
              style={{
                backgroundImage: `url(${data?.destinationWithType?.destinationType?.metadata.icon})`,
              }}
            />
            <Typography component="div" fontWeight={600}>
              {destination.name}
            </Typography>
          </Stack>
        </CardContent>
      </Card>

      <EditResourceDialog
        kind="destination"
        fullWidth
        maxWidth="sm"
        title={destination.name!}
        description={""}
        parameters={data.destinationWithType.destination.spec.parameters ?? []}
        parameterDefinitions={
          data.destinationWithType.destinationType.spec.parameters
        }
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

export const ResourceDestinationCard = memo(ResourceDestinationCardComponent);
