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
          flow_id: "k8s-namespace-1",
          image_tag: "kurtosistech/cartservice:main",
          is_baseline: true,
        },
        {
          flow_id: "dev-hr7dwojzkk",
          image_tag: "kurtosistech/cartservice:demo-on-sale",
        },
      ],
    },
    {
      id: "checkoutservice",
      label: "checkoutservice",
      type: "service",
      versions: [
        {
          flow_id: "k8s-namespace-1",
          image_tag: "kurtosistech/checkoutservice:main",
          is_baseline: true,
        },
        {
          flow_id: "dev-hr7dwojzkk",
          image_tag: "kurtosistech/checkoutservice:demo-on-sale",
        },
      ],
    },
    {
      id: "frontend",
      label: "frontend",
      type: "service",
      versions: [
        {
          flow_id: "k8s-namespace-1",
          image_tag: "kurtosistech/frontend:main",
          is_baseline: true,
        },
        {
          flow_id: "dev-hr7dwojzkk",
          image_tag: "kurtosistech/frontend:demo-on-sale",
        },
      ],
    },
    {
      id: "emailservice",
      label: "emailservice",
      type: "service",
      versions: [
        {
          flow_id: "k8s-namespace-1",
          image_tag: "kurtosistech/emailservice:main",
          is_baseline: true,
        },
      ],
    },
    {
      id: "paymentservice",
      label: "paymentservice",
      type: "service",
      versions: [
        {
          flow_id: "k8s-namespace-1",
          image_tag: "kurtosistech/paymentservice:main",
          is_baseline: true,
        },
      ],
    },
    {
      id: "postgres",
      label: "postgres",
      type: "service",
      versions: [
        {
          flow_id: "k8s-namespace-1",
          image_tag: "kurtosistech/postgres:main",
          is_baseline: true,
        },
      ],
    },
    {
      id: "shippingservice",
      label: "shippingservice",
      type: "service",
      versions: [
        {
          flow_id: "k8s-namespace-1",
          image_tag: "kurtosistech/shippingservice:main",
          is_baseline: true,
        },
      ],
    },
    {
      id: "ingress",
      label: "ingress",
      type: "gateway",
      versions: [
        {
          flow_id: "k8s-namespace-1",
          image_tag: "kurtosistech/gateway:main",
          is_baseline: true,
        },
      ],
    },
  ],
};

export default data;
