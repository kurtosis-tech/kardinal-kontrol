import { ReactNode } from "react";
import { Flex } from "@chakra-ui/react";
import CytoscapeComponent from "react-cytoscapejs";
import cytoscape from "cytoscape";
// import sbgnStylesheet from "cytoscape-sbgn-stylesheet";
import dagre from "cytoscape-dagre";

cytoscape.use(dagre);

const layout = {
  name: "dagre",
  nodeSep: 50,
  nodeDimensionsIncludeLabels: true,
  rankDir: "LR",
  align: "UL",
};

const data = {
  edges: [
    {
      source: "gateway",
      target: "voting-app-ui (prod)",
    },
    {
      source: "gateway",
      target: "voting-app-ui (dev)",
    },
    {
      source: "voting-app-ui (prod)",
      target: "redis-prod (prod)",
    },
    {
      source: "voting-app-ui (dev)",
      target: "kardinal-db-sidecar (dev)",
    },
    {
      source: "kardinal-db-sidecar (dev)",
      target: "redis-prod (prod)",
    },
  ],
  nodes: [
    {
      id: "gateway",
      label: "gateway",
    },
    {
      id: "voting-app-ui",
      label: "voting-app-ui",
    },
    {
      id: "voting-app-ui (prod)",
      label: "voting-app-ui (prod)",
      parent: "voting-app-ui",
    },
    {
      id: "voting-app-ui (dev)",
      label: "voting-app-ui (dev)",
      parent: "voting-app-ui",
    },
    {
      id: "kardinal-db-sidecar (dev)",
      label: "kardinal-db-sidecar (dev)",
    },
    {
      id: "redis-prod (prod)",
      label: "redis-prod (prod)",
    },
  ],
};

const elements = [...data.nodes, ...data.edges].map((element) => ({
  data: element,
}));
// [
//   { data: { id: "one", label: "Node 1" }, position: { x: 0, y: 0 } },
//   { data: { id: "two", label: "Node 2" }, position: { x: 100, y: 0 } },
//   { data: { source: "one", target: "two", label: "Edge from Node1 to Node2" } },
// ];

// const stylesheet = [
//   {
//     selector: "node",
//     style: {
//       label: "data(label)",
//       shape: "round-rectangle",
//       height: 20,
//       width: 20,
//     },
//   },
//   {
//     selector: "edge",
//     style: {
//       width: 2,
//       "target-arrow-shape": "triangle-cross",
//       "line-style": "solid",
//       "curve-style": "bezier",
//     },
//   },
// ];
const stylesheet = [
  {
    selector: "node",
    css: {
      shape: "round-rectangle",
      content: "data(id)",
      "text-valign": "bottom",
      "border-color": "#ccc",
      "border-style": "solid",
      "border-width": "2px",
      "background-color": "#fff",
    },
  },
  {
    selector: ":parent",
    css: {
      "text-valign": "top",
      "text-halign": "center",
      shape: "round-rectangle",
      "corner-radius": "10",
      "background-color": "#f5f5f5",
      padding: 10,
    },
  },
  {
    selector: "node#e",
    css: {
      "corner-radius": "10",
      padding: 0,
    },
  },
  {
    selector: "edge",
    css: {
      width: 2,
      "curve-style": "bezier",
      "target-arrow-shape": "triangle-cross",
    },
  },
];
// interface SBGNStyle {
//   selector: string;
//   properties: Map<string, unknown>;
// }
const Review = ({ children }: { children?: ReactNode }) => {
  // const stylesheetArr: SBGNStyle[] = Array.from(sbgnStylesheet(cytoscape));
  // const stylesheet = stylesheetArr.map(({ selector, properties }) => ({
  //   selector,
  //   style: properties,
  // }));
  // console.log(Array.from(stylesheet));
  return (
    <Flex
      alignItems={"center"}
      justifyContent={"center"}
      height={"100%"}
      flexDir={"column"}
      gap={8}
    >
      {children}
      <CytoscapeComponent
        elements={elements}
        style={{ width: "100%", height: "100%" }}
        layout={layout}
        // @ts-expect-error cytoscape
        stylesheet={stylesheet}
      />
    </Flex>
  );
};

export default Review;
