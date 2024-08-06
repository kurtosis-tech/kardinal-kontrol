import CytoscapeComponent from "react-cytoscapejs";
import type { ExtendedNode, GraphData } from "./types";

export const extendNodeData = (
  node: cytoscape.NodeDataDefinition,
): ExtendedNode => {
  const versions = node.versions ?? ["UNKNOWN"];
  return {
    data: {
      ...node,
      versions,
    },
    classes: versions.length >= 2 ? "dev" : "prod",
  };
};

export const extendEdgeData =
  (nodes: ExtendedNode[]) =>
  (
    edge: cytoscape.EdgeDataDefinition,
  ): { data: cytoscape.EdgeDataDefinition; classes: string } => {
    const source = nodes.find((n) => n.data.id === edge.source);
    const sourceIsDev = source?.classes.includes("dev");
    const target = nodes.find((n) => n.data.id === edge.target);
    const targetIsDev = target?.classes.includes("dev");
    return {
      data: {
        ...edge,
        // overwrite the UUID returned from the API with a consistent
        // identifier so we can make renders in the UI idempotent to equivalent
        // topology responses
        id: `edge-${edge.source}-${edge.target}`,
      },
      classes: [
        targetIsDev ? "dev" : "prod",
        sourceIsDev && targetIsDev ? "ghost" : "",
      ].join(" "),
    };
  };

export const normalizeData = (data: GraphData) => {
  const nodes = data.nodes.map(extendNodeData);
  const edges = data.edges.map(extendEdgeData(nodes));
  return CytoscapeComponent.normalizeElements({
    nodes,
    // sort edges so that dev edges are rendered on top of prod edges
    edges: edges.sort((a: { classes: string }) =>
      a.classes.includes("dev") ? -1 : 1,
    ),
  });
};
