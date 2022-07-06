import {
  Autocomplete,
  Chip,
  FormControlLabel,
  Switch,
  TextField,
} from "@mui/material";
import { isFunction } from "lodash";
import { ChangeEvent, useState } from "react";
import { ParameterDefinition, ParameterType } from "../../graphql/generated";
import { validateNameField } from "../../utils/forms/validate-name-field";
import { useValidationContext } from "./ValidationContext";
import { classes as classesUtil } from "../../utils/styles";

import styles from "./parameter-input.module.scss";

interface ParamInputProps {
  classes?: { [name: string]: string };
  definition: ParameterDefinition;
  value?: any;
  onValueChange?: (v: any) => void;
}

export const ParameterInput: React.FC<ParamInputProps> = (props) => {
  let classes = props.classes;
  if (props.definition.relevantIf != null) {
    classes = Object.assign(classes || {}, {
      root: classesUtil([classes?.root, styles.indent]),
    });
  }
  switch (props.definition.type) {
    case ParameterType.String:
      return <StringParamInput classes={classes} {...props} />;
    case ParameterType.Strings:
      return <StringsInput classes={classes} {...props} />;
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

export const StringsInput: React.FC<ParamInputProps> = ({
  classes,
  definition,
  value,
  onValueChange,
}) => {
  const [inputValue, setInputValue] = useState("");
  const options = inputValue === "" ? [] : [inputValue];

  return (
    <Autocomplete
      classes={classes}
      value={value}
      onChange={(e, v: string[]) => onValueChange && onValueChange(v)}
      onInputChange={(e, newValue) => setInputValue(newValue)}
      inputValue={inputValue}
      multiple
      options={options}
      id="tags-filled"
      freeSolo
      renderTags={(value: readonly string[], getTagProps) =>
        value.map((option: string, index: number) => (
          <Chip
            variant="outlined"
            label={option}
            {...getTagProps({ index })}
            classes={{ label: styles.chip }}
          />
        ))
      }
      renderInput={(params) => (
        <TextField {...params} label={definition.label} size={"small"} />
      )}
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
