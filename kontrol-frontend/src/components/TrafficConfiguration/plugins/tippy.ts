import cytoscapePopper from "cytoscape-popper";
import tippy, { Instance } from "tippy.js";
import "./tippy.css";

// @ts-expect-error WIP
const tippyFactory: cytoscapePopper.PopperFactory = (ref, content) => {
  // Since tippy constructor requires DOM element/elements, create a placeholder
  const dummyDomEle = document.createElement("div");

  const tip = tippy(dummyDomEle, {
    getReferenceClientRect: ref.getBoundingClientRect,
    trigger: "manual", // mandatory
    // dom element inside the tippy:
    content: content,
    // your own preferences:
    arrow: true,
    placement: "right",
    hideOnClick: false,
    sticky: "reference",

    // if interactive:
    interactive: true,
    appendTo: document.body, // or append dummyDomEle to document.body
  });

  return tip;
};

export const createTooltip = (
  node: cytoscape.NodeSingular,
): Instance | null => {
  console.log("node: ", node);
  const versions = node.data("versions");
  if (!versions || versions.length === 0) {
    return null;
  }
  const tip = node.popper({
    content: () => {
      const elem = document.createElement("div");
      // TODO: sanitize
      elem.innerHTML = `Versions: <ul>${versions.map((v: string) => `<li>${v.toString()}</li>`).join("")}</ul>`;
      elem.classList.add("tooltip");
      return elem;
    },
  }) as unknown as Instance;

  tip.show();

  return tip;
};

export type TooltipInstance = Instance;

export default cytoscapePopper(tippyFactory);
