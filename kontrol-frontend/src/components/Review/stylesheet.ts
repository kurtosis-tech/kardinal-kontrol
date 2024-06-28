const stylesheet = [
  {
    selector: "node",
    css: {
      shape: "round-rectangle",
      content: "data(label)",
      "text-valign": "bottom",
      "border-color": "#ccc",
      "border-style": "solid",
      "border-width": "1px",
      "background-color": "#fff",
      "z-index": 2, // change?
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
    selector: "node[id ^= 'traffic:']",
    css: {
      "corner-radius": "10",
      "background-color": "green",
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
