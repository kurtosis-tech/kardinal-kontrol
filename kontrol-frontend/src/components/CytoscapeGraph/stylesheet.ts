export const trafficNodeSelector = "node[id^='traffic:']";

type Color = "blue" | "red" | "purple" | "orange" | "gray" | "green";

export const colors: Record<Color, string> = {
  blue: "#2170CB",
  red: "#DD1E1E",
  purple: "#9053C6",
  orange: "#EF5B2B",
  gray: "#8d8d8d",
  green: "#00D085",
};

const svgDot = (count: string, fill: string) =>
  "data:image/svg+xml;utf8," +
  encodeURIComponent(`
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle r="10" cx="12" cy="12" fill="${fill}" />
    <text x="12" y="16" text-anchor="middle" fill="white" font-family="monospace" font-weight="bold">${count}</text>
  </svg>
`);

const stylesheet = [
  {
    selector: "node",
    css: {
      shape: "round-rectangle",
      "corner-radius": 12,
      content: "data(label)",
      "background-image": (elem: cytoscape.NodeSingular) => {
        const versions = elem.data("versions");
        const isExternal = elem.data("type") === "external";
        return svgDot(
          isExternal ? "~" : versions.length.toString(),
          isExternal
            ? colors.purple
            : versions.length >= 2
              ? colors.orange
              : colors.gray,
        );
      },
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

      // taxi
      "curve-style": "round-taxi",
      "source-endpoint": "outside-to-node",
      "target-endpoint": "outside-to-node",
      "taxi-direction": "rightward",
      "taxi-turn": "70px",
      "taxi-radius": "48px",
      "target-arrow-shape": "triangle",
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
    selector: "edge.dev",
    css: {
      width: 3,
      "z-index": 10,
    },
  },
  {
    selector: ".ghost",
    css: {
      ghost: "yes",
      "ghost-offset-y": 12,
      "ghost-offset-x": 12,
      "ghost-opacity": 0.4,
    },
  },
  {
    selector: "edge.ghost",
    css: {
      "ghost-offset-x": 0,
      width: 4,
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
    selector: ".selected",
    css: {
      "underlay-color": "#999",
      "underlay-padding": 24,
      underlayOpacity: 0.2,
    },
  },
  {
    selector: ".external",
    css: {
      "line-style": "dashed",
      "border-style": "dotted",
      "background-color": "#f2ebf8",
      "border-color": colors.purple,
      color: "#5A5A59",
    },
  },
];

export default stylesheet;
