import * as rtkQueryRaw from '@reduxjs/toolkit/dist/query/react/index.js';
import { FetchBaseQueryError } from '@reduxjs/toolkit/query/react';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import { dagGetQuery, DagGetRequest, DagGetResponse } from './DagGet';
import {
  dagResultGetQuery,
  DagResultGetRequest,
  DagResultGetResponse,
} from './DagResultGet';
import {
  storageMigrationListQuery,
  storageMigrationListRequest,
  storageMigrationListResponse,
} from './ListStorageMigrations';
import {
  workflowGetQuery,
  WorkflowGetRequest,
  WorkflowGetResponse,
} from './WorkflowGet';
import { versionNumberGetQuery, VersionNumberGetRequest, VersionNumberGetResponse } from './VersionNumberGet';

const { createApi, fetchBaseQuery } = ((rtkQueryRaw as any).default ??
  rtkQueryRaw) as typeof rtkQueryRaw;

const transformErrorResponse = (resp: FetchBaseQueryError) =>
  (resp.data as { error: string })?.error;

export const aqueductApi = createApi({
  reducerPath: 'aqueductApi',
  baseQuery: fetchBaseQuery({ baseUrl: `${apiAddress}/api/v2/` }),
  keepUnusedDataFor: 60,
  endpoints: (builder) => ({
    dagGet: builder.query<DagGetResponse, DagGetRequest>({
      query: (req) => dagGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    dagResultGet: builder.query<DagResultGetResponse, DagResultGetRequest>({
      query: (req) => dagResultGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    storageMigrationList: builder.query<
      storageMigrationListResponse,
      storageMigrationListRequest
    >({
      query: (req) => storageMigrationListQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    workflowGet: builder.query<WorkflowGetResponse, WorkflowGetRequest>({
      query: (req) => workflowGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    versionNumberGet: builder.query<VersionNumberGetResponse, VersionNumberGetRequest>({
      query: (req) => versionNumberGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
  }),
});

export const {
  useDagGetQuery,
  useDagResultGetQuery,
  useStorageMigrationListQuery,
  useWorkflowGetQuery,
  useVersionNumberGetQuery
} = aqueductApi;
