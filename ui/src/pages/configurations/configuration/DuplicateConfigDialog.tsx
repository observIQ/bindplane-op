import {
  Button,
  Dialog,
  DialogContent,
  DialogProps,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { isFunction } from "lodash";
import { useSnackbar } from "notistack";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useGetConfigNamesQuery } from "../../../graphql/generated";
import { validateNameField } from "../../../utils/forms/validate-name-field";
import { copyConfig } from "../../../utils/rest/copy-config";

interface Props extends DialogProps {
  currentConfigName: string;
}

export const DuplicateConfigDialog: React.FC<Props> = ({
  currentConfigName,
  ...dialogProps
}) => {
  const [newName, setNewName] = useState("");
  const [touched, setTouched] = useState(false);

  const { data, error } = useGetConfigNamesQuery();
  const configNames = data?.configurations.configurations.map(
    (c) => c.metadata.name
  );
  const formError = validateNameField(newName, "configuration", configNames);

  const { enqueueSnackbar } = useSnackbar();
  const navigate = useNavigate();

  useEffect(() => {
    if (!dialogProps.open) {
      setTouched(false);
      setNewName("");
    }
  }, [dialogProps.open]);

  useEffect(() => {
    if (error != null) {
      const message = "Error retrieving configuration names.";
      enqueueSnackbar(message, { key: message, variant: "error" });
    }
  }, [enqueueSnackbar, error]);

  async function handleSave(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    const status = await copyConfig({
      existingName: currentConfigName,
      newName: newName,
    });

    let message: string;
    switch (status) {
      case "conflict":
        message = "Looks like a configuration with that name already exists.";
        enqueueSnackbar(message, { key: message, variant: "warning" });
        break;
      case "error":
        message = "Oops, something went wrong. Failed to duplicate.";
        enqueueSnackbar(message, { key: message, variant: "error" });
        break;
      case "created":
        message = "Successfully duplicated!";
        enqueueSnackbar(message, { key: message, variant: "success" });
        navigate(`/configurations/${newName}`);
        break;
    }
  }

  function handleChange(
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) {
    if (!touched) {
      setTouched(true);
    }
    setNewName(e.target.value);
  }

  return (
    <Dialog {...dialogProps}>
      <DialogContent>
        <Typography variant="h6" marginBottom={2}>
          Duplicate Configuration
        </Typography>
        <Typography variant="body2">
          Clicking save will create a new Configuration with identical sources
          and destinations.
        </Typography>
        <form onSubmit={handleSave}>
          <TextField
            value={newName}
            onChange={handleChange}
            size="small"
            label="Name"
            helperText={touched && formError ? formError : undefined}
            name="name"
            fullWidth
            error={touched && formError != null}
            margin="normal"
            required
            onBlur={() => setTouched(true)}
          />

          <Stack direction="row" justifyContent="space-between">
            <Button
              color="secondary"
              onClick={() => {
                isFunction(dialogProps.onClose) &&
                  dialogProps.onClose({}, "backdropClick");
              }}
            >
              Cancel
            </Button>
            <Button
              variant="contained"
              disabled={formError != null}
              type="submit"
            >
              Save
            </Button>
          </Stack>
        </form>
      </DialogContent>
    </Dialog>
  );
};
