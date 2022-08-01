import { gql } from "@apollo/client";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Button,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableRow,
  Typography,
} from "@mui/material";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { AssistedWizardFormValues, ResourceConfigurationAction } from ".";
import { PlusCircleIcon } from "../../../../components/Icons";
import {
  DialogResource,
  ResourceDialog,
} from "../../../../components/ResourceDialog";
import { useWizard } from "../../../../components/Wizard/WizardContext";
import {
  Destination,
  DestinationType,
  SourceType,
  useDestinationsAndTypesQuery,
} from "../../../../graphql/generated";
import {
  APIVersion,
  Resource,
  ResourceKind,
  UpdateStatus,
} from "../../../../types/resources";
import { applyResources } from "../../../../utils/rest/apply-resources";
import { EditResourceDialog } from "../../../../components/EditResourceDialog";
import { classes } from "../../../../utils/styles";
import { ConfirmDeleteResourceDialog } from "../../../../components/ConfirmDeleteResourceDialog";
import { useSnackbar } from "notistack";
import { BPConfiguration } from "../../../../utils/classes/configuration";
import { BPResourceConfiguration } from "../../../../utils/classes/resource-configuration";

import styles from "./assisted-config-wizard.module.scss";
import mixins from "../../../../styles/mixins.module.scss";

type ResourceType = SourceType | DestinationType;

gql`
  query DestinationsAndTypes {
    destinationTypes {
      kind
      apiVersion
      metadata {
        id
        name
        displayName
        description
        icon
      }
      spec {
        version
        parameters {
          label
          type
          name
          description
          default
          validValues
          relevantIf {
            name
            value
            operator
          }
          required
          options {
            creatable
          }
        }
        supportedPlatforms
        telemetryTypes
      }
    }
    destinations {
      metadata {
        name
      }
      spec {
        type
        parameters {
          name
          value
        }
      }
    }
  }
`;

export const StepThree: React.FC = () => {
  const { goToStep, formValues, setValues } =
    useWizard<AssistedWizardFormValues>();

  const snackbar = useSnackbar();

  const [addDestinationOpen, setAddDestinationOpen] = useState(false);
  const [editingDestination, setEditingDestination] = useState<boolean>(false);
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);

  const { data } = useDestinationsAndTypesQuery({
    fetchPolicy: "network-only",
  });
  const navigate = useNavigate();

  async function onSaveConfiguration() {
    // Resources to create, could be just the Configuration or
    // a Configuration and a Destination.
    const resources: Resource[] = [];

    // Create the destination resource if present and not an existing destination.
    if (formValues.destination != null && formValues.destination.create) {
      const destinationResource: Destination = {
        apiVersion: APIVersion.V1_BETA,
        kind: ResourceKind.DESTINATION,
        spec: {
          type: formValues.destination.resourceConfiguration.type!,
          parameters: formValues.destination.resourceConfiguration.parameters,
        },
        metadata: {
          name: formValues.destination.resourceConfiguration.name!,
          id: formValues.destination.resourceConfiguration.name!,
        },
      };

      resources.push(destinationResource);
    }

    // Create the configuration
    const configuration = new BPConfiguration({
      metadata: {
        id: formValues.name,
        name: formValues.name,
        description: formValues.description,
        labels: {
          platform: formValues.platform,
        },
      },
    });

    for (const src of formValues.sources) {
      configuration.addSource(src);
    }

    if (formValues.destination) {
      configuration.addDestination({
        name: formValues.destination.resourceConfiguration.name,
      });
    }

    configuration.addMatchLabels({ configuration: formValues.name });

    resources.push(configuration);

    // Apply Resources
    try {
      const { updates } = await applyResources(resources);
      const update = updates.find(
        (u) => u.resource.metadata.name === formValues.name
      );

      if (update == null) {
        throw new Error(
          `failed to create configuration, no configuration returned with name ${formValues.name}`
        );
      }

      // Configuration was returned but not created, likely if it was valid.
      if (update.status !== UpdateStatus.CREATED) {
        throw new Error(
          `failed to create configuration, got status ${update.status}`
        );
      }

      const configPagePath = `/configurations/${update.resource.metadata.name}`;
      // Redirect to configuration page
      navigate(configPagePath);
    } catch (err) {
      snackbar.enqueueSnackbar("Failed to create configuration.", {
        variant: "error",
      });
      console.error(err);
      return;
    }
  }

  // This is the callback passed to the Dialog when a new Destination is created.
  // Here we need to set the formValues.destination with a new Resource configuration
  // with create = true.
  function onNewDestinationSave(
    values: { [key: string]: any },
    resourceType: ResourceType
  ) {
    const resourceConfiguration = new BPResourceConfiguration({
      type: resourceType.metadata.name,
    });
    resourceConfiguration.setParamsFromMap(values);
    const destinationConfiguration: ResourceConfigurationAction = {
      resourceConfiguration,
      create: true,
    };

    setValues({ destination: destinationConfiguration });
    setAddDestinationOpen(false);
  }

  function onChooseExistingDestination(resource: DialogResource) {
    setValues({
      destination: {
        resourceConfiguration: {
          name: resource.metadata.name,
          type: resource.spec.type,
        },
        create: false,
      },
    });
    setAddDestinationOpen(false);
  }

  function onEditDestinationSave(values: { [key: string]: any }) {
    const newDestination = new BPResourceConfiguration({
      name: formValues.destination?.resourceConfiguration.name,
      type: formValues.destination?.resourceConfiguration.type,
    });
    newDestination.setParamsFromMap(values);

    setValues({
      destination: { resourceConfiguration: newDestination, create: true },
    });
    setEditingDestination(false);
  }

  function onDestinationDelete() {
    setConfirmDeleteOpen(false);
    setValues({ destination: null });
    setEditingDestination(false);
  }

  function renderEditDestinationDialog() {
    const currentDestinationType = data?.destinationTypes.find(
      (st) =>
        st.metadata.name === formValues.destination?.resourceConfiguration.type
    );

    if (currentDestinationType == null || formValues.destination == null) {
      return null;
    }

    return (
      <EditResourceDialog
        fullWidth
        maxWidth="sm"
        title={formValues.destination.resourceConfiguration.name!}
        description={currentDestinationType.metadata.description ?? ""}
        parameterDefinitions={currentDestinationType.spec.parameters}
        parameters={
          formValues.destination?.resourceConfiguration.parameters ?? []
        }
        open={editingDestination}
        onClose={() => setEditingDestination(false)}
        onCancel={() => {
          setEditingDestination(false);
        }}
        onDelete={() => setConfirmDeleteOpen(true)}
        onSave={onEditDestinationSave}
        kind={"destination"}
      />
    );
  }

  function renderDestinationAccordion() {
    if (formValues.destination == null) {
      return null;
    }

    const destinationConfig = new BPResourceConfiguration(
      formValues.destination!.resourceConfiguration
    );

    const destinationType = data?.destinationTypes.find(
      (dt) => dt.metadata.name === destinationConfig.type
    );

    const icon = destinationType?.metadata.icon;
    return (
      /* ---------------------------- Accordion Header ---------------------------- */
      <Accordion data-testid="destination-accordion">
        <AccordionSummary>
          <Stack direction={"row"} alignItems="center" spacing={1}>
            <span
              className={styles.icon}
              style={{ backgroundImage: `url(${icon})` }}
            />
            <Typography fontWeight={600}>{destinationConfig.name}</Typography>
          </Stack>
        </AccordionSummary>

        {/* --------------------------- Accordion Dropdown --------------------------- */}
        <AccordionDetails>
          {destinationConfig.hasConfigurationParameters() ? (
            <Table>
              <TableBody>
                {destinationConfig.parameters!.map((p) => {
                  const label =
                    destinationType?.spec.parameters.find(
                      (param) => param.name === p.name
                    )?.label ?? p.name;
                  return (
                    <TableRow key={p.name}>
                      <TableCell>{label}</TableCell>
                      <TableCell classes={{ root: styles["break-word-cell"] }}>
                        {String(p.value)}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          ) : (
            <Typography>No configuration.</Typography>
          )}

          {/* ------------------------- Edit and Remove Buttons ------------------------ */}
          <Stack
            direction="row"
            className={classes([mixins["float-right"], mixins["my-2"]])}
          >
            <Button color="error" onClick={() => setConfirmDeleteOpen(true)}>
              Remove
            </Button>

            {formValues.destination.create && (
              // You can only edit a destination you just created from the wizard.
              <Button
                onClick={() => setEditingDestination(true)}
                classes={{
                  root: classes([mixins["float-right"], mixins["my-2"]]),
                }}
              >
                Edit
              </Button>
            )}
          </Stack>
        </AccordionDetails>
      </Accordion>
    );
  }

  return (
    <>
      <div className={styles.container} data-testid="step-three">
        <Typography variant="h6" marginBottom="2rem">
          Add a destination you'd like to send your metrics and logs to
        </Typography>
        <Typography variant="body2" marginBottom={"1rem"}>
          A destination simply represents where you'd like to send your
          telemetry data. You can configure that here. Depending on the
          destination you choose, we'll configure specific OTel processors for
          you, ensuring your data shows up in a useful state.
        </Typography>

        {/* ------------------- Add Destination button or Accordion ------------------ */}
        <div>
          {formValues.destination == null ? (
            <Button
              variant="contained"
              endIcon={<PlusCircleIcon />}
              onClick={() => setAddDestinationOpen(true)}
              data-testid="add-destination-button"
            >
              Add Destination
            </Button>
          ) : (
            renderDestinationAccordion()
          )}
        </div>
      </div>

      {/* ---------------------- Back and Save footer buttons ---------------------- */}
      <Stack direction="row" justifyContent={"space-between"}>
        <Button
          variant="outlined"
          color="secondary"
          onClick={() => goToStep(1)}
        >
          Back
        </Button>
        <Button
          data-testid="save-button"
          variant="contained"
          onClick={onSaveConfiguration}
        >
          Save
        </Button>
      </Stack>

      {renderEditDestinationDialog()}

      <ConfirmDeleteResourceDialog
        open={confirmDeleteOpen}
        onDelete={onDestinationDelete}
        onCancel={() => setConfirmDeleteOpen(false)}
        action="remove"
      >
        <Typography>
          Are you sure you want to remove this destination?
        </Typography>
      </ConfirmDeleteResourceDialog>

      <ResourceDialog
        title="Choose a Destination"
        kind="destination"
        open={addDestinationOpen}
        onClose={() => setAddDestinationOpen(false)}
        onSaveNew={onNewDestinationSave}
        onSaveExisting={onChooseExistingDestination}
        resourceTypes={data?.destinationTypes ?? []}
        resources={data?.destinations}
      />
    </>
  );
};
