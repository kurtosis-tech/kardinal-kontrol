import { ClusterTopology } from "@/types";

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
      versions: [
        {
          flowId: "k8s-namespace-1",
          imageTag: "kurtosistech/cartservice:main",
          isBaseline: true,
        },
        {
          flowId: "dev-hr7dwojzkk",
          imageTag: "kurtosistech/cartservice:demo-on-sale",
          isBaseline: false,
        },
      ],
    },
    {
      id: "checkoutservice",
      label: "checkoutservice",
      type: "service",
      versions: [
        {
          flowId: "k8s-namespace-1",
          imageTag: "kurtosistech/checkoutservice:main",
          isBaseline: true,
        },
        {
          flowId: "dev-hr7dwojzkk",
          imageTag: "kurtosistech/checkoutservice:demo-on-sale",
          isBaseline: false,
        },
      ],
    },
    {
      id: "frontend",
      label: "frontend",
      type: "service",
      versions: [
        {
          flowId: "k8s-namespace-1",
          imageTag: "kurtosistech/frontend:main",
          isBaseline: true,
        },
        {
          flowId: "dev-hr7dwojzkk",
          imageTag: "kurtosistech/frontend:demo-on-sale",
          isBaseline: false,
        },
      ],
    },
    {
      id: "emailservice",
      label: "emailservice",
      type: "service",
      versions: [
        {
          flowId: "k8s-namespace-1",
          imageTag: "kurtosistech/emailservice:main",
          isBaseline: true,
        },
      ],
    },
    {
      id: "paymentservice",
      label: "paymentservice",
      type: "service",
      versions: [
        {
          flowId: "k8s-namespace-1",
          imageTag: "kurtosistech/paymentservice:main",
          isBaseline: true,
        },
      ],
    },
    {
      id: "postgres",
      label: "postgres",
      type: "service",
      versions: [
        {
          flowId: "k8s-namespace-1",
          imageTag: "kurtosistech/postgres:main",
          isBaseline: true,
        },
      ],
    },
    {
      id: "shippingservice",
      label: "shippingservice",
      type: "service",
      versions: [
        {
          flowId: "k8s-namespace-1",
          imageTag: "kurtosistech/shippingservice:main",
          isBaseline: true,
        },
      ],
    },
    {
      id: "ingress",
      label: "ingress",
      type: "gateway",
      versions: [
        {
          flowId: "k8s-namespace-1",
          imageTag: "kurtosistech/gateway:main",
          isBaseline: true,
        },
      ],
    },
  ],
};

export default data;
