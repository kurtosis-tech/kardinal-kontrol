import type { components } from "cli-kontrol-api/api/typescript/client/types";

interface ServiceVersion {
  flow_id: string;
  image_tag: string;
  is_baseline?: boolean;
}
type OldNode = components["schemas"]["Node"];
interface NewNode extends OldNode {
  id: string;
  label: string;
  versions: ServiceVersion[];
}

export type Node = NewNode;
export type Edge = components["schemas"]["Edge"];

interface NewClusterTopology extends components["schemas"]["ClusterTopology"] {
  nodes: NewNode[];
}

export type ClusterTopology = NewClusterTopology;

export interface ExtendedNode {
  data: Node;
  classes: string;
}

// Type aliases for ease of use
export type Template = components["schemas"]["Template"];
export type Flow = components["schemas"]["Flow"];
