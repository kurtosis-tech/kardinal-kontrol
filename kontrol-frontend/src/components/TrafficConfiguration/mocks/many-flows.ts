import { GraphData } from ".";

const data: GraphData = {
  edges: [
    {
      source: "gateway",
      target: "order-service",
      classes: "",
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
      classes: "",
    },
    {
      source: "order-service",
      target: "stripe",
      classes: "",
    },
  ],

  nodes: [
    {
      id: "gateway",
      label: "gateway",
      type: "gateway",
      classes: "production summary",
    },
    // service versions
    {
      id: "order-service",
      label: "order-service [5 flows] ",
      expandedLabel: `order-service\n├─ dev\n└─ latest`,
      classes: "production ghost summary",
      summaryCount: 5,
      type: "service-version",
    },
    {
      id: "analytics-service",
      label: "analytics-service [3 flows] ",
      expandedLabel: `analytics-service\n└─ latest`,
      classes: "production summary ghost",
      summaryCount: 4,
      type: "service-version",
    },
    // parent nodes
    {
      id: "stripe",
      label: "stripe [5 flows] ",
      expandedLabel: `stripe\n├─ kardinal-stripe-plugin\n└─ stripe-production`,
      type: "stripe",
      classes: "production ghost summary",
      summaryCount: 2,
    },
    {
      id: "postgres",
      label: "postgres [8 flows] ",
      expandedLabel: `postgres\n├─ kardinal-postgres-sidecar\n└─ postgres-production`,
      classes: "production ghost summary",
      summaryCount: 4,
      type: "postgres",
    },
  ],
};

export default data;
