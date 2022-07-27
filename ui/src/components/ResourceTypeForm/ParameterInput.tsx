import {
  Autocomplete,
  Button,
  Chip,
  FormControl,
  FormControlLabel,
  FormHelperText,
  Grid,
  InputLabel,
  OutlinedInput,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { isArray, isEmpty, isFunction } from "lodash";
import { ChangeEvent, useMemo, useState } from "react";
import { ParameterDefinition, ParameterType } from "../../graphql/generated";
import { validateNameField } from "../../utils/forms/validate-name-field";
import { useValidationContext } from "./ValidationContext";
import { classes as classesUtil } from "../../utils/styles";
import { YamlEditor } from "../YamlEditor";
import { PlusCircleIcon } from "../Icons";
import { validateStringsField } from "./validation-functions";

import styles from "./parameter-input.module.scss";

interface ParamInputProps<T> {
  classes?: { [name: string]: string };
  definition: ParameterDefinition;
  value?: T;
  onValueChange?: (v: T) => void;
}

export const ParameterInput: React.FC<ParamInputProps<any>> = (props) => {
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
    case ParameterType.Enums:
      return <EnumsParamInput classes={classes} {...props} />;
    case ParameterType.Bool:
      return <BoolParamInput classes={classes} {...props} />;
    case ParameterType.Int:
      return <IntParamInput classes={classes} {...props} />;
    case ParameterType.Map:
      return <MapParamInput classes={classes} {...props} />;
    case ParameterType.Yaml:
      return <YamlParamInput classes={classes} {...props} />;
  }
};

export const StringParamInput: React.FC<ParamInputProps<string>> = ({
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

export const EnumParamInput: React.FC<ParamInputProps<string>> = ({
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

export const EnumsParamInput: React.FC<ParamInputProps<string[]>> = ({
  classes,
  definition,
  value,
  onValueChange,
}) => {
  function handleToggleValue(
    event: ChangeEvent<HTMLInputElement>,
    checked: boolean,
    toggleValue: string
  ) {
    const newValue = [...(value ?? [])];
    if (checked) {
      // Make sure that toggleValue is in new value array
      if (!newValue.includes(toggleValue)) {
        newValue.push(toggleValue);
      }
    } else {
      // Remove the toggle value from the array
      const atIndex = newValue.findIndex((v) => v === toggleValue);
      if (atIndex > -1) {
        newValue.splice(atIndex, 1);
      }
    }

    onValueChange && onValueChange(newValue);
  }

  return (
    <>
      <InputLabel>{definition.label}</InputLabel>
      <FormHelperText>{definition.description}</FormHelperText>
      <Stack>
        {definition.validValues!.map((vv) => (
          <FormControlLabel
            key={`${definition.name}-label-${vv}`}
            control={
              <Switch
                key={`${definition.name}-switch-${vv}`}
                size="small"
                onChange={(e, c) => handleToggleValue(e, c, vv)}
                checked={value?.includes(vv)}
              />
            }
            label={vv}
          />
        ))}
      </Stack>
    </>
  );
};

export const YamlParamInput: React.FC<ParamInputProps<string>> = ({
  classes,
  definition,
  value,
  onValueChange,
}) => {
  const [isFocused, setFocused] = useState(false);

  const shrinkLabel = isFocused || !isEmpty(value);

  function handleValueChange(e: ChangeEvent<HTMLTextAreaElement>) {
    isFunction(onValueChange) && onValueChange(e.target.value);
  }

  return (
    <FormControl fullWidth classes={classes} required={definition.required}>
      <InputLabel
        shrink={shrinkLabel}
        htmlFor={definition.name}
        style={{
          backgroundColor: "#fff",
          color: shrinkLabel ? "#4abaeb" : undefined,
          padding: shrinkLabel ? "0 10px 0 5px" : undefined,
        }}
      >
        {definition.label}
      </InputLabel>
      <YamlEditor
        required={definition.required}
        name={definition.name}
        value={value ?? ""}
        onValueChange={handleValueChange}
        onFocus={() => setFocused(true)}
        onBlur={() => setFocused(false)}
        minHeight={200}
      />
      <FormHelperText>{definition.description}</FormHelperText>
    </FormControl>
  );
};

export const MapParamInput: React.FC<ParamInputProps<Record<string, string>>> =
  ({ classes, definition, value, onValueChange }) => {
    const initValue = valueToTupleArray(value);
    const [controlValue, setControlValue] = useState<Tuple[]>(initValue);

    const onChangeInput = useMemo(() => {
      return function (
        e: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>,
        row: number,
        index: number
      ) {
        setControlValue((prev) => {
          const newVal = [...prev];
          newVal[row][index] = e.target.value;
          return newVal;
        });
      };
    }, [setControlValue]);

    function handleBlur() {
      const mapValue = tupleArrayToMap(controlValue);
      onValueChange && onValueChange(mapValue);
    }

    return (
      <>
        <div>
          <label aria-required={definition.required} htmlFor={definition.name}>
            {definition.label}
            {definition.required && " *"}
          </label>

          <FormHelperText>{definition.description}</FormHelperText>

          <Grid container spacing={1}>
            <Grid item xs={6}>
              <Typography fontWeight={600}>Key</Typography>
            </Grid>
            <Grid item xs={6}>
              <Typography fontWeight={600}>Value</Typography>
            </Grid>
          </Grid>
          <Grid container spacing={1}>
            {controlValue.map(([k, v], rowIndex) => {
              if (rowIndex === controlValue.length - 1) {
                return null;
              }
              return (
                <>
                  <Grid key={`${definition.name}-${rowIndex}-0`} item xs={6}>
                    <OutlinedInput
                      key={`${definition.name}-${rowIndex}-0-input`}
                      size="small"
                      type="text"
                      value={k}
                      onChange={(e) => onChangeInput(e, rowIndex, 0)}
                      onBlur={handleBlur}
                    />
                  </Grid>
                  <Grid key={`${definition.name}-${rowIndex}-1`} item xs={6}>
                    <OutlinedInput
                      key={`${definition.name}-${rowIndex}-1-input`}
                      size="small"
                      type="text"
                      value={v}
                      onChange={(e) => onChangeInput(e, rowIndex, 1)}
                      onBlur={handleBlur}
                    />
                  </Grid>
                </>
              );
            })}
            <Grid item xs={6}>
              <OutlinedInput
                size="small"
                type="text"
                value={controlValue[controlValue.length - 1][0]}
                onChange={(e) => onChangeInput(e, controlValue.length - 1, 0)}
                onBlur={handleBlur}
              />
            </Grid>
            <Grid item xs={6}>
              <OutlinedInput
                size="small"
                type="text"
                value={controlValue[controlValue.length - 1][1]}
                onChange={(e) => onChangeInput(e, controlValue.length - 1, 1)}
                onBlur={handleBlur}
              />
            </Grid>
          </Grid>

          <Button
            startIcon={<PlusCircleIcon />}
            onClick={() => setControlValue((prev) => addRow(prev))}
          >
            Add
          </Button>
        </div>
      </>
    );
  };

export const StringsInput: React.FC<ParamInputProps<string[]>> = ({
  classes,
  definition,
  value,
  onValueChange,
}) => {
  const [inputValue, setInputValue] = useState("");
  const { setError, touched, errors, touch } = useValidationContext();

  // handleChipClick edits the selected chips value.
  function handleChipClick(ix: number) {
    if (!isArray(value)) {
      return;
    }

    // Edit the chips value
    setInputValue(value[ix]);

    // Remove the chip from the values because its being edited.
    const copy = [...value];
    copy.splice(ix, 1);
    isFunction(onValueChange) && onValueChange(copy);
  }

  // Make sure we "enter" the value if a user leaves the
  // input without hitting enter
  function handleBlur() {
    touch(definition.name);
    if (!isEmpty(inputValue)) {
      handleValueChange([...(value ?? []), inputValue]);
    }
  }

  function handleValueChange(newValue: string[]) {
    onValueChange && onValueChange(newValue);

    setError(
      definition.name,
      validateStringsField(newValue, definition.required)
    );
  }

  const label = definition.required
    ? `${definition.label} *`
    : `${definition.label}`;

  return (
    <FormControl fullWidth>
      <Autocomplete
        options={[]}
        multiple
        disableClearable
        freeSolo
        classes={classes}
        // value and onChange pertain to the string[] value of the input
        value={value ?? []}
        onChange={(e, v: string[]) => handleValueChange(v)}
        // inputValue and onInputChange refer to the latest string value being entered
        inputValue={inputValue}
        onInputChange={(e, newValue) => setInputValue(newValue)}
        onBlur={handleBlur}
        renderTags={(value: readonly string[], getTagProps) =>
          value.map((option: string, index: number) => (
            <Chip
              size="small"
              variant="outlined"
              label={option}
              {...getTagProps({ index })}
              classes={{ label: styles.chip }}
              onClick={() => handleChipClick(index)}
            />
          ))
        }
        renderInput={(params) => (
          <TextField
            {...params}
            label={label}
            size={"small"}
            helperText={definition.description}
            id={definition.name}
          />
        )}
      />
      {touched[definition.name] && errors[definition.name] && (
        <FormHelperText error={true}>{errors[definition.name]}</FormHelperText>
      )}
    </FormControl>
  );
};

export const BoolParamInput: React.FC<ParamInputProps<boolean>> = ({
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

export const IntParamInput: React.FC<ParamInputProps<number>> = ({
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

interface ResourceNameInputProps
  extends Omit<ParamInputProps<string>, "definition"> {
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

// Utility functions
export type Tuple = [string, string];

export function valueToTupleArray(value: any): Tuple[] {
  try {
    const tuples = Object.entries(value);

    tuples.push(["", ""]);
    return tuples as Tuple[];
  } catch (err) {
    return [["", ""]];
  }
}

export function tupleArrayToMap(tuples: Tuple[]): Record<string, string> {
  const mapValue: Record<string, string> = {};
  for (const [k, v] of tuples) {
    if (k === "") {
      continue;
    }

    mapValue[k] = v;
  }

  return mapValue;
}

function addRow(tuples: Tuple[]): Tuple[] {
  const newTuples = [...tuples];
  newTuples.push(["", ""]);
  return newTuples;
}
