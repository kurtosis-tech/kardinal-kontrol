import { Flex, IconButton } from "@chakra-ui/react";
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
import { FiExternalLink, FiEye, FiEyeOff } from "react-icons/fi";
import { useEffect, useState } from "react";
import { ClusterTopology, NodeVersion, Node } from "@/types";
import { Link } from "@chakra-ui/react";
import { useFlowsContext } from "@/contexts/FlowsContext";

const Legend = () => {
  const { getTopology } = useApi();
  const [topology, setTopology] = useState<ClusterTopology | null>(null);
  const { flows, flowVisibility, setFlowVisibility } = useFlowsContext();

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
    getTopology().then(setTopology);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const sortedFlows = flows.sort((a, b) => {
    return a["flow-id"].localeCompare(b["flow-id"]);
  });

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
              <Th>Show/Hide</Th>
            </Tr>
          </Thead>
          <Tbody>
            {sortedFlows.map((flow) => {
              // TODO: Update these when this PR is merged:
              // https://github.com/kurtosis-tech/kardinal/pull/234
              const flowId = flow["flow-id"];
              const flowUrls = flow["access-entry"];
              const isBaseline = servicesForFlowId(flowId, false).length === 0;
              return (
                <Tr key={flowId}>
                  <Td>{flowId}</Td>
                  <Td textTransform={"capitalize"}>{isBaseline.toString()}</Td>
                  <Td whiteSpace={"nowrap"}>
                    {flowUrls[0]?.hostname != null ? (
                      <Link
                        href={`http://${flowUrls[0]}`}
                        isExternal
                        target="_blank"
                        display={"flex"}
                        alignItems={"center"}
                        gap={2}
                      >
                        {flowUrls[0].hostname}
                        <FiExternalLink size={12} />
                      </Link>
                    ) : (
                      "—"
                    )}
                  </Td>
                  <Td textAlign={"right"}>
                    {flowVisibility[flowId] === true ? (
                      <IconButton
                        aria-label="Hide"
                        h={"24px"}
                        minH={"24px"}
                        w={"24px"}
                        minW={"24px"}
                        color={isBaseline ? "gray.300" : "gray.600"}
                        icon={<FiEye size={16} />}
                        disabled={isBaseline}
                        onClick={() => {
                          if (isBaseline) return; // cant hide baseline flow
                          setFlowVisibility(flowId, false);
                        }}
                      />
                    ) : (
                      <IconButton
                        aria-label="Show"
                        h={"24px"}
                        minH={"24px"}
                        w={"24px"}
                        minW={"24px"}
                        disabled={isBaseline}
                        icon={<FiEyeOff size={16} />}
                        color={isBaseline ? "gray.300" : "gray.600"}
                        cursor={isBaseline ? "not-allowed" : "pointer"}
                        onClick={() => {
                          setFlowVisibility(flowId, true);
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
