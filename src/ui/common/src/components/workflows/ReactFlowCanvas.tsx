import React, { useEffect } from 'react';
import ReactFlow, {
  Node as ReactFlowNode,
  useReactFlow,
} from 'react-flow-renderer';
import { useSelector } from 'react-redux';

import { RootState } from '../../stores/store';
import { EdgeTypes, ReactFlowNodeData } from '../../utils/reactflow';
import nodeTypes from './nodes/nodeTypes';

const connectionLineStyle = { stroke: '#fff' };
const snapGrid = [20, 20];

type ReactFlowCanvasProps = {
  onPaneClicked: (event: React.MouseEvent) => void;
  switchSideSheet: (
    event: React.MouseEvent,
    element: ReactFlowNode<ReactFlowNodeData>
  ) => void;
};

const ReactFlowCanvas: React.FC<ReactFlowCanvasProps> = ({
  onPaneClicked,
  switchSideSheet,
}) => {
  const dagPositionState = useSelector(
    (state: RootState) => state.workflowReducer.selectedDagPosition
  );

  const currentNode = useSelector(
    (state: RootState) => state.nodeSelectionReducer.selected
  );

  const { fitView } = useReactFlow();
  useEffect(() => {
    setTimeout(fitView, 1000);
  }, [dagPositionState, fitView]);

  useEffect(() => {
    // NOTE(vikram): There's a timeout here because there seems to be a
    // race condition between calling `fitView` and the viewport
    // updating. This might be because of the width transition we use, but
    // we're not 100% sure.
    setTimeout(fitView, 100);
  }, [currentNode, fitView]);

  const { edges, nodes } = dagPositionState.result ?? { edges: [], nodes: [] };

  return (
    <ReactFlow
      onPaneClick={onPaneClicked}
      nodes={nodes}
      edges={edges}
      onNodeClick={switchSideSheet}
      nodeTypes={nodeTypes}
      connectionLineStyle={connectionLineStyle}
      snapToGrid={true}
      snapGrid={snapGrid as [number, number]}
      defaultZoom={1}
      edgeTypes={EdgeTypes}
      minZoom={0.25}
    />
  );
};

export default ReactFlowCanvas;
