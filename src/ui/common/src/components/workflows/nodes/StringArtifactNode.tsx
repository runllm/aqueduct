import { faFileText } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const stringArtifactNodeIcon = faFileText;

const StringArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={stringArtifactNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="String"
    />
  );
};

export default memo(StringArtifactNode);
