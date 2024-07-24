import CytoscapeGraph from "@/components/TrafficConfiguration/CytoscapeGraph";
import * as mockData from "@/components/TrafficConfiguration/mocks";
import cytoscape from "cytoscape";

const layout = {
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

const Page = ({ variant }: { variant: "dev" | "dev2" | "main" | "all" }) => {
  return (
    <CytoscapeGraph
      elements={
        variant === "dev"
          ? mockData.devFlow
          : variant === "dev2"
            ? mockData.devFlow2
            : variant === "all"
              ? mockData.manyFlows
              : mockData.noDevFlows
      }
      layout={layout}
    />
  );
};

export default Page;
