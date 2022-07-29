import { Maybe } from "graphql/jsutils/Maybe";
import { isEqual } from "lodash";
import { useState } from "react";
import {
  CreateProcessorConfigureView,
  CreateProcessorSelectView,
  EditProcessorView,
  initFormValues,
  MainView,
  ValidationContextProvider,
} from ".";
import {
  ParameterDefinition,
  Parameter,
  ResourceConfiguration,
  GetProcessorTypesQuery,
  ParameterType,
} from "../../graphql/generated";
import { BPResourceConfiguration } from "../../utils/classes";
import { validateStringsField, validateMapField } from "./validation-functions";

enum Page {
  MAIN,
  CREATE_PROCESSOR_SELECT,
  CREATE_PROCESSOR_CONFIGURE,
  EDIT_PROCESSOR,
}

export type ProcessorType = GetProcessorTypesQuery["processorTypes"][0];

export interface FormValues {
  // The name of the Source or Destination
  name?: string;
  // The values for the Parameters
  [key: string]: any;
  // The inline processors configured for the Source or Destination
  processors?: ResourceConfiguration[];
}

interface ResourceFormProps {
  // Display name for the resource
  title: string;

  description: string;

  // Used to determine some form values.
  kind: "destination" | "source" | "processor";

  parameterDefinitions: ParameterDefinition[];

  // If present the form will use these values as defaults
  parameters?: Maybe<Parameter[]>;

  // If present the form will have a name field at the top and will be sent
  // as the formValues["name"] key.
  includeNameField?: boolean;

  // Used to validate the name field if includeNameField is present.
  existingResourceNames?: string[];

  // Any inline processors for the resource, only applies to Sources
  processors?: Maybe<ResourceConfiguration[]>;

  // If true will allow the form to add inline processors to the resource.
  enableProcessors?: boolean;

  // If present the form will display a "delete" button which calls
  // the onDelete callback.
  onDelete?: () => void;

  // The callback when the resource is saved.
  onSave?: (formValues: { [key: string]: any }) => void;
  // The copy on the primary button, defaults to "Save"
  saveButtonLabel?: string;

  // The callback when cancel is clicked.
  onBack?: () => void;
}

interface ComponentProps extends ResourceFormProps {
  initValues: Record<string, any>;
}

const ResourceConfigurationFormComponent: React.FC<ComponentProps> = ({
  title,
  description,
  parameters,
  parameterDefinitions,
  enableProcessors,
  includeNameField,
  existingResourceNames,
  kind,
  onDelete,
  onSave,
  saveButtonLabel,
  onBack,
  initValues,
}) => {
  const [formValues, setFormValues] = useState<FormValues>(initValues);

  // This is passed down to determine whether to enable the primary save button.
  // If no parameters are passed down, then the form is new and is "dirty".
  const isDirty = parameters == null || !isEqual(initValues, formValues);

  const [page, setPage] = useState<Page>(Page.MAIN);
  const [newProcessorType, setNewProcessorType] =
    useState<ProcessorType | null>(null);
  const [editingProcessorIndex, setEditingProcessorIndex] =
    useState<number>(-1);

  function handleAddProcessor() {
    setPage(Page.CREATE_PROCESSOR_SELECT);
  }

  function handleReturnToMain() {
    setPage(Page.MAIN);
    setNewProcessorType(null);
    setEditingProcessorIndex(-1);
  }

  function handleSelectNewProcessor(pt: ProcessorType) {
    setPage(Page.CREATE_PROCESSOR_CONFIGURE);
    setNewProcessorType(pt);
  }

  function handleEditProcessorClick(editingIndex: number) {
    setEditingProcessorIndex(editingIndex);
    setPage(Page.EDIT_PROCESSOR);
  }

  function handleEditProcessorSave(processorFormValues: FormValues) {
    const processorConfig = new BPResourceConfiguration();
    processorConfig.setParamsFromMap(processorFormValues);
    processorConfig.type = formValues.processors![editingProcessorIndex].type;

    // Replace the processor at index
    const newProcessors = [...(formValues.processors ?? [])];
    if (newProcessors[editingProcessorIndex] != null) {
      newProcessors[editingProcessorIndex] = processorConfig;
    } else {
      newProcessors.push(processorConfig);
    }

    setFormValues((prev) => ({ ...prev, processors: newProcessors }));
    setPage(Page.MAIN);
  }

  function handleNewProcessorSave(processorFormValues: FormValues) {
    const processorConfig = new BPResourceConfiguration();
    processorConfig.setParamsFromMap(processorFormValues);
    processorConfig.type = newProcessorType!.metadata.name;

    const newProcessors = [...(formValues.processors ?? [])];
    newProcessors.push(processorConfig);

    setFormValues((prev) => ({ ...prev, processors: newProcessors }));
    setPage(Page.MAIN);
  }

  function handleRemoveProcessor(removeIndex: number) {
    const newProcessors = [...(formValues.processors ?? [])];
    newProcessors.splice(removeIndex, 1);

    setFormValues((prev) => ({ ...prev, processors: newProcessors }));
    setPage(Page.MAIN);
    setEditingProcessorIndex(-1);
  }

  switch (page) {
    case Page.MAIN:
      return (
        <MainView
          title={title}
          description={description}
          kind={kind}
          formValues={formValues}
          includeNameField={includeNameField}
          setFormValues={setFormValues}
          existingResourceNames={existingResourceNames}
          parameterDefinitions={parameterDefinitions}
          enableProcessors={enableProcessors}
          onBack={onBack}
          onSave={onSave}
          saveButtonLabel={saveButtonLabel}
          onDelete={onDelete}
          onAddProcessor={handleAddProcessor}
          onEditProcessor={handleEditProcessorClick}
          onRemoveProcessor={handleRemoveProcessor}
          disableSave={!isDirty}
        />
      );
    case Page.CREATE_PROCESSOR_SELECT:
      return (
        <CreateProcessorSelectView
          title={title}
          onBack={handleReturnToMain}
          onSelect={handleSelectNewProcessor}
        />
      );
    case Page.CREATE_PROCESSOR_CONFIGURE:
      return (
        <CreateProcessorConfigureView
          onBack={handleReturnToMain}
          onSave={handleNewProcessorSave!}
          title={title}
          processorType={newProcessorType!}
        />
      );
    case Page.EDIT_PROCESSOR:
      return (
        <EditProcessorView
          title={title}
          processors={formValues.processors!}
          editingIndex={editingProcessorIndex}
          onEditProcessorSave={handleEditProcessorSave!}
          onRemove={handleRemoveProcessor}
          onBack={handleReturnToMain}
        />
      );
  }
};

export const ResourceConfigForm: React.FC<ResourceFormProps> = (props) => {
  const { parameterDefinitions, parameters, processors, includeNameField } =
    props;
  const initValues = initFormValues(
    parameterDefinitions,
    parameters,
    processors,
    includeNameField
  );

  // Get initial errors
  const initErrors: Record<string, string | null> = {};
  for (const definition of props.parameterDefinitions) {
    switch (definition.type) {
      case ParameterType.Strings:
        initErrors[definition.name] = validateStringsField(
          initValues[definition.name],
          definition.required
        );
        break;
      case ParameterType.Map:
        initErrors[definition.name] = validateMapField(
          initValues[definition.name],
          definition.required
        );
        break;
      default:
        initErrors[definition.name] = null;
    }
  }

  return (
    <ValidationContextProvider initErrors={initErrors}>
      <ResourceConfigurationFormComponent initValues={initValues} {...props} />
    </ValidationContextProvider>
  );
};
