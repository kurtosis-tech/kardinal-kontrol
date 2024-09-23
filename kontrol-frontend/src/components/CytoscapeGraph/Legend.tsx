import { NodeVersion } from "@/types";
import { Flex, IconButton } from "@chakra-ui/react";
import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
} from "@/components/Table";
import { useState } from "react";
import { FiEye, FiEyeOff } from "react-icons/fi";

interface Props {
  elements: cytoscape.ElementDefinition[];
}

const Legend = ({ elements }: Props) => {
  const [highlightedFlowId, setHighlightedFlowId] = useState<string | null>(
    null,
  );

  const serviceVersions: NodeVersion[] = elements
    .map((element) => element.data.versions)
    .flat()
    .filter(Boolean);

  const flowIds = serviceVersions.map((version) => version.flowId);
  const uniqueFlowIds = [...new Set(flowIds)];

  const servicesForFlowId = (flowId: string): cytoscape.ElementDefinition[] => {
    const services = elements.filter(
      (element) =>
        element.data.versions != null &&
        element.data.versions.some(
          (version: NodeVersion) =>
            version.flowId === flowId && !version.isBaseline,
        ),
    );
    return services;
  };

  return (
    <Flex
      position={"absolute"}
      bg={"white"}
      minH={100}
      minW={240}
      top={0}
      left={0}
      borderRadius={12}
      zIndex={5}
    >
      <TableContainer>
        <Table variant="simple">
          <Thead>
            <Tr>
              <Th>Flow ID</Th>
              <Th>Deployed Services</Th>
              <Th>Baseline</Th>
              <Th>Show/Hide</Th>
            </Tr>
          </Thead>
          <Tbody>
            {uniqueFlowIds.map((flowId) => {
              const sffi = servicesForFlowId(flowId);
              const isBaseline = sffi.length === 0;
              return (
                <Tr key={flowId}>
                  <Td>{flowId}</Td>
                  <Td>{sffi.length === 0 ? "All" : sffi.length}</Td>
                  <Td textTransform={"capitalize"}>{isBaseline.toString()}</Td>
                  <Td textAlign={"right"}>
                    {highlightedFlowId === flowId ? (
                      <IconButton
                        aria-label="Hide"
                        h={"24px"}
                        minH={"24px"}
                        w={"24px"}
                        minW={"24px"}
                        color={isBaseline ? "gray.300" : "gray.600"}
                        icon={<FiEyeOff size={16} />}
                        disabled={isBaseline}
                        onClick={() => {
                          setHighlightedFlowId(null);
                        }}
                      />
                    ) : (
                      <IconButton
                        aria-label="Show"
                        h={"24px"}
                        minH={"24px"}
                        w={"24px"}
                        minW={"24px"}
                        icon={<FiEye size={16} />}
                        disabled={isBaseline}
                        color={isBaseline ? "gray.300" : "gray.600"}
                        cursor={isBaseline ? "not-allowed" : "pointer"}
                        onClick={() => {
                          setHighlightedFlowId(flowId);
                        }}
                      />
                    )}
                  </Td>
                </Tr>
              );
            })}
          </Tbody>
        </Table>
      </TableContainer>
    </Flex>
  );
};

export default Legend;
