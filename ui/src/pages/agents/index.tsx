import { Button, Typography } from "@mui/material";
import { GridRowParams, GridSelectionModel } from "@mui/x-data-grid";
import React, { useState } from "react";
import { Link } from "react-router-dom";
import { CardContainer } from "../../components/CardContainer";
import { PlusCircleIcon } from "../../components/Icons";
import { AgentsTable } from "../../components/Tables/AgentsTable";
import { classes } from "../../utils/styles";
import { deleteAgents } from "../../utils/rest/delete-agents";
import { useSnackbar } from "notistack";
import { Agent } from "../../graphql/generated";
import { AgentStatus } from "../../types/agents";
import { ConfirmDeleteResourceDialog } from "../../components/ConfirmDeleteResourceDialog";
import { withRequireLogin } from "../../contexts/RequireLogin";
import { withNavBar } from "../../components/NavBar";

import mixins from "../../styles/mixins.module.scss";

export const AgentsPageContent: React.FC = () => {
  const [selectedAgents, setSelectedAgents] = useState<GridSelectionModel>([]);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

  const { enqueueSnackbar } = useSnackbar();

  function handleSelect(g: GridSelectionModel) {
    setSelectedAgents(g);
  }

  async function handleDeleteAgents() {
    try {
      await deleteAgents(selectedAgents as string[]);
      setDeleteConfirmOpen(false);
    } catch (err) {
      console.log(err);
      enqueueSnackbar("Failed to delete agents.", { variant: "error" });
    }
  }

  function isRowSelectable(params: GridRowParams<Agent>): boolean {
    return params.row.status === AgentStatus.DISCONNECTED;
  }

  return (
    <>
      <ConfirmDeleteResourceDialog
        onDelete={handleDeleteAgents}
        onCancel={() => setDeleteConfirmOpen(false)}
        action={"delete"}
        open={deleteConfirmOpen}
        title={`Delete ${selectedAgents.length} Agent${
          selectedAgents.length > 1 ? "s" : ""
        }?`}
      >
        <>
          <Typography>
            Agents will reappear in BindPlane OP if reconnected.
          </Typography>
        </>
      </ConfirmDeleteResourceDialog>
      <CardContainer>
        <Button
          component={Link}
          variant={"contained"}
          classes={{ root: mixins["float-right"] }}
          to="/agents/install"
          startIcon={<PlusCircleIcon />}
        >
          Install Agents
        </Button>

        {selectedAgents.length > 0 && (
          <Button
            variant="contained"
            color="error"
            classes={{ root: classes([mixins["float-right"], mixins["mr-3"]]) }}
            onClick={() => setDeleteConfirmOpen(true)}
          >
            Delete {selectedAgents.length} Agent
            {selectedAgents.length > 1 && "s"}
          </Button>
        )}

        <Typography variant="h5" className={mixins["mb-5"]}>
          Agents
        </Typography>

        <AgentsTable
          onAgentsSelected={handleSelect}
          isRowSelectable={isRowSelectable}
        />
      </CardContainer>
    </>
  );
};

export const AgentsPage = withRequireLogin(withNavBar(AgentsPageContent));
