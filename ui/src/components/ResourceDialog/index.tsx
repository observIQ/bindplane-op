import {
  Button,
  Dialog,
  DialogContent,
  DialogProps,
  Stack,
  Typography,
} from "@mui/material";
import { useEffect, useState } from "react";
import { ResourceConfigForm } from "../ResourceTypeForm";
import { ResourceButton } from "./ResourceButton";
import {
  DestinationType,
  SourceType,
  Parameter,
  Maybe,
} from "../../graphql/generated";
import { isFunction, sortBy } from "lodash";

import mixins from "../../styles/mixins.module.scss";

type ResourceType = SourceType | DestinationType;

export type DialogResource = {
  metadata: {
    name: string;
  };
  spec: { type: string; parameters?: Maybe<Parameter[]> };
};

interface ResourceDialogProps extends DialogProps {
  // Displayed on the first step of the Dialog.
  title: string;

  kind: "destination" | "source";

  // Either SourceType[] or DestinationType[]
  resourceTypes: ResourceType[];

  // If present, it will allow users to choose existing Resources
  resources?: DialogResource[];

  // Callback for the save button when creating a new Resource.
  onSaveNew?: (
    parameters: { [key: string]: any },
    resourceType: ResourceType
  ) => void;

  // Callback for saving an existing resource.
  onSaveExisting?: (resource: DialogResource) => void;
}

// There are three possible views to this dialog -
// 1. The "Select" view is the first view, which displays all available ResourceTypes.
//    This is displayed anytime selected == null
//    To go do this step its sufficient to set selected to null.
//
// 2. The "Choose" view is an optional second view, shown after the Select view
//    and there are existing resources from the resources prop
//    such that resource.spec.type === selected.metadata.name
//
// 3. The "Configure" view.  The final step for creation, contains the ResourceTypeForm
//    This is displayed when selected != null and createNew == true.
//
export const ResourceDialog: React.FC<ResourceDialogProps> = ({
  title,
  kind,
  resourceTypes,
  resources,
  onSaveNew,
  onSaveExisting,
  ...dialogProps
}) => {
  const [selected, setSelected] = useState<ResourceType | null>(null);
  const [createNew, setCreateNew] = useState(false);

  // This resets the form after close.
  useEffect(() => {
    let timer: ReturnType<typeof setTimeout>;
    if (dialogProps.open === false) {
      timer = setTimeout(() => clearResource(), 500);
    }

    return () => clearTimeout(timer);
  }, [dialogProps.open]);

  function clearResource() {
    setSelected(null);
    setCreateNew(false);
  }

  function handleSaveNew(
    parameters: { [key: string]: any },
    resourceType: ResourceType
  ) {
    isFunction(onSaveNew) && onSaveNew(parameters, resourceType);
    clearResource();
  }

  function handleSaveExisting(resource: DialogResource) {
    isFunction(onSaveExisting) && onSaveExisting(resource);
    clearResource();
  }

  function renderContent() {
    if (selected == null) {
      // Step one is to select a ResourceType
      return renderSelectView();
    } else if (
      resources?.some((r) => r.spec.type === selected.metadata.name) &&
      !createNew
    ) {
      // There are existing resources that match the selected type.
      // Show the "Choose View" to allow users to pick existing resources.
      return renderChooseView();
    } else {
      // Show the form to configure a new source
      return renderConfigureView();
    }
  }

  function renderSelectView() {
    // Sort the list by display name.
    // .sort mutates the array, and since resourceTypes
    // is a prop so we need to copy it
    const copy = resourceTypes.slice();
    const sortedTypes = copy.sort((a, b) =>
      a.metadata
        .displayName!.toLowerCase()
        .localeCompare(b.metadata.displayName!.toLowerCase())
    );

    return (
      <>
        <Typography variant="h6" className={mixins["mb-5"]}>
          {title}
        </Typography>
        <Stack spacing={1}>
          {sortedTypes.map((resourceType) => {
            const matchingResourcesExist = resources?.some(
              (resource) => resource.spec.type === resourceType.metadata.name
            );

            // Either we send the directly to the form if there are no existing resources
            // of that type, or we send them to the Choose View by just setting the selected.
            function onSelect() {
              setSelected(resourceType);
              if (!matchingResourcesExist) {
                setCreateNew(true);
              }
            }
            return (
              <ResourceButton
                key={resourceType.metadata.name}
                icon={resourceType.metadata.icon!}
                displayName={resourceType.metadata.displayName!}
                onSelect={onSelect}
                telemetryTypes={resourceType.spec.telemetryTypes}
              />
            );
          })}
        </Stack>
      </>
    );
  }

  function renderChooseView() {
    const matchingResources = resources?.filter(
      (r) => r.spec.type === selected!.metadata.name
    );

    return (
      <>
        <Typography variant="h6" className={mixins["mb-5"]}>
          Choose an existing {kind} or create a new one
        </Typography>

        <Stack spacing={1}>
          {matchingResources?.map((resource) => {
            return (
              <ResourceButton
                key={resource.metadata.name}
                icon={selected?.metadata.icon!}
                displayName={resource.metadata.name}
                onSelect={() => handleSaveExisting(resource)}
              />
            );
          })}
          <Button
            variant="contained"
            color="primary"
            onClick={() => setCreateNew(true)}
          >
            Create New
          </Button>
        </Stack>

        <Button
          variant="contained"
          color="secondary"
          classes={{ root: mixins["mt-3"] }}
          onClick={clearResource}
        >
          Back
        </Button>
      </>
    );
  }

  function renderConfigureView() {
    if (selected === null) {
      return <></>;
    }

    return (
      <ResourceConfigForm
        kind={kind}
        includeNameField={kind === "destination" && createNew}
        existingResourceNames={resources?.map((r) => r.metadata.name)}
        onBack={clearResource}
        onSave={(fv) => handleSaveNew(fv, selected)}
        title={selected.metadata.displayName ?? ""}
        description={selected.metadata.description ?? ""}
        parameterDefinitions={selected.spec.parameters ?? []}
      />
    );
  }

  return (
    <Dialog
      {...dialogProps}
      fullWidth
      maxWidth="sm"
      data-testid="resource-dialog"
    >
      <DialogContent>{renderContent()}</DialogContent>
    </Dialog>
  );
};
