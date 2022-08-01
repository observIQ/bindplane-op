import { Resource, ResourceKind, ResourceStatus } from "./resources";

export interface ApplyPayload {
  resources: Resource[];
}

export interface ApplyResponse {
  updates: ResourceStatus[];
}

export type DeletePayload = {
  // Technically only the name field would be required.
  resources: {
    metadata: { name: string };
    kind: ResourceKind;
  }[];
};
export type DeleteResponse = ApplyResponse;

export interface ErrorResponse {
  errors: string[];
}

export interface InstallCommandResponse {
  command: string;
}

export interface PatchLabelsPayload {
  labels: { [key: string]: string };
}

export interface PatchLabelsResponse {
  labels: { [key: string]: string };
  errors: string[];
}

export interface LabelAgentsPayload {
  ids: string[];
  labels: { [key: string]: string };
  overwrite?: boolean;
}

export interface LabelAgentsResponse {
  errors: string[];
}

export interface CopyConfigPayload {
  name: string;
}

export type CopyConfigResponse = CopyConfigPayload;
