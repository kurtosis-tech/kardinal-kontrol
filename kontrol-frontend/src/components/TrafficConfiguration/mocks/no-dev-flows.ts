import { GraphData } from "../types";

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
      versions: ["prod"],
    },
    {
      id: "analytics-service",
      label: "analytics-service",
      type: "service-version",
      classes: "production",
      versions: ["prod"],
    },
    {
      id: "order-service",
      label: "order-service",
      type: "service-version",
      classes: "production",
      versions: ["prod"],
    },
    {
      id: "stripe",
      label: "stripe",
      type: "stripe",
      classes: "production",
      versions: ["prod"],
    },
    {
      id: "postgres",
      label: "postgres",
      type: "postgres",
      classes: "production",
      versions: ["prod"],
    },
  ],
};

export default data;
