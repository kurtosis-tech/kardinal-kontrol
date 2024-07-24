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
    },
    {
      source: "gateway",
      target: "analytics-service:latest",
    },
    {
      source: "analytics-service:latest",
      target: "redis",
    },
    // postgres
    {
      source: "order-service:latest",
      target: "postgres",
    },
    {
      source: "order-service:dev",
      target: "kardinal-postgres-sidecar",
    },
    {
      source: "kardinal-postgres-sidecar",
      target: "postgres",
    },
    // redis
    {
      source: "order-service:latest",
      target: "redis",
    },
    {
      source: "order-service:dev",
      target: "kardinal-redis-sidecar",
    },
    {
      source: "kardinal-redis-sidecar",
      target: "redis",
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
      id: "order-service:group",
      label: "order-service",
    },
    // sidecars
    {
      id: "kardinal-redis-sidecar",
      label: "kardinal-redis-sidecar",
      type: "redis",
      parent: "redis:group",
    },
    {
      id: "kardinal-postgres-sidecar",
      label: "kardinal-postgres-sidecar",
      parent: "postgres:group",
      type: "postgres",
    },
    // service versions
    {
      id: "order-service:dev",
      parent: "order-service:group",
      label: "order-service:dev",
      type: "service-version",
    },
    {
      id: "order-service:latest",
      label: "order-service:latest",
      parent: "order-service:group",
      type: "service-version",
    },
    {
      id: "analytics-service:latest",
      label: "analytics-service:latest",
      parent: "analytics-service:group",
      type: "service-version",
    },
    // prod services
    {
      id: "redis",
      label: "redis",
      type: "redis",
      parent: "redis:group",
    },
    {
      id: "postgres",
      label: "postgres",
      type: "postgres",
      parent: "postgres:group",
    },
  ],
};

export default data;
