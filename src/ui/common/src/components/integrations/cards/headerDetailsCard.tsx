import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { Integration } from '../../../utils/integrations';
import ExecutionStatus from '../../../utils/shared';
import { StatusIndicator } from '../../workflows/workflowStatus';
import IntegrationLogo from '../logo';
import {theme} from "../../../styles/theme/theme";

type ResourceHeaderDetailsCardProps = {
  integration: Integration;

  // Eg: "Used by 2 workflows"
  numWorkflowsUsingMsg: string;
};

export const ResourceHeaderDetailsCard: React.FC<
  ResourceHeaderDetailsCardProps
> = ({ integration, numWorkflowsUsingMsg }) => {
  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <Box display="flex" flexDirection="row" alignItems="center">
        <IntegrationLogo
          service={integration.service}
          size="medium"
          activated
        />

        <Box display="flex" flexDirection="column" sx={{ ml: 2, mr: 2 }}>
          <Box display="flex" flexDirection="row" alignItems={'center'}>
            <Typography sx={{ fontWeight: 400, mr: 2 }} variant="h5">
              {integration.name}
            </Typography>

            <StatusIndicator
              status={
                integration.exec_state?.status || ExecutionStatus.Succeeded
              }
              size="20px"
            />
          </Box>

          <Typography variant="caption" sx={{ fontWeight: 300 }}>
            {new Date(integration.createdAt * 1000).toLocaleString()}
          </Typography>

          <Typography variant="body2" sx={{ fontWeight: 300 }}>
            {numWorkflowsUsingMsg}
          </Typography>
        </Box>
      </Box>

      { integration.exec_state?.status === ExecutionStatus.Failed && (
              <Box
                  sx={{
                    backgroundColor: theme.palette.red[100],
                    color: theme.palette.red[600],
                    p: 2,
                    paddingBottom: '16px',
                    paddingTop: '16px',
                    height: 'fit-content',
                  }}
              >
                <pre style={{ margin: '0px' }}>
                  {`${integration.exec_state?.error.tip}\n\n${integration.exec_state?.error.context}`}
                </pre>
              </Box>
          )
      }

    </Box>
  );
};
