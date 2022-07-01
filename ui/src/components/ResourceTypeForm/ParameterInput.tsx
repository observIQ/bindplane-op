import { FormControlLabel, Switch, TextField } from "@mui/material";
import { isArray, isFunction } from "lodash";
import { ChangeEvent } from "react";
import { ParameterDefinition, ParameterType } from "../../graphql/generated";
import { validateNameField } from "../../utils/forms/validate-name-field";
import { useValidationContext } from "./ValidationContext";

import styles from "./parameter-input.module.scss";
import { classes as classesUtil } from '../../utils/styles';

interface ParamInputProps {
  classes?: { [name: string]: string };
  definition: ParameterDefinition;
  value?: any;
  onValueChange?: (v: any) => void;
}

export const ParameterInput: React.FC<ParamInputProps> = (props) => {
  let classes = props.classes;
  if (props.definition.relevantIf != null) {
    classes = Object.assign(classes || {}, { root: classesUtil([classes?.root, styles.indent]) });
  }
  switch (props.definition.type) {
    case ParameterType.String:
      return <StringParamInput classes={classes} {...props} />;
    case ParameterType.Strings:
      return <StringsParamInput classes={classes} {...props} />;
    case ParameterType.Enum:
      return <EnumParamInput classes={classes} {...props} />;
    case ParameterType.Bool:
      return <BoolParamInput classes={classes} {...props} />;
    case ParameterType.Int:
      return <IntParamInput classes={classes} {...props} />;
  }
};

export const StringParamInput: React.FC<ParamInputProps> = ({
  classes,
  definition,
  value,
  onValueChange,
}) => {
  return (
    <TextField
      classes={classes}
      value={value}
      onChange={(e: ChangeEvent<HTMLInputElement>) =>
        isFunction(onValueChange) && onValueChange(e.target.value)
      }
      name={definition.name}
      fullWidth
      size="small"
      label={definition.label}
      helperText={definition.description}
      required={definition.required}
      autoComplete="off"
      autoCorrect="off"
      autoCapitalize="off"
      spellCheck="false"
    />
  );
};

export const EnumParamInput: React.FC<ParamInputProps> = ({
  classes,
  definition,
  value,
  onValueChange,
}) => {
  return (
    <TextField
      classes={classes}
      value={value}
      onChange={(e: ChangeEvent<HTMLInputElement>) =>
        isFunction(onValueChange) && onValueChange(e.target.value)
      }
      name={definition.name}
      fullWidth
      size="small"
      label={definition.label}
      helperText={definition.description}
      required={definition.required}
      select
      SelectProps={{ native: true }}
    >
      {definition.validValues?.map((v) => (
        <option key={v} value={v}>
          {v}
        </option>
      ))}
    </TextField>
  );
};

export const StringsParamInput: React.FC<ParamInputProps> = ({
  classes,
  definition,
  value: arrayValue,
  onValueChange,
}) => {
  // TODO (dsvanlani) This will not hold up very long, but for now save the state as an
  // array of strings split by comma.  This should eventually be a multi string input
  const value = isArray(arrayValue) ? arrayValue.join(",") : undefined;
  function onChange(e: ChangeEvent<HTMLInputElement>) {
    const newValue = e.target.value.split(",");
    isFunction(onValueChange) && onValueChange(newValue);
  }

  return (
    <TextField
      classes={classes}
      value={value}
      onChange={onChange}
      name={definition.name}
      fullWidth
      size="small"
      label={definition.label}
      helperText={definition.description}
      required={definition.required}
      autoComplete="off"
      autoCorrect="off"
      autoCapitalize="off"
      spellCheck="false"
    />
  );
};

export const BoolParamInput: React.FC<ParamInputProps> = ({
  classes,
  definition,
  value,
  onValueChange,
}) => {
  return (
    <FormControlLabel
      classes={classes}
      control={
        <Switch
          onChange={(e) => {
            isFunction(onValueChange) && onValueChange(e.target.checked);
          }}
          name={definition.name}
          checked={value}
        />
      }
      label={definition.label}
    />
  );
};

export const IntParamInput: React.FC<ParamInputProps> = ({
  classes,
  definition,
  value,
  onValueChange,
}) => {
  // TODO dsvanlani This should probably be a custom text input with validation
  return (
    <TextField
      classes={classes}
      value={value}
      onChange={(e: ChangeEvent<HTMLInputElement>) =>
        isFunction(onValueChange) && onValueChange(Number(e.target.value))
      }
      name={definition.name}
      fullWidth
      size="small"
      label={definition.label}
      helperText={definition.description}
      required={definition.required}
      autoComplete="off"
      autoCorrect="off"
      autoCapitalize="off"
      spellCheck="false"
      type={"number"}
    />
  );
};

interface ResourceNameInputProps extends Omit<ParamInputProps, "definition"> {
  existingNames?: string[];
  kind: "source" | "destination" | "configuration";
}

export const ResourceNameInput: React.FC<ResourceNameInputProps> = ({
  classes,
  value,
  onValueChange,
  existingNames,
  kind,
}) => {
  const { errors, setError, touched, touch } = useValidationContext();

  function handleChange(e: ChangeEvent<HTMLInputElement>) {
    if (!isFunction(onValueChange)) {
      return;
    }

    onValueChange(e.target.value);
    const error = validateNameField(e.target.value, kind, existingNames);
    setError("name", error);
  }

  return (
    <TextField
      classes={classes}
      onBlur={() => touch("name")}
      value={value}
      onChange={handleChange}
      inputProps={{
        "data-testid": "name-field",
      }}
      error={errors.name != null && touched.name}
      helperText={errors.name}
      color={errors.name != null ? "error" : "primary"}
      name={"name"}
      fullWidth
      size="small"
      label={"Name"}
      required
      autoComplete="off"
      autoCorrect="off"
      autoCapitalize="off"
      spellCheck="false"
    />
  );
};
