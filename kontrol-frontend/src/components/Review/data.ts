const data = {
  edges: [
    {
      source: "gateway",
      target: "voting-app-ui--prod",
    },
    {
      source: "gateway",
      target: "voting-app-ui--dev",
    },
    {
      source: "voting-app-ui--prod",
      target: "redis-prod--prod",
    },
    {
      source: "voting-app-ui--dev",
      target: "kardinal-db-sidecar--dev",
    },
    {
      source: "kardinal-db-sidecar--dev",
      target: "redis-prod--prod",
    },
  ],
  nodes: [
    {
      id: "gateway",
      label: "gateway",
    },
    {
      id: "voting-app-ui",
      label: "voting-app-ui",
    },
    {
      id: "voting-app-ui--prod",
      label: "voting-app-ui (prod)",
      parent: "voting-app-ui",
    },
    {
      id: "voting-app-ui--dev",
      label: "voting-app-ui (dev)",
      parent: "voting-app-ui",
    },
    {
      id: "kardinal-db-sidecar--dev",
      label: "kardinal-db-sidecar (dev)",
    },
    {
      id: "redis-prod--prod",
      label: "redis-prod (prod)",
    },
  ],
};

export default data;
