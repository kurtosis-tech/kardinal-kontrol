export interface ExtendedNodeDefinition extends cytoscape.NodeDataDefinition {
  versions: string[];
}

export interface GraphData {
  nodes: ExtendedNodeDefinition[];
  edges: cytoscape.EdgeDataDefinition[];
}
