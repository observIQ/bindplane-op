import { Grid, Button, Typography } from "@mui/material";
import { isFunction } from "lodash";
import { ParameterDefinition } from "../../graphql/generated";
import { classes } from "../../utils/styles";
import {
  ButtonFooter,
  FormTitle,
  ParameterInput,
  ResourceNameInput,
  satisfiesRelevantIf,
  useValidationContext,
  isValid,
} from ".";
import { InlineProcessorContainer } from "./InlineProcessorContainer";

import mixins from "../../styles/mixins.module.scss";

interface MainProps {
  title: string;
  description: string;
  kind: "source" | "destination" | "processor";
  formValues: { [key: string]: any };
  setFormValues: React.Dispatch<
    React.SetStateAction<{
      [key: string]: any;
    }>
  >;
  includeNameField?: boolean;
  existingResourceNames?: string[];
  parameterDefinitions: ParameterDefinition[];
  enableProcessors?: boolean;
  onBack?: () => void;
  onSave?: (formValues: { [key: string]: any }) => void;
  saveButtonLabel?: string;
  onDelete?: () => void;
  onAddProcessor: () => void;
  onEditProcessor: (editingIndex: number) => void;
  onRemoveProcessor: (removeIndex: number) => void;
  disableSave?: boolean;
}

export const MainView: React.FC<MainProps> = ({
  title,
  description,
  kind,
  formValues,
  includeNameField,
  setFormValues,
  existingResourceNames,
  parameterDefinitions,
  enableProcessors,
  onBack,
  onSave,
  saveButtonLabel,
  onDelete,
  onAddProcessor,
  onEditProcessor,
  disableSave,
}) => {
  const { errors } = useValidationContext();

  function renderParameter(p: ParameterDefinition) {
    const onValueChange = (v: string) => {
      setFormValues((prev) => ({ ...prev, [p.name]: v }));
    };

    if (satisfiesRelevantIf(formValues, p)) {
      return (
        <Grid key={p.name} item>
          <ParameterInput
            definition={p}
            value={formValues[p.name]}
            onValueChange={onValueChange}
          />
        </Grid>
      );
    }

    return null;
  }

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    isFunction(onSave) && onSave(formValues);
  }

  const primaryButton: JSX.Element = (
    <Button
      disabled={!isValid(errors) || disableSave}
      type="submit"
      variant="contained"
      data-testid="resource-form-save"
    >
      {saveButtonLabel ?? "Save"}
    </Button>
  );

  const backButton: JSX.Element = (
    <Button variant="contained" color="secondary" onClick={onBack}>
      Back
    </Button>
  );

  const deleteButton: JSX.Element | undefined = isFunction(onDelete) ? (
    <Button
      variant="outlined"
      color="error"
      onClick={onDelete}
      classes={{ root: mixins["mr-2"] }}
    >
      Delete
    </Button>
  ) : undefined;

  return (
    <>
      <FormTitle title={title} description={description} />

      <form onSubmit={handleSubmit} data-testid="resource-form">
        <Typography fontWeight={600} marginBottom={2}>
          Configure
        </Typography>
        <Grid
          container
          direction={"column"}
          spacing={3}
          className={classes([mixins["form-width"], mixins["mb-5"]])}
        >
          {includeNameField && (
            <Grid item>
              <ResourceNameInput
                kind={kind}
                value={formValues.name}
                onValueChange={(v: string) =>
                  setFormValues((prev) => ({ ...prev, name: v }))
                }
                existingNames={existingResourceNames}
              />
            </Grid>
          )}
          {parameterDefinitions.length === 0 ? (
            <Grid item>
              <Typography>No additional configuration needed.</Typography>
            </Grid>
          ) : (
            parameterDefinitions.map((p) => renderParameter(p))
          )}
        </Grid>

        {enableProcessors && (
          <InlineProcessorContainer
            processors={formValues.processors ?? []}
            onAddProcessor={onAddProcessor}
            onEditProcessor={onEditProcessor}
            setFormValues={setFormValues}
          />
        )}

        <ButtonFooter
          backButton={backButton}
          secondaryButton={deleteButton ?? <></>}
          primaryButton={primaryButton}
        />
      </form>
    </>
  );
};
