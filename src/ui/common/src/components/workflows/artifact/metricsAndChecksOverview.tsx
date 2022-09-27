import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { ArtifactResultResponse } from '../../../handlers/responses/artifact';
import { DataSchema } from '../../../utils/data';
import OperatorExecStateTable, {
  OperatorExecStateTableType,
} from '../../tables/OperatorExecStateTable';

// This file contains two major components to show metrics and checks
// associated with a given artifact.

const schema: DataSchema = {
  fields: [
    { name: 'Title', type: 'varchar' },
    { name: 'Value', type: 'varchar' },
  ],
  pandas_version: '',
};

type MetricsOverviewProps = {
  metrics: ArtifactResultResponse[];
};

export const MetricsOverview: React.FC<MetricsOverviewProps> = ({
  metrics,
}) => {
  const metricTableEntries = {
    schema: schema,
    data: metrics
      .map((metricArtf) => {
        let name = metricArtf.name;
        if (name.endsWith('artifact') || name.endsWith('Aritfact')) {
          name = name.slice(0, 0 - 'artifact'.length);
        }
        return {
          title: name,
          value: metricArtf.result?.content_serialized,
        };
      })
      .filter((x) => !!x.value),
  };

  return (
    <Box width="100%">
      <Typography variant="h5" component="div" marginBottom="8px">
        Metrics
      </Typography>
      {metricTableEntries.data.length > 0 ? (
        <OperatorExecStateTable
          schema={metricTableEntries.schema}
          rows={metricTableEntries}
          tableType={OperatorExecStateTableType.Metric}
        />
      ) : (
        <Typography variant="body2">
          This artifact has no associated downstream Metrics.
        </Typography>
      )}
    </Box>
  );
};

export type ChecksOverviewProps = {
  checks: ArtifactResultResponse[];
};

export const ChecksOverview: React.FC<ChecksOverviewProps> = ({ checks }) => {
  const checkTableEntries = {
    schema: schema,
    data: checks
      .map((checkArtf) => {
        let name = checkArtf.name;
        if (name.endsWith('artifact') || name.endsWith('Aritfact')) {
          name = name.slice(0, 0 - 'artifact'.length);
        }
        return {
          title: name,
          value: checkArtf.result?.content_serialized,
        };
      })
      .filter((x) => !!x.value),
  };

  return (
    <Box width="100%">
      <Typography variant="h5" component="div" marginBottom="8px">
        Checks
      </Typography>
      {checkTableEntries.data.length > 0 ? (
        <OperatorExecStateTable
          schema={checkTableEntries.schema}
          rows={checkTableEntries}
          tableType={OperatorExecStateTableType.Check}
        />
      ) : (
        <Typography variant="body2">
          This artifact has no associated downstream Checks.
        </Typography>
      )}
    </Box>
  );
};
