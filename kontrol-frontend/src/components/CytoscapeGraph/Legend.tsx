import { Flex } from "@chakra-ui/react";
import { useApi } from "@/contexts/ApiContext";
import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
} from "@/components/Table";
import { FiExternalLink } from "react-icons/fi";
import { useEffect, useState } from "react";
import { ClusterTopology, NodeVersion, Node } from "@/types";
import { Link } from "@chakra-ui/react";

const Legend = () => {
  const { flows, getFlows, getTopology } = useApi();
  const [topology, setTopology] = useState<ClusterTopology | null>(null);

  const servicesForFlowId = (flowId: string, isBaseline: boolean): Node[] => {
    if (topology == null) {
      return [];
    }
    const services = topology.nodes.filter(
      (node) =>
        node.versions != null &&
        node.versions.some(
          (version: NodeVersion) =>
            version.flowId === flowId && version.isBaseline === isBaseline,
        ),
    );
    return services;
  };

  useEffect(() => {
    getFlows();
    getTopology().then(setTopology);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

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
              <Th>Baseline</Th>
              <Th>URL</Th>
              {/* <Th>Show/Hide</Th> */}
            </Tr>
          </Thead>
          <Tbody>
            {flows.map((flow) => {
              // TODO: Update these when this PR is merged:
              // https://github.com/kurtosis-tech/kardinal/pull/234
              const flowId = flow["flow-id"];
              const flowUrls = flow["flow-urls"];
              const isBaseline = servicesForFlowId(flowId, false).length === 0;
              return (
                <Tr key={flowId}>
                  <Td>{flowId}</Td>
                  <Td textTransform={"capitalize"}>{isBaseline.toString()}</Td>
                  <Td whiteSpace={"nowrap"}>
                    <Link
                      href={`http://${flowUrls[0]}`}
                      isExternal
                      target="_blank"
                      display={"flex"}
                      alignItems={"center"}
                      gap={2}
                    >
                      {flowUrls[0]}
                      <FiExternalLink size={12} />
                    </Link>
                  </Td>
                  {/* TODO: Uncomment when this feature is fully implemented
                   *
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
                          // setHighlightedFlowId(null);
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
                          // setHighlightedFlowId(flowId);
                        }}
                      />
                    )}
                  </Td>
                   **/}
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
