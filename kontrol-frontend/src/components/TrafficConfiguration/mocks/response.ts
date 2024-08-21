import { ClusterTopology } from "../types";

const data: ClusterTopology = {
  edges: [
    {
      source: "cartservice",
      target: "postgres",
    },
    {
      source: "checkoutservice",
      target: "cartservice",
    },
    {
      source: "checkoutservice",
      target: "emailservice",
    },
    {
      source: "checkoutservice",
      target: "paymentservice",
    },
    {
      source: "checkoutservice",
      target: "shippingservice",
    },
    {
      source: "frontend",
      target: "cartservice",
    },
    {
      source: "frontend",
      target: "checkoutservice",
    },
    {
      source: "frontend",
      target: "shippingservice",
    },
    {
      source: "ingress",
      target: "frontend",
    },
  ],
  nodes: [
    {
      id: "cartservice",
      label: "cartservice",
      type: "service",
      versions: ["dev-hr7dwojzkk", "prod"],
    },
    {
      id: "checkoutservice",
      label: "checkoutservice",
      type: "service",
      versions: ["dev-hr7dwojzkk", "prod"],
    },
    {
      id: "frontend",
      label: "frontend",
      type: "service",
      versions: ["dev-hr7dwojzkk", "prod"],
    },
    {
      id: "emailservice",
      label: "emailservice",
      type: "service",
      versions: ["prod"],
    },
    {
      id: "paymentservice",
      label: "paymentservice",
      type: "service",
      versions: ["prod"],
    },
    {
      id: "postgres",
      label: "postgres",
      type: "service",
      versions: ["prod"],
    },
    {
      id: "shippingservice",
      label: "shippingservice",
      type: "service",
      versions: ["prod"],
    },
    {
      id: "ingress",
      label: "ingress",
      type: "gateway",
      versions: ["prod"],
    },
  ],
};

export default data;
