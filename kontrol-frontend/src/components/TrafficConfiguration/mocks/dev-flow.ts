import { GraphData } from "../types";

const data: GraphData = {
  edges: [
    {
      source: "gateway",
      target: "order-service",
    },
    {
      source: "gateway",
      target: "analytics-service",
    },
    {
      source: "analytics-service",
      target: "postgres",
    },
    {
      source: "order-service",
      target: "postgres",
    },
    {
      source: "order-service",
      target: "stripe",
    },
  ],

  nodes: [
    {
      id: "gateway",
      label: "gateway",
      type: "gateway",
      versions: ["prod"],
    },
    // service versions
    {
      id: "order-service",
      label: "order-service",
      expandedLabel: `order-service\n├─ dev\n└─ latest`,
      type: "service-version",
      versions: ["prod", "dev"],
    },
    {
      id: "analytics-service",
      label: "analytics-service",
      versions: ["prod"],
      type: "service-version",
    },
    // parent nodes
    {
      id: "stripe",
      label: "stripe",
      expandedLabel: `stripe\n├─ kardinal-stripe-plugin\n└─ stripe-production`,
      type: "stripe",
      versions: ["prod", "dev"],
    },
    {
      id: "postgres",
      label: "postgres",
      expandedLabel: `postgres\n├─ kardinal-postgres-sidecar\n└─ postgres-production`,
      type: "postgres",
      versions: ["prod", "dev"],
    },
  ],
};

export default data;
