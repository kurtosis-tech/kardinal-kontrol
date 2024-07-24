import { GraphData } from ".";
import { colors } from "../stylesheet";

const data: GraphData = {
  edges: [
    {
      source: "gateway",
      target: "voting-app-ui",
      classes: "dev",
    },
    {
      source: "voting-app-ui",
      target: "redis",
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
      id: "voting-app-ui",
      label: "voting-app-ui",
      classes: "dev ghost dot dot",
      tooltip: [
        { name: "kurtosistech/demo-voting-app-ui", color: colors.gray },
        { name: "kurtosistech/demo-voting-app-ui-v2", color: colors.orange },
      ],
      type: "service-version",
    },
    // parent nodes
    {
      id: "redis",
      label: "redis",
      classes: "dev ghost dot dot",
      type: "redis",
      tooltip: [
        { name: "bitnami/redis:6.0.8", color: colors.gray },
        { name: "kurtosistech/redis-proxy-overlay", color: colors.orange },
      ],
    },
  ],
};

export default data;
