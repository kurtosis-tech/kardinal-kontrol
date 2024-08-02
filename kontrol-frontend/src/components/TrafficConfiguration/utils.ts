import { ExtendedNodeDefinition } from "./types";

export const enrichNodeData = (
  node: cytoscape.NodeDataDefinition,
): { data: ExtendedNodeDefinition; classes: string } => {
  const versions = node.versions ?? ["UNKNOWN"];
  return {
    data: {
      ...node,
      versions,
    },
    classes: versions.length >= 2 ? "dev" : "prod",
  };
};

export const enrichEdgeData = (
  edge: cytoscape.EdgeDataDefinition,
): { data: cytoscape.EdgeDataDefinition; classes: string } => {
  return {
    data: edge,
    classes: "prod",
  };
};
