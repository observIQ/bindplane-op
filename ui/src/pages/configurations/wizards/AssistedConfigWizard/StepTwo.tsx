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
import { useWizard } from "../../../../components/Wizard/WizardContext";
import { PlusCircleIcon } from "../../../../components/Icons";
import { ResourceDialog } from "../../../../components/ResourceDialog";
import { useState } from "react";
import { gql } from "@apollo/client";
import {
  DestinationType,
  ParameterDefinition,
  ResourceConfiguration,
  SourceType,
  useSourceTypesQuery,
} from "../../../../graphql/generated";
import { AssistedWizardFormValues } from ".";
import { EditResourceDialog } from "../../../../components/EditResourceDialog";
import { ConfirmDeleteResourceDialog } from "../../../../components/ConfirmDeleteResourceDialog";
import { classes } from "../../../../utils/styles";
import { BPResourceConfiguration } from "../../../../utils/classes/resource-configuration";

import styles from "./assisted-config-wizard.module.scss";
import mixins from "../../../../styles/mixins.module.scss";

type ResourceType = SourceType | DestinationType;

gql`
  query sourceTypes {
    sourceTypes {
      apiVersion
      kind
      metadata {
        id
        name
        displayName
        description
        icon
      }
      spec {
        parameters {
          name
          label
          description
          relevantIf {
            name
            operator
            value
          }
          required
          type
          validValues
          default
        }
        supportedPlatforms
        version
        telemetryTypes
      }
    }
  }
`;

export const StepTwo: React.FC = (props) => {
  const { formValues, setValues, goToStep } =
    useWizard<AssistedWizardFormValues>();

  const [open, setOpen] = useState(false);
  const [editingSourceIx, setEditingSourceIx] = useState(-1);
  const [removeModalOpen, setRemoveModalOpen] = useState(false);

  const { data } = useSourceTypesQuery();

  function onSave(values: { [name: string]: any }, sourceType: ResourceType) {
    const sourceConfig = new BPResourceConfiguration();
    sourceConfig.setParamsFromMap(values);
    sourceConfig.type = sourceType.metadata.name;

    const sources = [...formValues.sources, sourceConfig];
    setValues({ sources: sources });
    setOpen(false);
  }

  function onEditSourceSave(values: { [key: string]: any }) {
    const newSource = new BPResourceConfiguration(
      formValues.sources[editingSourceIx]
    );

    // Replace the parameters with edited values
    newSource.setParamsFromMap(values);

    const newSources = [...formValues.sources];
    newSources[editingSourceIx] = newSource;

    setValues({ sources: newSources });
    setEditingSourceIx(-1);
  }

  function deleteSelectedSource() {
    const newSources = [...formValues.sources];
    newSources.splice(editingSourceIx, 1);

    setValues({ sources: newSources });
  }

  function onSourceRemove() {
    setRemoveModalOpen(false);
    deleteSelectedSource();
    setEditingSourceIx(-1);
  }

  function renderSourceAccordion(
    s: ResourceConfiguration,
    ix: number
  ): JSX.Element {
    const sourceType = data?.sourceTypes.find(
      (st: SourceType) => st.metadata.name === s.type
    );

    if (sourceType == null) {
      // TODO (dsvanlani) error toast and exit
      return <></>;
    }
    const displayName = sourceType.metadata.displayName;
    const icon = sourceType.metadata.icon;

    return (
      <Accordion key={s.type} data-testid="source-accordion">
        <AccordionSummary>
          <Stack direction={"row"} alignItems="center" spacing={1}>
            <span
              className={styles.icon}
              style={{ backgroundImage: `url(${icon})` }}
            />
            <Typography fontWeight={600}>{displayName}</Typography>
          </Stack>
        </AccordionSummary>
        <AccordionDetails>
          <Table>
            <TableBody>
              {s.parameters?.map((param) => {
                const label =
                  sourceType.spec.parameters.find(
                    (def: ParameterDefinition) => def.name === param.name
                  )?.label ?? param.name;

                if (param.value == null) return null;
                return (
                  <TableRow key={param.name}>
                    <TableCell>{label}</TableCell>
                    <TableCell classes={{ root: styles["break-word-cell"] }}>
                      {String(param.value)}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
          <Stack
            direction="row"
            className={classes([mixins["float-right"], mixins["my-2"]])}
          >
            <Button
              color="error"
              onClick={() => {
                setRemoveModalOpen(true);
              }}
            >
              Delete
            </Button>
            <Button onClick={() => setEditingSourceIx(ix)}>Edit</Button>
          </Stack>
        </AccordionDetails>
      </Accordion>
    );
  }

  function openResourceDialog() {
    setOpen(true);
  }

  return (
    <>
      <div className={styles.container} data-testid="step-two">
        {/* ---------------------------------- Copy ---------------------------------- */}
        <Typography variant="h6" classes={{ root: mixins["mb-5"] }}>
          Add sources from which you'd like to collect metrics and logs{" "}
        </Typography>
        <Typography variant="body2" classes={{ root: mixins["mb-3"] }}>
          A source is a combination of OpenTelemetry receivers and processors
          that allows you to collect metrics and logs from a specific
          technology. Ensuring the right combination of these components is one
          of the most challenging aspects of building an OTel configuration
          file. With BindPlane, we handle that all for you.
        </Typography>
        <Typography variant="body2" classes={{ root: mixins["mb-3"] }}>
          Start adding sources now:
        </Typography>

        <div>
          <div className={mixins["mb-3"]}>
            {formValues.sources.map((s, ix) => renderSourceAccordion(s, ix))}
          </div>
          <Button
            variant="contained"
            endIcon={<PlusCircleIcon />}
            onClick={openResourceDialog}
          >
            Add Source
          </Button>

          <ResourceDialog
            title="Choose a Source"
            kind="source"
            open={open}
            onClose={() => setOpen(false)}
            resourceTypes={data?.sourceTypes ?? []}
            onSaveNew={onSave}
          />
        </div>
      </div>

      <EditResourceDialog
        fullWidth
        maxWidth="sm"
        parameters={formValues.sources[editingSourceIx]?.parameters ?? []}
        open={editingSourceIx !== -1}
        onClose={() => setEditingSourceIx(-1)}
        onCancel={() => {
          setEditingSourceIx(-1);
        }}
        onDelete={() => setRemoveModalOpen(true)}
        onSave={onEditSourceSave}
        parameterDefinitions={
          data?.sourceTypes.find(
            (s: SourceType) =>
              formValues.sources[editingSourceIx]?.type === s.metadata.name
          )?.spec.parameters ?? []
        }
        title={""}
        description={""}
        kind={"source"}
      />

      <ConfirmDeleteResourceDialog
        open={removeModalOpen}
        onDelete={onSourceRemove}
        onCancel={() => setRemoveModalOpen(false)}
        action="remove"
      >
        <Typography>Are you sure you want to remove this source?</Typography>
      </ConfirmDeleteResourceDialog>

      <Stack direction={"row"} justifyContent="space-between">
        <Button
          variant="outlined"
          color="secondary"
          onClick={() => goToStep(0)}
        >
          Back
        </Button>
        <Button variant="contained" onClick={() => goToStep(2)}>
          Next
        </Button>
      </Stack>
    </>
  );
};
