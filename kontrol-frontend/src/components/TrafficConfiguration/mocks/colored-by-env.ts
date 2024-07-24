import { GraphData } from ".";

const data: GraphData = {
  edges: [
    {
      source: "gateway",
      target: "order-service:latest",
      classes: "dashed",
    },
    {
      source: "gateway",
      target: "order-service:dev",
      classes: "dev",
    },
    {
      source: "gateway",
      target: "analytics-service:latest",
      classes: "dashed",
    },
    {
      source: "analytics-service:latest",
      target: "redis:group",
      classes: "dashed",
    },
    // postgres
    {
      source: "order-service:latest",
      target: "postgres:group",
      classes: "dashed",
    },
    {
      source: "order-service:dev",
      target: "kardinal-postgres-sidecar",
      classes: "dev",
    },
    {
      source: "order-service:latest",
      target: "redis:group",
      classes: "dashed",
    },
    {
      source: "order-service:dev",
      target: "kardinal-redis-sidecar",
      classes: "dev",
    },
  ],
  nodes: [
    {
      id: "gateway",
      label: "gateway",
      type: "gateway",
      classes: "production dashed",
    },
    // parent nodes
    {
      id: "redis:group",
      label: "redis",
      type: "redis",
      classes: "parent-service production dashed",
    },
    {
      id: "postgres:group",
      label: "postgres",
      classes: "parent-service production dashed",
      type: "postgres",
    },
    // sidecars
    {
      id: "kardinal-redis-sidecar",
      classes: "sidecar dev",
      label: "kardinal-redis-sidecar",
      type: "redis",
      parent: "redis:group",
    },
    {
      id: "kardinal-postgres-sidecar",
      classes: "sidecar dev",
      label: "kardinal-postgres-sidecar",
      parent: "postgres:group",
      type: "postgres",
    },
    // service versions
    {
      id: "order-service:dev",
      classes: "dev",
      label: "order-service [dev]",
      type: "service-version",
    },
    {
      id: "order-service:latest",
      label: "order-service",
      classes: "production dashed",
      parent: "latest:group",
      type: "service-version",
    },
    {
      id: "analytics-service:latest",
      label: "analytics-service",
      classes: "production dashed",
      parent: "latest:group",
      type: "service-version",
    },
  ],
};

export default data;
