import { Button, Stack, Typography } from "@mui/material";
import { memo, useState } from "react";
import { CardContainer } from "../../../components/CardContainer";
import { PlusCircleIcon } from "../../../components/Icons";
import {
  Configuration,
  DestinationType,
  ResourceConfiguration,
  SourceType,
  useSourceTypesQuery,
} from "../../../graphql/generated";
import { SourceCard } from "./SourceCard";
import { ResourceDialog } from "../../../components/ResourceDialog";
import { applyResources } from "../../../utils/rest/apply-resources";
import { EditResourceDialog } from "../../../components/EditResourceDialog";
import { ConfirmDeleteResourceDialog } from "../../../components/ConfirmDeleteResourceDialog";
import { useSnackbar } from "notistack";
import { ShowPageConfig } from ".";
import { cloneIntoConfig } from "./utils";
import {
  BPConfiguration,
  BPResourceConfiguration,
} from "../../../utils/classes";
import { UpdateStatus } from "../../../types/resources";

import styles from "./configuration-page.module.scss";
import mixins from "../../../styles/mixins.module.scss";

type ResourceType = SourceType | DestinationType;

const SourcesSectionComponent: React.FC<{
  configuration: NonNullable<ShowPageConfig>;
  refetch: () => {};
}> = ({ configuration, refetch }) => {
  const sources = configuration.spec?.sources || [];

  const { enqueueSnackbar } = useSnackbar();

  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [editingSourceIx, setEditingSourceIx] = useState<number>(-1);
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);

  const { data } = useSourceTypesQuery();

  async function onNewSourceSave(
    values: { [key: string]: any },
    sourceType: ResourceType
  ) {
    const newSourceConfig = new BPResourceConfiguration();
    newSourceConfig.type = sourceType.metadata.name;
    newSourceConfig.setParamsFromMap(values);

    const updatedConfig = new BPConfiguration(configuration);
    updatedConfig.addSource(newSourceConfig);
    try {
      const update = await updatedConfig.apply();
      if (update.status === UpdateStatus.INVALID) {
        console.error(update);
        throw new Error("failed to add source to configuration.");
      }

      setAddDialogOpen(false);
      refetch();
    } catch (err) {
      enqueueSnackbar("Failed to save source.", {
        variant: "error",
      });
      console.error(err);
    }
  }

  async function onEditSourceSave(values: { [key: string]: any }) {
    const sourceConfig = new BPResourceConfiguration(sources[editingSourceIx]);
    sourceConfig.setParamsFromMap(values);

    const updatedConfig = new BPConfiguration(configuration);
    updatedConfig.replaceSource(sourceConfig, editingSourceIx);
    try {
      const update = await updatedConfig.apply();
      if (update.status === UpdateStatus.INVALID) {
        console.error(update);
        throw new Error("failed to save source on configuration");
      }

      setEditingSourceIx(-1);
      refetch();
    } catch (err) {
      enqueueSnackbar("Failed to save source.", {
        variant: "error",
        autoHideDuration: 5000,
      });
      console.error(err);
    }
  }

  async function onEditSourceDelete() {
    setConfirmDeleteOpen(false);

    const newSources = [...sources];
    newSources.splice(editingSourceIx, 1);

    // Copy the configuration with the new sources
    const newConfig = cloneIntoConfig(configuration) as Configuration;

    newConfig.spec.sources = newSources;

    // Apply the new configuration
    try {
      await applyResources([newConfig]);
      refetch();
      setEditingSourceIx(-1);
    } catch (err) {
      console.error(err);
      enqueueSnackbar("Failed to delete source.", {
        variant: "error",
        autoHideDuration: 5000,
      });
    }
  }

  return (
    <>
      <CardContainer>
        <div className={styles["title-button-row"]}>
          <Typography variant="h5">Sources</Typography>
          <Button
            onClick={() => setAddDialogOpen(true)}
            variant="contained"
            classes={{ root: mixins["float-right"] }}
            startIcon={<PlusCircleIcon />}
          >
            Add Source
          </Button>
        </div>

        <Stack direction="row" spacing={2}>
          {sources.map((source, ix) => {
            function onClick() {
              setEditingSourceIx(ix);
            }

            return <SourceCard key={ix} source={source} onClick={onClick} />;
          })}
        </Stack>
      </CardContainer>

      <EditResourceDialog
        fullWidth
        maxWidth="sm"
        title={
          findSourceType(data?.sourceTypes ?? [], sources[editingSourceIx])
            ?.metadata.displayName ?? ""
        }
        description={
          findSourceType(data?.sourceTypes ?? [], sources[editingSourceIx])
            ?.metadata.description ?? ""
        }
        kind="source"
        parameterDefinitions={
          findSourceType(data?.sourceTypes ?? [], sources[editingSourceIx])
            ?.spec.parameters ?? []
        }
        parameters={sources[editingSourceIx]?.parameters ?? []}
        processors={sources[editingSourceIx]?.processors}
        enableProcessors
        open={editingSourceIx !== -1}
        onClose={() => setEditingSourceIx(-1)}
        onCancel={() => {
          setEditingSourceIx(-1);
        }}
        onDelete={() => setConfirmDeleteOpen(true)}
        onSave={onEditSourceSave}
      />

      <ConfirmDeleteResourceDialog
        open={confirmDeleteOpen}
        onDelete={onEditSourceDelete}
        onCancel={() => setConfirmDeleteOpen(false)}
        action="remove"
      >
        <Typography>Are you sure you want to remove this source?</Typography>
      </ConfirmDeleteResourceDialog>

      <ResourceDialog
        title={"Add a Source"}
        kind="source"
        resourceTypes={data?.sourceTypes ?? []}
        open={addDialogOpen}
        onSaveNew={onNewSourceSave}
        onClose={() => setAddDialogOpen(false)}
      />
    </>
  );
};

export const SourcesSection = memo(SourcesSectionComponent);

function findSourceType(
  sourceTypes: SourceType[],
  source?: ResourceConfiguration
) {
  if (source == null) return undefined;
  return sourceTypes.find((st) => st.metadata.name === source.type);
}
