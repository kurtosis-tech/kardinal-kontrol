import cytoscapePopper from "cytoscape-popper";
import tippy, { Instance } from "tippy.js";
import { NodeVersion } from "@/types";
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
): Instance | null => {
  const versions: NodeVersion[] = node.data("versions");
  console.log("versions", versions);
  if (!versions || versions.length === 0) {
    return null;
  }
  const tip = node.popper({
    content: () => {
      const elem = document.createElement("div");
      // TODO: sanitize
      elem.innerHTML = `
      <table>
        <thead>
          <tr>
            <th>Flow ID</th>
            <th>Image Tag</th>
          </tr>
        </thead>
        <tbody>
          ${versions.map((v: NodeVersion) => `<tr><td>${v.flowId}</td><td>${v.imageTag}</td><tr>`).join("")}
        </tbody>
      </table>
      `;
      elem.classList.add("tooltip");
      return elem;
    },
  }) as unknown as Instance;

  tip.show();

  return tip;
};

export type TooltipInstance = Instance;

export default cytoscapePopper(tippyFactory);
