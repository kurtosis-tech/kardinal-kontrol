const getBackgroundImage = (ele: { data: (s: string) => string }) => {
  switch (ele.data("type")) {
    case "service-version":
      return "url('/icons/kubernetes.svg')";
    case "redis":
      return "url('/icons/redis.svg')";
    case "gateway":
      return "url('/icons/gateway.svg')";
    case "postgres":
      return "url('/icons/storage.svg')";
    default:
      return "none";
  }
};
const stylesheet = [
  {
    selector: "node",
    css: {
      shape: "round-rectangle",
      content: "data(label)",
      "text-valign": "bottom",
      "background-color": "#fff",
      "z-index": 2, // change?
      "background-image": getBackgroundImage,
      "background-fit": "cover",
      width: 50,
      height: 50,
    },
  },
  {
    selector: ":parent",
    css: {
      "text-valign": "top",
      "text-halign": "center",
      shape: "round-rectangle",
      "corner-radius": "10",
      "background-color": "#f5f5f5",
      padding: 10,
    },
  },
  {
    selector: "node[id^='traffic:']",
    css: {
      "corner-radius": "10",
      "background-color": "blue",
      padding: 0,
      width: 5,
      height: 5,
      content: "",
      "z-index": 1, // change?
      "border-width": "0px",
    },
  },
  {
    selector: "edge",
    css: {
      width: 1,
      "curve-style": "bezier",
      "target-arrow-shape": "triangle",
    },
  },
];

export default stylesheet;
