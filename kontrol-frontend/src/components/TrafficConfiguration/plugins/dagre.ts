import dagre from "cytoscape-dagre";

export const dagreLayout = {
  name: "dagre",
  spacingFactor: 1.6,
  nodeSep: 40,
  nodeDimensionsIncludeLabels: true,
  // ranker: "network-simplex",
  rankDir: "LR",
  align: "UL",
  transform: (_: unknown, pos: cytoscape.Position) => {
    const roundedPos = {
      ...pos,
      y: Math.round(Math.ceil(pos.y / 100)) * 100,
    };
    return roundedPos;
  },
};

export default dagre;
