import cytoscape from "cytoscape";
import CytoscapeComponent from "react-cytoscapejs";
interface GraphData {
  nodes: cytoscape.NodeDataDefinition[];
  edges: cytoscape.EdgeDataDefinition[];
}

const fullGraphData: GraphData = {
  edges: [
    {
      source: "gateway",
      target: "voting-app-ui--prod",
    },
    {
      source: "gateway",
      target: "voting-app-ui--dev",
    },
    {
      source: "voting-app-ui--prod",
      target: "redis-prod--prod",
    },
    {
      source: "voting-app-ui--dev",
      target: "kardinal-db-sidecar--dev",
    },
    {
      source: "kardinal-db-sidecar--dev",
      target: "redis-prod--prod",
    },
    {
      source: "voting-app-ui--prod",
      target: "postgres-prod--prod",
    },
  ],
  nodes: [
    {
      id: "gateway",
      label: "gateway",
      type: "gateway",
    },
    {
      id: "voting-app-ui",
      label: "voting-app-ui",
    },
    {
      id: "voting-app-ui--prod",
      label: "voting-app-ui (prod)",
      parent: "voting-app-ui",
      type: "service-version",
    },
    {
      id: "voting-app-ui--dev",
      label: "voting-app-ui (dev)",
      parent: "voting-app-ui",
      type: "service-version",
    },
    {
      id: "kardinal-db-sidecar--dev",
      label: "kardinal-db-sidecar (dev)",
      type: "redis",
    },
    {
      id: "redis-prod--prod",
      label: "redis-prod (prod)",
      type: "redis",
    },
    {
      id: "postgres-prod--prod",
      label: "postgres-prod (prod)",
      type: "postgres",
    },
  ],
};

const smallGraphData: GraphData = {
  edges: [
    {
      source: "gateway",
      target: "voting-app-ui--prod",
    },
    {
      source: "voting-app-ui--prod",
      target: "redis-prod--prod",
    },
  ],
  nodes: [
    {
      id: "gateway",
      label: "gateway",
      type: "gateway",
    },
    {
      id: "voting-app-ui",
      label: "voting-app-ui",
    },
    {
      id: "voting-app-ui--prod",
      label: "voting-app-ui (prod)",
      parent: "voting-app-ui",
      type: "service-version",
    },
    {
      id: "redis-prod--prod",
      label: "redis-prod (prod)",
      type: "redis",
    },
  ],
};

const normalizeData = (data: GraphData) =>
  CytoscapeComponent.normalizeElements({
    nodes: data.nodes.map((node) => ({
      data: node,
    })),
    edges: data.edges.map((edge) => ({
      data: edge,
    })),
  });

export const fullGraph = normalizeData(fullGraphData);
export const smallGraph = normalizeData(smallGraphData);
