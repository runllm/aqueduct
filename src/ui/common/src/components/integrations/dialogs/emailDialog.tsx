import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, {useState} from 'react';

import { EmailConfig } from '../../../utils/integrations';
import { IntegrationTextInputField } from './IntegrationTextInputField';
import NotificationLevelSelector from '../../notifications/NotificationLevelSelector';
import { NotificationLogLevel } from 'src';

const Placeholders = {
  host: 'smtp.myprovider.com',
  port: '',
  user: 'mysender@myprovider.com',
  password: '******',
  reciever: 'myreciever@myprovider.com',
  level: 'succeeded',
};

type Props = {
  onUpdateField: (field: keyof EmailConfig, value: string) => void;
  value?: EmailConfig;
};

export const EmailDialog: React.FC<Props> = ({
  onUpdateField,
  value,
}) => {
  const [receiver, setReceiver] = useState('')
  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Host *"
        description="The hostname address of the email SMTP server."
        placeholder={Placeholders.host}
        onChange={(event) => onUpdateField('host', event.target.value)}
        value={value?.host ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Port *"
        description="The port number of the email SMTP server."
        placeholder={Placeholders.port}
        onChange={(event) => onUpdateField('port', event.target.value)}
        value={value?.port ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Sender *"
        description="The email address of the sender."
        placeholder={Placeholders.user}
        onChange={(event) => onUpdateField('user', event.target.value)}
        value={value?.user ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={false}
        label="Password"
        description="The password corresponding to the above username."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => {
          if (!!event.target.value) {
            onUpdateField('password', event.target.value);
          }
        }}
        value={value?.password ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Receiver *"
        description="The email address of the receiver."
        placeholder={Placeholders.reciever}
        onChange={(event) => {
          onUpdateField('targets_serialized', JSON.stringify([event.target.value]))
          setReceiver(event.target.value)
        }}
        value={receiver ?? null}
      />

      <Box sx={{ mt: 2 }}>
        <Box sx={{ my: 1 }}>
          <Typography variant="body1" sx={{ fontWeight: 'bold' }}>
            Level *
          </Typography>
          <Typography variant="body2" sx={{ color: 'darkGray' }}>
            The notification level to send emails.
          </Typography>
        </Box>
        <NotificationLevelSelector
          level={value?.level as NotificationLogLevel}
          onSelectLevel={(level) => onUpdateField('level', level)}
        />
      </Box>

    </Box>
  );
};
