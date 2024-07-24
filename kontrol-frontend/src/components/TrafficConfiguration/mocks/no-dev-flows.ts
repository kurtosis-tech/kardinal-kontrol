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
      classes: "production",
    },
    {
      id: "analytics-service",
      label: "analytics-service",
      type: "service-version",
      classes: "production",
    },
    {
      id: "order-service",
      label: "order-service",
      type: "service-version",
      classes: "production",
    },
    {
      id: "stripe",
      label: "stripe",
      type: "stripe",
      classes: "production",
    },
    {
      id: "postgres",
      label: "postgres",
      type: "postgres",
      classes: "production",
    },
  ],
};

export default data;
