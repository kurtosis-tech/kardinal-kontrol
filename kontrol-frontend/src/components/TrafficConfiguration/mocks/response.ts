import { GraphData } from "../types";

const data: GraphData = {
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
      target: "productcatalogservice",
    },
    {
      source: "checkoutservice",
      target: "shippingservice",
    },
    {
      source: "frontend",
      target: "adservice",
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
      target: "productcatalogservice",
    },
    {
      source: "frontend",
      target: "recommendationservice",
    },
    {
      source: "frontend",
      target: "shippingservice",
    },
    {
      source: "frontend-external",
      target: "frontend",
    },
    {
      source: "recommendationservice",
      target: "productcatalogservice",
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
      id: "adservice",
      label: "adservice",
      type: "service",
      versions: ["prod"],
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
      id: "productcatalogservice",
      label: "productcatalogservice",
      type: "service",
      versions: ["prod"],
    },
    {
      id: "recommendationservice",
      label: "recommendationservice",
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
      id: "frontend-external",
      label: "frontend-external",
      type: "gateway",
      versions: [],
    },
  ],
};
export default data;
