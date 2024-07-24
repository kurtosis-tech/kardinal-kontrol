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
      id: "dev:group",
      label: "dev",
    },
    {
      id: "latest:group",
      label: "latest",
    },
    {
      id: "kontrol-plane:group",
      label: "kontrol-plane",
    },
    // sidecars
    {
      id: "kardinal-redis-sidecar",
      label: "kardinal-redis-sidecar",
      type: "redis",
      parent: "kontrol-plane:group",
    },
    {
      id: "kardinal-postgres-sidecar",
      label: "kardinal-postgres-sidecar",
      parent: "kontrol-plane:group",
      type: "postgres",
    },
    // service versions
    {
      id: "order-service:dev",
      parent: "dev:group",
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
    // prod services
    {
      id: "redis",
      label: "redis",
      type: "redis",
      parent: "latest:group",
    },
    {
      id: "postgres",
      label: "postgres",
      type: "postgres",
      parent: "latest:group",
    },
  ],
};

export default data;
