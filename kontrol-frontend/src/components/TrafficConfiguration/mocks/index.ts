import CytoscapeComponent from "react-cytoscapejs";
import cytoscape from "cytoscape";

import groupedByServiceData from "./grouped-by-service";
import groupedByVersionData from "./grouped-by-version";
import noDevFlowsData from "./no-dev-flows";
import nestedSidecarsData from "./nested-sidecars";
import coloredByEnvData from "./colored-by-env";
import devFlowData from "./dev-flow";
import devFlow2Data from "./dev-flow-2";
import manyFlowsData from "./many-flows";
import demoFlowData from "./demo-dev-flow";

export interface GraphData {
  nodes: cytoscape.NodeDataDefinition[];
  edges: cytoscape.EdgeDataDefinition[];
}

export const normalizeData = (data: GraphData) =>
  CytoscapeComponent.normalizeElements({
    nodes: data.nodes.map((node) => ({
      data: node,
      classes: node.classes,
    })),
    edges: data.edges.map((edge) => ({
      data: edge,
      classes: edge.classes,
    })),
  });

export const groupedByService = normalizeData(groupedByServiceData);
export const noDevFlows = normalizeData(noDevFlowsData);
export const groupedByVersion = normalizeData(groupedByVersionData);
export const nestedSidecars = normalizeData(nestedSidecarsData);
export const coloredByEnv = normalizeData(coloredByEnvData);
export const devFlow = normalizeData(devFlowData);
export const devFlow2 = normalizeData(devFlow2Data);
export const manyFlows = normalizeData(manyFlowsData);
export const demoFlow = normalizeData(demoFlowData);
