import { Checkbox, FormControlLabel } from '@mui/material';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import { useController, useFormContext } from 'react-hook-form';

import {
  FileData,
  GCSConfig,
  IntegrationDialogProps,
} from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationFileUploadField } from './IntegrationFileUploadField';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: GCSConfig = {
  bucket: 'aqueduct',
  use_as_storage: '',
};

interface GCSDialogProps extends IntegrationDialogProps {
  setMigrateStorage: React.Dispatch<React.SetStateAction<boolean>>;
}

export const GCSDialog: React.FC<GCSDialogProps> = ({
  editMode,
  setMigrateStorage,
}) => {
  // Setup for the checkbox component.
  const { control, setValue, getValues } = useFormContext();
  const { field } = useController({
    control,
    name: 'use_as_storage',
    defaultValue: 'true',
    rules: { required: true },
  });

  const [fileName, setFileName] = useState<string>(null);
  const setFile = (fileData: FileData | null) => {
    setFileName(fileData?.name ?? null);
    setValue('service_account_credentials', fileData?.data);
  };

  useEffect(() => {
    setMigrateStorage(true);
  }, [setMigrateStorage]);

  const fileData =
    fileName && !!getValues('service_account_credentials')
      ? {
          name: fileName,
          data: getValues('service_account_credentials'),
        }
      : null;

  const fileUploadDescription = (
    <>
      <>Follow the instructions </>
      <Link
        sx={{ fontSize: 'inherit' }}
        target="_blank"
        href="https://docs.aqueducthq.com/integrations/data-systems/non-sql-systems/google-cloud-storage"
      >
        here
      </Link>
      <> to create a service account and get the service account key file.</>
    </>
  );

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        name="bucket"
        spellCheck={false}
        required={true}
        label="Bucket*"
        description="The name of the GCS bucket."
        placeholder={Placeholders.bucket}
        onChange={(event) => setValue('bucket', event.target.value)}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disabled={editMode}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationFileUploadField
        name="service_account_credentials"
        label={'Service Account Credentials*'}
        description={fileUploadDescription}
        required={true}
        file={fileData}
        placeholder={'Upload your service account key file.'}
        onFiles={(files) => {
          const file = files[0];
          readCredentialsFile(file, setFile);
        }}
        displayFile={null}
        onReset={() => {
          setFile(null);
        }}
      />

      <FormControlLabel
        label="Use this integration for Aqueduct metadata storage."
        control={
          <Checkbox
            ref={field.ref}
            checked={field.value === 'true'}
            onChange={(event) => {
              const updatedValue = event.target.checked ? 'true' : 'false';
              field.onChange(updatedValue);
            }}
            disabled={true}
          />
        }
      />

      <Typography>
        We currently only support using Google Cloud Storage as the Aqueduct
        metadata storage. Support for using it as a data integration will be
        added soon.
      </Typography>
    </Box>
  );
};

export function readCredentialsFile(
  file: File,
  setFile: (credentials: FileData) => void
): void {
  const reader = new FileReader();
  reader.onloadend = function (event) {
    const content = event.target.result as string;
    setFile({ name: file.name, data: content });
  };
  reader.readAsText(file);
}

export function isGCSConfigComplete(config: GCSConfig): boolean {
  return !!config.bucket && !!config.service_account_credentials;
}
