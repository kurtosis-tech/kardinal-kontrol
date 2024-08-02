import CytoscapeComponent from "react-cytoscapejs";

import noDevFlowsData from "./no-dev-flows";
import devFlowData from "./dev-flow";
import devFlow2Data from "./dev-flow-2";
import type { GraphData } from "../types";
import { enrichEdgeData, enrichNodeData } from "../utils";

export const normalizeData = (data: GraphData) =>
  CytoscapeComponent.normalizeElements({
    nodes: data.nodes.map(enrichNodeData),
    edges: data.edges.map(enrichEdgeData),
  });

export const noDevFlows = normalizeData(noDevFlowsData);
export const devFlow = normalizeData(devFlowData);
export const devFlow2 = normalizeData(devFlow2Data);
