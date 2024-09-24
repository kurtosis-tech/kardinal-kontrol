import { useEffect } from "react";
import ReactDOM from "react-dom";
import cytoscape from "cytoscape";
import { NodeVersion } from "@/types";
import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
} from "@/components/Table";

interface Props {
  element: HTMLElement | null;
  node: cytoscape.NodeSingular | null;
}

const TooltipPortal = ({ element, node }: Props) => {
  useEffect(() => {
    // Cleanup function to remove the element when the component unmounts
    return () => {
      if (element != null && element.parentNode != null) {
        element.parentNode.removeChild(element);
      }
    };
  }, [element]);

  if (element == null || node == null) {
    return null;
  }

  const versions: NodeVersion[] = node.data("versions");

  // Render the portal only when the dom element is available
  return ReactDOM.createPortal(
    <TableContainer>
      <Table>
        <Thead>
          <Tr>
            <Th>Flow ID</Th>
            <Th>Image Tag</Th>
          </Tr>
        </Thead>
        <Tbody>
          {versions.map(({ flowId, imageTag }) => {
            return (
              <Tr key={flowId}>
                <Td>{flowId}</Td>
                <Td>{imageTag ?? "N/A (external service)"}</Td>
              </Tr>
            );
          })}
        </Tbody>
      </Table>
    </TableContainer>,

    element,
  );
};
export default TooltipPortal;
