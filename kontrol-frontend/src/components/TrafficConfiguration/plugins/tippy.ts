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

interface TooltipContentItem {
  color: string;
  name: string;
}

export const createTooltip = (
  node: cytoscape.NodeSingular,
  items: TooltipContentItem[],
): Instance => {
  const tip = node.popper({
    content: () => {
      const elem = document.createElement("div");
      // TODO: sanitize
      elem.innerHTML = items
        .map(
          (item) =>
            `<span class="dot" style="background-color:${item.color}"></span>${item.name}`,
        )
        .join("<br/>");
      elem.classList.add("tooltip");
      // elem.style.padding = "10p";
      // elem.style.backgroundColor = "white";
      // elem.style.borderRadius = "5px";
      // elem.style.boxShadow = "0 0 10px 0 rgba(0, 0, 0, 0.1)";
      return elem;
    },
  }) as unknown as Instance;

  tip.show();

  return tip;
};

export type TooltipInstance = Instance;

export default cytoscapePopper(tippyFactory);
