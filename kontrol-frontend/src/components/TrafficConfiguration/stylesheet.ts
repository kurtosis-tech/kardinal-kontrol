export const trafficNodeSelector = "node[id^='traffic:']";

import * as icons from "./icons";

type Color = "blue" | "red" | "purple" | "orange" | "gray" | "green";

export const colors: Record<Color, string> = {
  blue: "#2170CB",
  red: "#DD1E1E",
  purple: "#9053C6",
  orange: "#EF5B2B",
  gray: "#8d8d8d",
  green: "#00D085",
};

const svgDot = (x: number, y: number, fill: string) =>
  `<circle r="4" cx="${x}" cy="${y}" fill="${fill}" />`;

const svgSrc = (elems: string[]) =>
  `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">${elems.join("")}</svg>`;

const encodeDataUri = (data: string) =>
  `data:image/svg+xml;utf8,${encodeURIComponent(data)}`;

const devSvgSrc = (dotCount: number, cols: Color[]) => {
  const dx = dotCount === 1 ? 4 : 0;
  const dy = dotCount <= 2 ? 8 : 0;
  return svgSrc([
    dotCount >= 1 ? svgDot(4 + dx, 4 + dy, colors[cols[0]] || colors.gray) : "",
    dotCount >= 2 ? svgDot(16 + dx, 4 + dy, colors[cols[1]]) : "",
    dotCount >= 3 ? svgDot(4 + dx, 16 + dy, colors[cols[2]]) : "",
    dotCount >= 4 ? svgDot(16 + dx, 16 + dy, colors[cols[3]]) : "",
  ]);
};

const inlineSvgBackground = (ele: { data: (s: string) => string }) => {
  const classes = ele.data("classes") || "";
  const dotCount = classes.match(/dot/g)?.length || 0;
  return encodeDataUri(
    devSvgSrc(
      dotCount,
      classes.includes("dev2")
        ? ["purple", "green"]
        : ["gray", "orange", "blue"],
    ),
  );
};

// const summaryBackground = (count: number) => {
//   return encodeDataUri(
//     svgSrc([`<text x="4" y="4" color="#000">${count}</text>`]),
//   );
// };
//
const getBackgroundImage = (ele: { data: (s: string) => string }) => {
  const type = ele.data("type");
  const classes = ele.data("classes") || "";

  // if (classes.includes("summary")) {
  //   // return summaryBackground(1);
  //   return "";
  // }

  return inlineSvgBackground(ele);
  if (classes.includes("dot")) {
  }

  const color = classes.includes("production")
    ? colors.gray // prod
    : classes.includes("dev2")
      ? colors.purple // dev2
      : colors.orange; // dev
  switch (type) {
    case "gateway":
      return encodeDataUri(icons.gateway(color));
    case "service-version":
      return encodeDataUri(icons.serviceVersion(color));
    case "postgres":
      return encodeDataUri(icons.postgres(color));
    case "stripe":
      return encodeDataUri(icons.stripe(color));
    case "redis":
      return encodeDataUri(icons.redis(color));
    default:
      return encodeDataUri("");
  }
};

// const nestedParentStyles = {
//   "border-style": "solid",
//   "text-valign": "center",
// };

const stylesheet = [
  {
    selector: "node",
    css: {
      shape: "round-rectangle",
      "corner-radius": 12,
      content: "data(label)",
      // "background-color": "#fff",
      // "z-index": 2, // change?
      "background-image": getBackgroundImage,
      "background-size": "24px 24px",
      "background-repeat": "no-repeat",
      "background-position-x": 16,
      "background-position-y": "50%",
      "text-margin-x": 16,
      "border-width": "1px",
      "text-halign": "center",
      "text-valign": "center",
      "font-size": 14,
      "font-family": "monospace",
      "background-color": "#EFF6FF",
      "border-color": "#c0bfbf",
      color: "#5A5A59",

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
  // {
  //   selector: trafficNodeSelector,
  //   css: {
  //     "background-color": "#00D085",
  //     padding: 0,
  //     width: 5,
  //     height: 5,
  //     content: "",
  //     "z-index": 1, // change?
  //     "border-width": "2px",
  //     "border-color": "#00B372",
  //   },
  // },
  // {
  //   selector: ":parent",
  //   css: {
  //     "text-valign": "top",
  //     "text-halign": "center",
  //     // "text-margin-y": 24,
  //     shape: "round-rectangle",
  //     "corner-radius": 12,
  //     "background-color": "#FAFAFA",
  //     "border-color": "#BFBFBF",
  //     padding: 48,
  //     color: "#525157",
  //   },
  // },
  // {
  //   selector: "node[type = 'service-version']",
  //   css: {
  //     "background-color": "#EFF6FF",
  //     "border-color": "#58A6FF",
  //     color: "#2170CB",
  //     ...nestedParentStyles,
  //   },
  // },
  // {
  //   selector: "node[type = 'redis']",
  //   css: {
  //     "background-color": "#FFF5F5",
  //     "border-color": "#FFA2A2",
  //     color: "#DD1E1E",
  //     ...nestedParentStyles,
  //   },
  // },
  // {
  //   selector: "node[type = 'gateway']",
  //   css: {
  //     "background-color": "#FCF7FF",
  //     "border-color": "#DFAEF9",
  //     color: "#9053C6",
  //     ...nestedParentStyles,
  //   },
  // },
  // {
  //   selector: "node[type = 'postgres']",
  //   css: {
  //     "background-color": "#FFF9ED",
  //     "border-color": "#FFB79F",
  //     color: "#EF5B2B",
  //     ...nestedParentStyles,
  //   },
  // },
  // {
  //   selector: ".parent-service",
  //   css: {
  //     "text-margin-x": 0,
  //     "text-valign": "bottom",
  //     padding: 16,
  //     "text-margin-y": -28,
  //     "background-offset-y": 24,
  //   },
  // },
  // {
  //   selector: ".sidecar",
  //   css: {
  //     height: 24,
  //     "background-image": "none",
  //     "text-margin-x": 0,
  //     width: function (ele: cytoscape.NodeSingular) {
  //       const labelLength = ele.data("label").length;
  //       return labelLength * 8 + 32;
  //     },
  //   },
  // },
  {
    selector: ".production",
    css: {
      // "background-color": "#EAEBF0",
      "background-color": "#EFF6FF",
      "border-color": "#c0bfbf",
      color: "#5A5A59",
    },
  },
  {
    selector: "edge",
    css: {
      // "edge-distances": "endpoints",
      // base
      "line-style": "solid",
      width: 2,
      "line-color": "#DCDCDC",

      // straight
      // "source-endpoint": "90deg",
      // "target-endpoint": "-90deg",
      // "curve-style": "straight",

      // taxi
      "curve-style": "round-taxi",
      "source-endpoint": "outside-to-node",
      "target-endpoint": "outside-to-node",
      "taxi-direction": "rightward",
      "taxi-turn": "70px",
      "target-arrow-shape": "triangle",
    },
  },
  {
    selector: ".dev2",
    css: {
      "background-color": "#FCF7FF",
      "border-color": "#DFAEF9",
      color: "#9053C6",
      "line-color": "#DFAEF9",
      "target-arrow-color": "#DFAEF9",
      "z-index": 3,
    },
  },
  {
    selector: ".dev",
    css: {
      "background-color": "#FFF9ED",
      "border-color": "#FFB79F",
      "line-color": "#FFB79F",
      "target-arrow-color": "#FFB79F",
      color: "#ef5b2b",
    },
  },
  {
    selector: ".dashed",
    css: {
      "line-style": "dotted",
      "border-style": "dotted",
    },
  },
  {
    selector: ".ghost",
    css: {
      ghost: "yes",
      "ghost-offset-y": 12,
      "ghost-offset-x": 12,
      "ghost-opacity": 0.5,
      // "underlay-color": "#999",
      // "underlay-padding": 24,
      // underlayOpacity: 0.2,
    },
  },
  {
    selector: ".selected",
    css: {
      // height: 72,
      // content: "data(expandedLabel)",
      // borderWidth: 3,
      // "text-wrap": "wrap",
      "underlay-color": "#999",
      "underlay-padding": 24,
      underlayOpacity: 0.2,
      // "line-height": 1.2,
      // "text-justification": "left",
      // padding: 0,
      // width: function (ele: cytoscape.NodeSingular) {
      //   let labelLength = ele.data("expandedLabel")?.length;
      //   if (!labelLength) {
      //     labelLength = ele.data("label")?.length;
      //   }
      //   return Math.min(labelLength * 8 + 72, 350);
      // },
    },
  },
  // {
  //   selector: ".summary",
  //   css: {
  //     "text-margin-x": 0,
  //     "background-image": "none",
  //   },
  // },
];

export default stylesheet;
