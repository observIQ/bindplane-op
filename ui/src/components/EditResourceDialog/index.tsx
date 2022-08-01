import { Dialog, DialogContent, DialogProps } from "@mui/material";
import {
  Maybe,
  Parameter,
  ParameterDefinition,
  ResourceConfiguration,
} from "../../graphql/generated";
import { ResourceConfigForm } from "../ResourceConfigForm";

interface EditResourceBaseProps extends DialogProps {
  onSave: (values: { [key: string]: any }) => void;
  onDelete?: () => void;
  onCancel: () => void;
  parameters: Parameter[];
  parameterDefinitions: ParameterDefinition[];
  processors?: Maybe<ResourceConfiguration[]>;
  enableProcessors?: boolean;
  title: string;
  description: string;
  includeNameField?: boolean;
  kind: "source" | "destination";
}

export const EditResourceDialog: React.FC<EditResourceBaseProps> = ({
  onSave,
  onDelete,
  onCancel,
  parameters,
  processors,
  enableProcessors,
  title,
  parameterDefinitions,
  description,
  kind,
  includeNameField = false,
  ...dialogProps
}) => {
  return (
    <Dialog {...dialogProps} onClose={onCancel}>
      <DialogContent>
        <ResourceConfigForm
          includeNameField={includeNameField}
          title={title}
          description={description}
          kind={kind}
          parameterDefinitions={parameterDefinitions}
          parameters={parameters}
          processors={processors}
          enableProcessors={enableProcessors}
          onSave={onSave}
          onDelete={onDelete}
          onBack={onCancel}
        />
      </DialogContent>
    </Dialog>
  );
};
