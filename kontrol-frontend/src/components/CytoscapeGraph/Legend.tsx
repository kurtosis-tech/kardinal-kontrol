import { NodeVersion } from "@/types";
import { Flex } from "@chakra-ui/react";
import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableCaption,
  TableContainer,
} from "@chakra-ui/react";

interface Props {
  elements: cytoscape.ElementDefinition[];
}

const Legend = ({ elements }: Props) => {
  const serviceVersions: NodeVersion[] = elements
    .map((element) => element.data.versions)
    .flat()
    .filter(Boolean);

  const flowIds = serviceVersions.map((version) => version.flowId);
  const uniqueFlowIds = [...new Set(flowIds)];

  const servicesForFlowId = (flowId: string): string => {
    const services = elements
      .map((element) => {
        const versions = element.data.versions;
        if (
          versions != null &&
          versions.length > 0 &&
          versions.some((version: NodeVersion) => version.flowId === flowId)
        ) {
          return element;
        }
        return undefined;
      })
      .filter(Boolean);
    // If flow ID is baseline flow, dont show all services
    if (
      services.some(
        (service) =>
          service?.data.versions.length === 1 &&
          service?.data.versions[0].isBaseline,
      )
    ) {
      return "—";
    }
    return services.map((service) => service?.data.label).join(", ");
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
      borderWidth={1}
      borderStyle={"solid"}
      borderColor={"gray.300"}
      padding={4}
    >
      <TableContainer>
        <Table variant="simple">
          <TableCaption>Legend</TableCaption>
          <Thead>
            <Tr>
              <Th>Flow ID</Th>
              <Th>Deployed Services</Th>
            </Tr>
          </Thead>
          <Tbody>
            {uniqueFlowIds.map((flowId) => {
              return (
                <Tr>
                  <Td>{flowId}</Td>
                  <Td>{servicesForFlowId(flowId)}</Td>
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
