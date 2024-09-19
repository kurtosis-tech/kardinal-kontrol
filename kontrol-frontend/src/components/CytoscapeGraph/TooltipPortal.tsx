import { NodeVersion } from "@/types";
import { useEffect } from "react";
import ReactDOM from "react-dom";

interface Props {
  element: HTMLElement | null;
  nodeVersions: NodeVersion[] | null;
}

const TooltipPortal = ({ element, nodeVersions }: Props) => {
  useEffect(() => {
    // Cleanup function to remove the element when the component unmounts
    return () => {
      if (element != null && element.parentNode != null) {
        element.parentNode.removeChild(element);
      }
    };
  }, [element]);

  // Render the portal only when the dom element is available
  return element
    ? ReactDOM.createPortal(
        <div>
          <h1>This is rendered through a portal</h1>
        </div>,
        element,
      )
    : null;
};
export default TooltipPortal;
