// This file should map exactly to
// `src/golang/lib/response/node.go`
import { ArtifactType, SerializationType } from '../../utils/artifacts';
import { OperatorSpec } from '../../utils/operators';
import { ExecState } from '../../utils/shared';

export type OperatorWithArtifactNodeResponse = {
  id: string;
  dag_id: string;
  artifact_id?: string;
  name: string;
  description: string;
  spec?: OperatorSpec;
  type?: ArtifactType;
  inputs: string[];
  outputs: string[];
};

export type OperatorWithArtifactNodeResultResponse = {
  id: string;
  operator_exec_state?: ExecState;

  artifact_id?: string;
  serialization_type?: SerializationType;
  content_path?: string;
  content_serialized?: string;
  artifact_exec_state?: ExecState;
};

export type ArtifactResponse = {
  id: string;
  dag_id: string;
  name: string;
  description: string;
  type: ArtifactType;
  input: string;
  outputs: string[];
};

export type ArtifactResultResponse = {
  id: string;
  artifact_id: string;
  serialization_type: SerializationType;
  content_path: string;
  content_serialized: string;
  exec_state?: ExecState;
};

export type OperatorResponse = {
  id: string;
  dag_id: string;
  name: string;
  description: string;
  spec?: OperatorSpec;
  inputs: string[];
  outputs: string[];
};

export type OperatorResultResponse = {
  id: string;
  operator_id: string;
  exec_state?: ExecState;
};

export type NodesResponse = {
  operators: OperatorResponse[];
  artifacts: ArtifactResponse[];
  // TODO: ENG-2987 Create separate sections for Metrics/Checks
  // metrics: OperatorWithArtifactNodeResponse[];
  // checks: OperatorWithArtifactNodeResponse[];
};

export type NodeResultsResponse = {
  operators: OperatorResultResponse[];
  artifacts: ArtifactResultResponse[];
  // TODO: ENG-2987 Create separate sections for Metrics/Checks
  // metrics: OperatorWithArtifactNodeResultResponse[];
  // checks: OperatorWithArtifactNodeResultResponse[];
};

export type NodeContentResponse = {
  name: string;
  data: string;
};
