// Before we fully migrate workflow result page, this store is
// only used for metrics and check details page.
import { createSlice } from '@reduxjs/toolkit';

import { handleGetArtifactResultContent } from '../handlers/getArtifactResultContent';
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';

export interface ArtifactResultContentState {
  contents: {
    [artifactResultId: string]: {
      status: LoadingStatus;
      data?: string;
    };
  };
}

const initialState: ArtifactResultContentState = { contents: {} };

export const artifactResultContentsSlice = createSlice({
  name: 'artifactResultContentsReducer',
  initialState: initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(
      handleGetArtifactResultContent.pending,
      (state, { meta }) => {
        console.log('pending');
        const id = meta.arg.artifactResultId;
        state.contents[id] = {
          status: { loading: LoadingStatusEnum.Loading, err: '' },
        };
      }
    );
    builder.addCase(
      handleGetArtifactResultContent.fulfilled,
      (state, { meta, payload }) => {
        console.log('fulfilled');
        const id = meta.arg.artifactResultId;
        state.contents[id] = {
          data: payload,
          status: { loading: LoadingStatusEnum.Succeeded, err: '' },
        };
      }
    );
    builder.addCase(
      handleGetArtifactResultContent.rejected,
      (state, { meta, payload }) => {
        console.log('rejected');
        const id = meta.arg.artifactResultId;

        state.contents[id] = {
          status: { loading: LoadingStatusEnum.Failed, err: payload as string },
        };
      }
    );
  },
});

export default artifactResultContentsSlice.reducer;
