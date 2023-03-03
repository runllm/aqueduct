import React from 'react';
import ReactFlow, { Node as ReactFlowNode } from 'react-flow-renderer';
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

  const { edges, nodes } = dagPositionState.result ?? { edges: [], nodes: [] };

  const canvasEdges = edges.map((edge) => {
    return {
      id: edge.id,
      source: edge.source,
      target: edge.target,
      type: edge.type,
      container: 'root',
    };
  });

  const canvasNodes = nodes.map((node) => {
    return {
      id: node.id,
      type: node.type,
      data: node.data,
      position: node.position,
    };
  });

  return (
    <ReactFlow
      onPaneClick={onPaneClicked}
      nodes={canvasNodes}
      edges={canvasEdges}
      onNodeClick={switchSideSheet}
      nodeTypes={nodeTypes}
      connectionLineStyle={connectionLineStyle}
      snapToGrid={true}
      snapGrid={snapGrid as [number, number]}
      defaultZoom={1}
      edgeTypes={EdgeTypes}
      minZoom={0.25}
      fitView={true}
    />
  );
};

export default ReactFlowCanvas;
