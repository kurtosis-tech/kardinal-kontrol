import { GraphData } from "../types";

const data: GraphData = {
  edges: [
    {
      source: "frontend-external",
      target: "frontend",
    },
    {
      source: "cartservice",
      target: "postgres",
    },
    {
      source: "checkoutservice",
      target: "paymentservice",
    },
    {
      source: "checkoutservice",
      target: "emailservice",
    },
    {
      source: "checkoutservice",
      target: "productcatalogservice",
    },
    {
      source: "checkoutservice",
      target: "cartservice",
    },
    {
      source: "checkoutservice",
      target: "shippingservice",
    },
    {
      source: "frontend",
      target: "checkoutservice",
    },
    {
      source: "frontend",
      target: "adservice",
    },
    {
      source: "frontend",
      target: "recommendationservice",
    },
    {
      source: "frontend",
      target: "productcatalogservice",
    },
    {
      source: "frontend",
      target: "cartservice",
    },
    {
      source: "frontend",
      target: "shippingservice",
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
      versions: ["prod", "dev-7uyjut175j"],
    },
    {
      id: "postgres",
      label: "postgres",
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
      id: "emailservice",
      label: "emailservice",
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
      id: "frontend",
      label: "frontend",
      type: "service",
      versions: ["prod", "dev-7uyjut175j"],
    },
    {
      id: "adservice",
      label: "adservice",
      type: "service",
      versions: ["prod"],
    },
    {
      id: "checkoutservice",
      label: "checkoutservice",
      type: "service",
      versions: ["prod", "dev-7uyjut175j"],
    },
    {
      id: "shippingservice",
      label: "shippingservice",
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
      id: "frontend-external",
      label: "frontend-external",
      type: "gateway",
    },
  ],
};
export default data;
