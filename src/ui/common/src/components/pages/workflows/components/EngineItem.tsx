import { Box, Typography } from '@mui/material';
import React from 'react';

export interface EngineItemProps {
  engineName: string;
  engineIconUrl: string;
}

export const EngineItem: React.FC<EngineItemProps> = ({
  engineName,
  engineIconUrl,
}) => {
  return (
    <Box display="flex" alignItems="left" justifyContent="left">
      <img
        src={engineIconUrl}
        style={{ marginTop: '4px', marginRight: '8px' }}
        width="16px"
        height="16px"
      />
      <Typography variant="body1">{engineName}</Typography>
    </Box>
  );
};

export default EngineItem;
