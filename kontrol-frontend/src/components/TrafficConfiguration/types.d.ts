export interface ExtendedNodeData extends cytoscape.NodeDataDefinition {
  versions?: string[];
}

export interface GraphData {
  nodes: ExtendedNodeData[];
  edges: cytoscape.EdgeDataDefinition[];
}

export interface ExtendedNode {
  data: ExtendedNodeData;
  classes: string;
}
