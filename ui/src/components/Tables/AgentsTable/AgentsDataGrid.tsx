import { Stack } from "@mui/material";
import {
  DataGrid,
  GridCellParams,
  GridColumns,
  GridDensityTypes,
  GridRowParams,
  GridSelectionModel,
  GridValueGetterParams,
} from "@mui/x-data-grid";
import React, { memo } from "react";
import { renderAgentLabels, renderAgentStatus } from "../utils";
import {
  Agent,
  AgentChangesSubscription,
} from "../../../graphql/generated";
import { Link } from "react-router-dom";
import { AgentStatus } from "../../../types/agents";
import { isFunction } from "lodash";

export enum AgentsTableField {
  NAME = "name",
  STATUS = "status",
  VERSION = "version",
  CONFIGURATION = "configuration",
  OPERATING_SYSTEM = "operatingSystem",
  LABELS = "labels",
}

export type AgentsTableChange =  AgentChangesSubscription["agentChanges"]["agentChanges"][0];
export type AgentsTableRow =
  AgentChangesSubscription["agentChanges"]["agentChanges"][0]["agent"];

interface AgentsDataGridProps {
  onAgentsSelected?: (agentIds: GridSelectionModel) => void;
  isRowSelectable?: (params: GridRowParams<AgentsTableRow>) => boolean;
  density?: GridDensityTypes;
  loading: boolean;
  minHeight?: string;
  agents?: AgentsTableRow[];
  columnFields?: AgentsTableField[];
}

const AgentsDataGridComponent: React.FC<AgentsDataGridProps> = ({
  onAgentsSelected,
  isRowSelectable,
  minHeight,
  loading,
  agents,
  columnFields,
  density,
}) => {
  const columns: GridColumns = (columnFields || []).map((field) => {
    switch (field) {
      case AgentsTableField.STATUS:
        return {
          field: AgentsTableField.STATUS,
          headerName: "Status",
          width: 150,
          renderCell: renderStatusDataCell,
        };
      case AgentsTableField.VERSION:
        return {
          field: AgentsTableField.VERSION,
          headerName: "Version",
          width: 100,
        };
      case AgentsTableField.CONFIGURATION:
        return {
          field: AgentsTableField.CONFIGURATION,
          headerName: "Configuration",
          width: 200,
          renderCell: renderConfigurationCell,
          valueGetter: (params: GridValueGetterParams<Agent>) => {
            const configuration = params.row.configurationResource;
            return configuration?.metadata?.name;
          },
        };
      case AgentsTableField.OPERATING_SYSTEM:
        return {
          field: AgentsTableField.OPERATING_SYSTEM,
          headerName: "Operating System",
          width: 200,
        };
      case AgentsTableField.LABELS:
        return {
          sortable: false,
          field: AgentsTableField.LABELS,
          headerName: "Labels",
          width: 300,
          renderCell: renderLabelDataCell,
          valueGetter: (params: GridValueGetterParams<Agent>) => {
            return params.row.labels;
          },
        };
      default:
        return {
          field: AgentsTableField.NAME,
          headerName: "Name",
          valueGetter: (params: GridValueGetterParams<Agent>) => {
            return params.row.name;
          },
          renderCell: renderNameDataCell,
          width: 325,
        };
    }
  });

  function handleSelect(s: GridSelectionModel) {
    if (!isFunction(onAgentsSelected)) {
      return;
    }

    onAgentsSelected(s);
  }

  return (
    <DataGrid
      checkboxSelection={isFunction(onAgentsSelected)}
      isRowSelectable={isRowSelectable}
      onSelectionModelChange={handleSelect}
      density={density}
      components={{
        NoRowsOverlay: () => (
          <Stack height="100%" alignItems="center" justifyContent="center">
            No Agents
          </Stack>
        ),
      }}
      style={{ minHeight }}
      loading={loading}
      disableSelectionOnClick
      columns={columns}
      rows={agents ?? []}
    />
  );
};

function renderConfigurationCell(cellParams: GridCellParams<string>) {
  const configName = cellParams.value;
  if (configName == null) {
    return <></>;
  }
  return <Link to={`/configurations/${configName}`}>{configName}</Link>;
}

function renderNameDataCell(
  cellParams: GridCellParams<{ name: string; id: string }, Agent>
): JSX.Element {
  return <Link to={`/agents/${cellParams.row.id}`}>{cellParams.row.name}</Link>;
}

function renderLabelDataCell(
  cellParams: GridCellParams<Record<string, string>>
): JSX.Element {
  return renderAgentLabels(cellParams.value);
}

function renderStatusDataCell(
  cellParams: GridCellParams<AgentStatus>
): JSX.Element {
  return renderAgentStatus(cellParams.value);
}

AgentsDataGridComponent.defaultProps = {
  minHeight: "calc(100vh - 300px)",
  columnFields: [
    AgentsTableField.NAME,
    AgentsTableField.STATUS,
    AgentsTableField.VERSION,
    AgentsTableField.CONFIGURATION,
    AgentsTableField.OPERATING_SYSTEM,
    AgentsTableField.LABELS,
  ],
};

export const AgentsDataGrid = memo(AgentsDataGridComponent);
