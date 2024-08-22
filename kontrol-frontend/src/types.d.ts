import type { components } from "cli-kontrol-api/api/typescript/client/types";

export type ClusterTopology = components["schemas"]["ClusterTopology"];
export type Node = components["schemas"]["Node"];
export type Edge = components["schemas"]["Edge"];

export interface ExtendedNode {
  data: Node;
  classes: string;
}
