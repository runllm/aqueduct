import { Box, Link, Typography } from '@mui/material';
import React, { useState } from 'react';

import { theme } from '../../styles/theme/theme';
import { getPathPrefix } from '../../utils/getPathPrefix';
import {
  Integration,
  NotificationIntegrationConfig,
} from '../../utils/integrations';
import { NotificationLogLevel } from '../../utils/notifications';
import { Button } from '../primitives/Button.styles';
import NotificationLevelSelector from './NotificationLevelSelector';

export type NotificationConfigsMap = {
  [integrationId: string]: NotificationIntegrationConfig;
};

type Props = {
  notifications: Integration[];
  onSave: (updatedConfigs: NotificationConfigsMap) => void;
  isSaving: boolean;
};

function ConfigDifference(
  initialConfigs: NotificationConfigsMap,
  currentConfigs: NotificationConfigsMap
): NotificationConfigsMap {
  const results = {};
  Object.entries(currentConfigs).forEach(([k, v]) => {
    const initialV = initialConfigs[k];
    if (initialV.enabled !== v.enabled || initialV.level !== v.level) {
      results[k] = v;
    }
  });

  return results;
}

const AccountNotificationSettingsSelector: React.FC<Props> = ({
  onSave,
  notifications,
  isSaving,
}) => {
  const initialConfigs = Object.fromEntries(
    notifications.map((x) => [x.id, x.config as NotificationIntegrationConfig])
  );
  const [configs, setConfigs] =
    useState<NotificationConfigsMap>(initialConfigs);

  if (!notifications.length) {
    return (
      <Typography variant="body1">
        You do not have any notification configured. You can add new
        notifications from the{' '}
        <Link href={`${getPathPrefix()}/integrations`} target="_blank">
          integrations
        </Link>{' '}
        page.
      </Typography>
    );
  }

  const configDifference = ConfigDifference(initialConfigs, configs);
  const showSaveAndCancel = Object.keys(configDifference).length > 0;
  const onCancel = () => setConfigs(initialConfigs);

  return (
    <Box display="flex" flexDirection="column" alignItems="left">
      {notifications.map((n) => (
        <Box
          key={n.id}
          display="flex"
          flexDirection="column"
          alignItems="left"
          padding={1}
          marginBottom={1}
          bgcolor={theme.palette.gray[100]}
        >
          <Typography variant="body1">{n.name}</Typography>
          <NotificationLevelSelector
            level={configs[n.id].level as NotificationLogLevel}
            onSelectLevel={(level) =>
              setConfigs({
                ...configs,
                [n.id]: { enabled: configs[n.id].enabled, level: level },
              })
            }
            disabled={configs[n.id].enabled === 'false'}
            disableSelectorMessage="Do not apply this notification to all workflows."
            onDisable={(disabled) =>
              setConfigs({
                ...configs,
                [n.id]: {
                  enabled: disabled ? 'false' : 'true',
                  level: configs[n.id].level,
                },
              })
            }
          />
        </Box>
      ))}
      {showSaveAndCancel && (
        <Box display="flex" flexDirection="row">
          <Button
            onClick={() => onSave(configDifference)}
            color="primary"
            disabled={isSaving}
          >
            Save
          </Button>
          <Button
            onClick={() => onCancel()}
            sx={{ marginLeft: 2 }}
            color="secondary"
          >
            Cancel
          </Button>
        </Box>
      )}
    </Box>
  );
};

export default AccountNotificationSettingsSelector;
