import { GraphData } from ".";

const data: GraphData = {
  edges: [
    {
      source: "gateway",
      target: "order-service:latest",
    },
    {
      source: "gateway",
      target: "order-service:dev",
      classes: "dashed",
    },
    {
      source: "gateway",
      target: "analytics-service:latest",
    },
    {
      source: "analytics-service:latest",
      target: "redis:group",
    },
    // postgres
    {
      source: "order-service:latest",
      target: "postgres:group",
    },
    {
      source: "order-service:dev",
      target: "kardinal-postgres-sidecar",
      classes: "dashed",
    },
    {
      source: "order-service:latest",
      target: "redis:group",
    },
    {
      source: "order-service:dev",
      target: "kardinal-redis-sidecar",
      classes: "dashed",
    },
  ],
  nodes: [
    {
      id: "gateway",
      label: "gateway",
      type: "gateway",
    },
    // parent nodes
    {
      id: "redis:group",
      label: "redis:latest",
      type: "redis",
      classes: "parent-service",
    },
    {
      id: "postgres:group",
      label: "postgres:latest",
      classes: "parent-service",
      type: "postgres",
    },
    // sidecars
    {
      id: "kardinal-redis-sidecar",
      classes: "sidecar",
      label: "kardinal-redis-sidecar",
      type: "redis",
      parent: "redis:group",
    },
    {
      id: "kardinal-postgres-sidecar",
      classes: "sidecar",
      label: "kardinal-postgres-sidecar",
      parent: "postgres:group",
      type: "postgres",
    },
    // service versions
    {
      id: "order-service:dev",
      classes: "dashed",
      label: "order-service:dev",
      type: "service-version",
    },
    {
      id: "order-service:latest",
      label: "order-service:latest",
      parent: "latest:group",
      type: "service-version",
    },
    {
      id: "analytics-service:latest",
      label: "analytics-service:latest",
      parent: "latest:group",
      type: "service-version",
    },
  ],
};

export default data;
