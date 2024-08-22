import dagre from "cytoscape-dagre";

export const dagreLayout = {
  name: "dagre",
  spacingFactor: 1.3,
  nodeSep: 40,
  nodeDimensionsIncludeLabels: true,
  // Possible values: 'network-simplex', 'tight-tree' or 'longest-path'
  ranker: "longest-path",
  rankDir: "LR",
  align: "UL",
};

export default dagre;
