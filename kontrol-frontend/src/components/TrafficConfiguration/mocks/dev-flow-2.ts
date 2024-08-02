import { GraphData } from "../types";

const data: GraphData = {
  edges: [
    {
      source: "gateway",
      target: "order-service",
      classes: "dashed",
    },
    {
      source: "gateway",
      target: "analytics-service",
      classes: "dev2",
    },
    {
      source: "analytics-service",
      target: "postgres",
      classes: "dev2",
    },
    {
      source: "order-service",
      target: "postgres",
      classes: "dashed",
    },
    {
      source: "order-service",
      target: "stripe",
      classes: "dashed",
    },
  ],

  nodes: [
    {
      id: "gateway",
      label: "gateway",
      type: "gateway",
      classes: "production",
      versions: ["prod"],
    },
    // service versions
    {
      id: "order-service",
      label: "order-service",
      expandedLabel: `order-service\n├─ dev\n└─ latest`,
      classes: "production",
      type: "service-version",
      versions: ["prod"],
    },
    {
      id: "analytics-service",
      label: "analytics-service",
      expandedLabel: `analytics-service\n└─ latest`,
      classes: "dev2 ghost",
      type: "service-version",
      versions: ["prod", "dev"],
    },
    // parent nodes
    {
      id: "stripe",
      label: "stripe",
      expandedLabel: `stripe\n├─ kardinal-stripe-plugin\n└─ stripe-production`,
      type: "stripe",
      classes: "production",
      versions: ["prod"],
    },
    {
      id: "postgres",
      label: "postgres",
      expandedLabel: `postgres\n├─ kardinal-postgres-sidecar\n└─ postgres-production`,
      classes: "dev2 ghost",
      type: "postgres",
      versions: ["prod", "dev"],
    },
  ],
};

export default data;
