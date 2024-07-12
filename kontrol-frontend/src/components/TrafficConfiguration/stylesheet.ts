export const trafficNodeSelector = "node[id^='traffic:']";

const getBackgroundImage = (ele: { data: (s: string) => string }) => {
  switch (ele.data("type")) {
    case "service-version":
      return "url('/icons/kubernetes.svg')";
    case "redis":
      return "url('/icons/redis.svg')";
    case "gateway":
      return "url('/icons/gateway.svg')";
    case "postgres":
      return "url('/icons/postgres.svg')";
    default:
      return "none";
  }
};
const stylesheet = [
  {
    selector: "node",
    css: {
      shape: "round-rectangle",
      "corner-radius": 12,
      content: "data(label)",
      "background-color": "#fff",
      "z-index": 2, // change?
      "background-image": getBackgroundImage,
      "background-size": "24px 24px",
      "background-repeat": "no-repeat",
      "background-position-x": 16,
      "text-margin-x": 16,
      "border-width": "1px",
      "text-halign": "center",
      "text-valign": "center",
      "font-size": 14,
      "font-family": "monospace",
      // the control over the height of the node has very strange behavior,
      // and this value (as close to zero as possible) is the only way to
      // reduce the height of the node
      height: 48,
      width: function (ele: cytoscape.NodeSingular) {
        const labelLength = ele.data("label").length;
        return labelLength * 8 + 72;
      },
    },
  },
  {
    selector: trafficNodeSelector,
    css: {
      "background-color": "#00D085",
      padding: 0,
      width: 5,
      height: 5,
      content: "",
      "z-index": 1, // change?
      "border-width": "2px",
      "border-color": "#00B372",
    },
  },
  {
    selector: "node[type = 'service-version']",
    css: {
      "background-color": "#EFF6FF",
      "border-color": "#58A6FF",
      color: "#2170CB",
    },
  },
  {
    selector: "node[type = 'redis']",
    css: {
      "background-color": "#FFF5F5",
      "border-color": "#FFA2A2",
      color: "#DD1E1E",
    },
  },
  {
    selector: "node[type = 'gateway']",
    css: {
      "background-color": "#FCF7FF",
      "border-color": "#DFAEF9",
      color: "#9053C6",
    },
  },
  {
    selector: "node[type = 'postgres']",
    css: {
      "background-color": "#FFF9ED",
      "border-color": "#FFB79F",
      color: "#EF5B2B",
    },
  },
  {
    selector: "edge",
    css: {
      "curve-style": "bezier",
      "line-style": "solid",
      width: 2,
      "line-color": "#DCDCDC",
    },
  },
  {
    selector: ":parent",
    css: {
      "text-valign": "top",
      "text-halign": "left",
      "text-margin-x": 112,
      "text-margin-y": -24,
      shape: "round-rectangle",
      "corner-radius": 12,
      "background-color": "#FAFAFA",
      "border-color": "#BFBFBF",
      "border-style": "dashed",
      padding: 32,
      color: "#525157",
    },
  },
];

export default stylesheet;
