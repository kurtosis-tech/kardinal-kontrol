// WIP
import {
  Box,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Flex,
  Text,
  Link,
  IconButton,
} from "@chakra-ui/react";
import { FaGithub, FaAmazon, FaStripeS } from "react-icons/fa";
import {
  FiGitBranch,
  FiPlay,
  FiEdit,
  FiTrash,
  FiGrid,
  FiColumns,
  FiMoreHorizontal,
} from "react-icons/fi";
import Chip from "@/components/Chip";

const FlowConfigurationTable = () => {
  return (
    <Box
      borderWidth="1px"
      borderRadius="lg"
      borderColor="gray.100"
      overflow="hidden"
      background={"white"}
    >
      <Flex
        justifyContent="space-between"
        alignItems="center"
        p={4}
        borderBottomWidth="1px"
        borderColor="gray.100"
      >
        <Text fontSize="md" fontWeight="semibold" ml={2}>
          Flow configuration
        </Text>
        <Flex justifyContent={"flex-end"} gap={1}>
          <IconButton
            icon={<FiColumns />}
            aria-label="List view"
            variant="ghost"
          />
          <IconButton
            icon={<FiGrid />}
            aria-label="Grid view"
            variant="ghost"
          />
          <IconButton
            icon={<FiMoreHorizontal />}
            aria-label="More options"
            variant="ghost"
          />
        </Flex>
      </Flex>
      <Table variant="simple">
        <Thead background={"gray.50"}>
          <Tr>
            <Th textTransform={"none"} fontSize={"sm"} fontWeight={500}>
              Flow template
            </Th>
            <Th textTransform={"none"} fontSize={"sm"} fontWeight={500}>
              Service
            </Th>
            <Th textTransform={"none"} fontSize={"sm"} fontWeight={500}>
              Data layer
            </Th>
            <Th
              textTransform={"none"}
              fontSize={"sm"}
              fontWeight={500}
              textAlign={"right"}
            >
              Actions
            </Th>
          </Tr>
        </Thead>
        <Tbody>
          <Tr>
            <Td>
              <Flex alignItems="center" gap={2}>
                <Box as={FiGitBranch} />
                <Link textDecor={"underline"}>Early development flow</Link>
              </Flex>
            </Td>
            <Td>
              <Flex alignItems="center" gap={2}>
                <Box as={FaGithub} />
                <Text textDecor={"underline"}>awesome-kardinal-example</Text>
              </Flex>
            </Td>
            <Td>
              <Flex alignItems="center" gap={1}>
                <Chip icon={<FaStripeS />}>Stripe</Chip>
                <Chip icon={<FaAmazon />}>Amazon RDS</Chip>
              </Flex>
            </Td>
            <Td pr={4}>
              <Flex gap={1} justifyContent={"flex-end"}>
                <IconButton
                  icon={<FiPlay />}
                  aria-label="Play"
                  variant="ghost"
                />
                <IconButton
                  icon={<FiEdit />}
                  aria-label="Edit"
                  variant="ghost"
                />
                <IconButton
                  icon={<FiTrash />}
                  color={"red"}
                  aria-label="Delete"
                  variant="ghost"
                />
              </Flex>
            </Td>
          </Tr>
        </Tbody>
      </Table>
    </Box>
  );
};

export default FlowConfigurationTable;
