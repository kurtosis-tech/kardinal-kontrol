import { GraphData } from ".";

const data: GraphData = {
  edges: [
    {
      source: "gateway",
      target: "order-service",
      classes: "dev",
    },
    {
      source: "gateway",
      target: "analytics-service",
      classes: "",
    },
    {
      source: "analytics-service",
      target: "postgres",
      classes: "",
    },
    {
      source: "order-service",
      target: "postgres",
      classes: "dev",
    },
    {
      source: "order-service",
      target: "stripe",
      classes: "dev",
    },
  ],

  nodes: [
    {
      id: "gateway",
      label: "gateway",
      type: "gateway",
      classes: "production dot",
    },
    // service versions
    {
      id: "order-service",
      label: "order-service",
      expandedLabel: `order-service\n├─ dev\n└─ latest`,
      classes: "dev ghost dot dot",
      type: "service-version",
    },
    {
      id: "analytics-service",
      label: "analytics-service",
      expandedLabel: `analytics-service\n└─ latest`,
      classes: "production dot",
      type: "service-version",
    },
    // parent nodes
    {
      id: "stripe",
      label: "stripe",
      expandedLabel: `stripe\n├─ kardinal-stripe-plugin\n└─ stripe-production`,
      type: "stripe",
      classes: "dev ghost dot dot",
    },
    {
      id: "postgres",
      label: "postgres",
      expandedLabel: `postgres\n├─ kardinal-postgres-sidecar\n└─ postgres-production`,
      classes: "dev ghost dot dot",
      type: "postgres",
    },
  ],
};

export default data;
