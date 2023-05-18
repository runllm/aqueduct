// This file should map exactly to
// src/golang/cmd/server/handler/v2/workflow_objects_get.go

import { SavedObject } from 'src/utils/workflows';

import { APIKeyParameter } from '../parameters/Header';
import { WorkflowIdParameter } from '../parameters/Path';

export type WorkflowObjectsGetRequest = APIKeyParameter & WorkflowIdParameter;

export type WorkflowObjectsGetResponse = { [id: string]: SavedObject };

export const workflowObjectsGetQuery = (req: WorkflowObjectsGetRequest) => ({
  url: `workflow/${req.workflowId}/objects`,
  headers: { 'api-key': req.apiKey },
});
