import {
  Typography,
  Table,
  TableBody,
  TableRow,
  TableCell,
  Card,
  CardHeader,
  Button,
  IconButton,
  CardContent,
  TextField,
  Stack,
} from "@mui/material";
import { memo, useRef, useState } from "react";
import { CardContainer } from "../../../components/CardContainer";
import { EditIcon } from "../../../components/Icons";
import { applyResources } from "../../../utils/rest/apply-resources";
import { ShowPageConfig } from ".";
import { cloneIntoConfig } from "./utils";
import { ConfirmDeleteResourceDialog } from "../../../components/ConfirmDeleteResourceDialog";
import { deleteResources } from "../../../utils/rest/delete-resources";
import { useSnackbar } from "notistack";
import { ResourceKind, UpdateStatus } from "../../../types/resources";
import { useNavigate } from "react-router-dom";

import styles from "./configuration-page.module.scss";
import mixins from "../../../styles/mixins.module.scss";
import { DuplicateConfigDialog } from "./DuplicateConfigDialog";

const DetailsSectionComponent: React.FC<{
  configuration: NonNullable<ShowPageConfig>;
  refetch: () => void;
  onSaveDescriptionSuccess: () => void;
  onSaveDescriptionError: () => void;
}> = ({
  configuration,
  refetch,
  onSaveDescriptionError,
  onSaveDescriptionSuccess,
}) => {
  const [duplicateDialogOpen, setDuplicateDialogOpen] = useState(false);
  const [openDeleteConfirm, setOpenDelete] = useState(false);
  const [editingDescription, setEditingDescription] = useState(false);
  const descriptionInputRef = useRef<HTMLTextAreaElement | null>(null);

  const snackbar = useSnackbar();
  const navigate = useNavigate();

  async function deleteConfiguration() {
    try {
      const configToDelete = {
        metadata: {
          name: configuration.metadata.name,
        },
        kind: ResourceKind.CONFIGURATION,
      };
      const { updates } = await deleteResources([configToDelete]);

      // Verify we get status deleted on this configuration.

      const update = updates.find(
        (u) => u.resource.metadata.name === configuration.metadata.name
      );

      if (update == null || update.status !== UpdateStatus.DELETED) {
        snackbar.enqueueSnackbar("Failed to delete configuration.", {
          variant: "error",
          autoHideDuration: 5000,
        });
        return;
      }

      navigate("/configurations");
    } catch (err) {
      snackbar.enqueueSnackbar("Failed to delete configuration.", {
        variant: "error",
        autoHideDuration: 5000,
      });
      console.error(err);
    }
  }

  async function saveDescription() {
    if (descriptionInputRef.current == null) {
      return;
    }

    try {
      const newConfig = cloneIntoConfig(configuration);
      newConfig.metadata.description = descriptionInputRef.current.value;
      await applyResources([newConfig]);

      onSaveDescriptionSuccess();
      setEditingDescription(false);
      refetch();
    } catch (err) {
      onSaveDescriptionError();
      console.error(err);
    }
  }

  return (
    <CardContainer>
      <Stack direction="row" justifyContent="space-between" alignItems="center">
        <Typography variant="h5" marginBottom="1rem">
          Details
        </Typography>
        <Stack direction="row" spacing={2}>
          <Button
            variant="outlined"
            onClick={() => setDuplicateDialogOpen(true)}
          >
            Duplicate
          </Button>

          <Button
            color="error"
            variant="contained"
            onClick={() => setOpenDelete(true)}
          >
            Delete
          </Button>
        </Stack>
      </Stack>

      {/* Agent Details Table */}
      <div className={styles["details-box"]}>
        <Table>
          <TableBody>
            <TableRow>
              <TableCell className={styles["row-width"]}>
                <Typography variant="overline">Name</Typography>
              </TableCell>
              <TableCell>{configuration.metadata.name}</TableCell>
            </TableRow>
            <TableRow>
              <TableCell>
                <Typography variant="overline">Platform</Typography>
              </TableCell>
              <TableCell>{configuration.metadata.labels.platform}</TableCell>
            </TableRow>
          </TableBody>
        </Table>

        {/* Description Box */}
        <Card className={styles.description} variant="outlined">
          <CardHeader
            titleTypographyProps={{
              variant: "overline",
            }}
            title="Description"
            action={
              editingDescription ? (
                <>
                  <Button
                    size="small"
                    color="inherit"
                    onClick={() => setEditingDescription(false)}
                    classes={{ root: mixins["mr-2"] }}
                  >
                    Cancel
                  </Button>
                  <Button
                    size="small"
                    color="primary"
                    variant="outlined"
                    onClick={saveDescription}
                  >
                    Save
                  </Button>
                </>
              ) : (
                <IconButton
                  size="small"
                  onClick={() => setEditingDescription(true)}
                >
                  <EditIcon />
                </IconButton>
              )
            }
          />

          <CardContent>
            <Typography variant="subtitle1"></Typography>

            {editingDescription ? (
              <TextField
                multiline
                inputRef={descriptionInputRef}
                defaultValue={configuration.metadata.description}
                fullWidth
              />
            ) : (
              <Typography variant="body1" whiteSpace="pre-wrap">
                {configuration.metadata.description}
              </Typography>
            )}
          </CardContent>
        </Card>
      </div>
      <ConfirmDeleteResourceDialog
        open={openDeleteConfirm}
        onCancel={() => setOpenDelete(false)}
        onDelete={deleteConfiguration}
        action="delete"
      >
        <Typography>
          Are you sure you want to delete this configuration?
        </Typography>
      </ConfirmDeleteResourceDialog>

      <DuplicateConfigDialog
        currentConfigName={configuration.metadata.name}
        open={duplicateDialogOpen}
        onClose={() => setDuplicateDialogOpen(false)}
        maxWidth="xs"
      />
    </CardContainer>
  );
};

export const DetailsSection = memo(DetailsSectionComponent);
