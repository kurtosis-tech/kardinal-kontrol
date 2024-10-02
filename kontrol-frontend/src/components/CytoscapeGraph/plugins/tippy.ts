import cytoscapePopper from "cytoscape-popper";
import tippy, { Instance } from "tippy.js";
import { NodeVersion } from "@/types";
import "./tippy.css";

// @ts-expect-error Poor typing upstream
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
    placement: "bottom",
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
): { instance: Instance; element: HTMLElement } | null => {
  const versions: NodeVersion[] = node.data("versions");
  if (!versions || versions.length === 0) {
    return null;
  }

  const element = document.createElement("div");
  element.classList.add("tooltip");

  const instance = node.popper({
    // tooltip content is rendered through the TooltipPortal component
    content: () => element,
  }) as unknown as Instance;

  instance.show();

  return { instance, element };
};

export type TooltipInstance = Instance;

export default cytoscapePopper(tippyFactory);
