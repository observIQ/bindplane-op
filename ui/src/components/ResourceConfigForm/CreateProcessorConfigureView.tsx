import { FormTitle, FormValues, ProcessorType, ResourceConfigForm } from ".";

interface CreateProcessorConfigureViewProps {
  title: string;
  processorType: ProcessorType;
  onBack: () => void;
  onSave: (formValues: FormValues) => void;
}

export const CreateProcessorConfigureView: React.FC<CreateProcessorConfigureViewProps> =
  ({ title, processorType, onSave, onBack }) => {
    return (
      <>
        <FormTitle title={title} crumbs={["Add a processor", "Configure"]} />
        <ResourceConfigForm
          title={""}
          description={""}
          kind={"processor"}
          parameterDefinitions={processorType.spec.parameters}
          onSave={onSave}
          saveButtonLabel={"Add Processor"}
          onBack={onBack}
        />
      </>
    );
  };
