import { faTags } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { handleGetServerConfig } from '../../../handlers/getServerConfig';
import { RootState } from '../../../stores/store';

import {
  Integration,
  S3Config,
  SupportedIntegrations,
} from '../../../utils/integrations';
import { LoadingStatus } from '../../../utils/shared';
import useUser from '../../hooks/useUser';
import { AqueductDemoCard } from './aqueductDemoCard';
import { BigQueryCard } from './bigqueryCard';
import { CondaCard } from './condaCard';
import { EmailCard } from './emailCard';
import { KubernetesCard } from './kubernetesCard';
import { LambdaDetailCard } from './lambdaCard';
import { MariaDbCard } from './mariadbCard';
import { MySqlCard } from './mysqlCard';
import { PostgresCard } from './postgresCard';
import { RedshiftCard } from './redshiftCard';
import { S3Card } from './s3Card';
import { SlackCard } from './slackCard';
import { SnowflakeCard } from './snowflakeCard';

type DetailIntegrationCardProps = {
  integration: Integration;
  connectStatus?: LoadingStatus;
};

export const DetailIntegrationCard: React.FC<DetailIntegrationCardProps> = ({
  integration,
}) => {

  const { user } = useUser();
  const dispatch = useDispatch();
  const serverConfig = useSelector(
    (state: RootState) => state.serverConfigReducer
  );

  const storageConfig = serverConfig?.config?.storageConfig;

  useEffect(() => {
    async function fetchServerConfig() {
      if (user) {
        await dispatch(handleGetServerConfig({ apiKey: user.apiKey }));
      }
    }

    fetchServerConfig();
  }, [user]);

  let serviceCard;
  switch (integration.service) {
    case 'Postgres':
      serviceCard = <PostgresCard integration={integration} />;
      break;
    case 'Snowflake':
      serviceCard = <SnowflakeCard integration={integration} />;
      break;
    case 'Aqueduct Demo':
      serviceCard = <AqueductDemoCard integration={integration} />;
      break;
    case 'MySQL':
      serviceCard = <MySqlCard integration={integration} />;
      break;
    case 'Redshift':
      serviceCard = <RedshiftCard integration={integration} />;
      break;
    case 'MariaDB':
      serviceCard = <MariaDbCard integration={integration} />;
      break;
    case 'BigQuery':
      serviceCard = <BigQueryCard integration={integration} />;
      break;
    case 'S3':
      serviceCard = <S3Card integration={integration} />;
      break;
    case 'Kubernetes':
      serviceCard = <KubernetesCard integration={integration} />;
      break;
    case 'Lambda':
      serviceCard = <LambdaDetailCard integration={integration} />;
      break;
    case 'Conda':
      serviceCard = <CondaCard integration={integration} />;
      break;
    case 'Email':
      serviceCard = <EmailCard integration={integration} />;
      break;
    case 'Slack':
      serviceCard = <SlackCard integration={integration} />;
      break;
    default:
      serviceCard = null;
  }

  let dataStorageInfo, dataStorageText = null;
  if (storageConfig && storageConfig.type === 'SQLite') {
    dataStorageInfo = (
      <Box component="span">
        <FontAwesomeIcon icon={faTags} />
      </Box>
    );

    dataStorageText = (
      <Typography variant={'body2'}>
        <strong>Storage Type:</strong> {dataStorageInfo}
      </Typography>
    );
  }

  let createdOnText = null;
  if (
    integration.service !== 'Kubernetes' &&
    integration.service !== 'Lambda'
  ) {
    createdOnText = (
      <Typography variant="body2">
        <strong>Connected On: </strong>
        {new Date(integration.createdAt * 1000).toLocaleString()}
      </Typography>
    );
  }
  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        width: '900px',
      }}
    >
      <Box sx={{ display: 'flex', flexDirection: 'row' }}>
        <img
          height="45px"
          src={SupportedIntegrations[integration.service].logo}
        />
        <Box sx={{ ml: 3 }}>
          <Box display="flex" flexDirection="row" marginBottom={1}>
            <Typography sx={{ fontFamily: 'Monospace' }} variant="h4">
              {integration.name}
            </Typography>
          </Box>
          <Box marginBottom={1}>{createdOnText}</Box>
          {dataStorageText}
          {serviceCard}
        </Box>
      </Box>
    </Box>
  );
};
