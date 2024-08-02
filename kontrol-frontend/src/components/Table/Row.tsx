import {
  Box,
  Tr,
  Td,
  Flex,
  Text,
  Link,
  IconButton,
  Stack,
} from "@chakra-ui/react";
import { FaGithub, FaStripeS, FaAmazon } from "react-icons/fa";
import {
  FiGitBranch,
  FiPlay,
  FiEdit,
  FiTrash,
  FiChevronRight,
  FiChevronDown,
} from "react-icons/fi";
import Chip from "../Chip";
import { useFlowsContext } from "@/contexts/FlowsContext";

const TableCell = ({
  children,
  isExpanded,
  expandedContent,
  pl,
}: {
  children: React.ReactNode;
  isExpanded: boolean;
  expandedContent: React.ReactNode;
  pl?: number;
}) => {
  return (
    <Td pl={pl} verticalAlign={"baseline"}>
      <Stack gap={3} justifyContent={"center"}>
        <Flex alignItems="center" gap={2}>
          {children}
        </Flex>
        {isExpanded && expandedContent}
      </Stack>
    </Td>
  );
};

interface Props {
  id: string;
  isExpanded: boolean;
  onExpandRow: (rowIndex: string | null) => void;
}

const Row = ({ id, isExpanded, onExpandRow }: Props) => {
  const { deleteExampleFlow } = useFlowsContext();
  return (
    <Tr>
      <Td width={12} p={0} verticalAlign={"baseline"}>
        <Flex gap={0} justifyContent={"center"}>
          <IconButton
            size="sm"
            width={8}
            onClick={() => onExpandRow(isExpanded ? null : id)}
            icon={isExpanded ? <FiChevronDown /> : <FiChevronRight />}
            aria-label="Expand"
            variant="ghost"
          />
        </Flex>
      </Td>
      <TableCell
        isExpanded={isExpanded}
        pl={0}
        expandedContent={
          <>
            <Text fontWeight={500} fontSize={"sm"} color={"gray.500"}>
              Base image tag
            </Text>
            <Text fontSize={"sm"} color={"gray.500"}>
              RC-1.0.0
            </Text>
          </>
        }
      >
        <Box as={FiGitBranch} />
        <Link textDecor={"underline"}>Early development flow</Link>
      </TableCell>
      <TableCell
        isExpanded={isExpanded}
        expandedContent={
          <>
            <Text fontWeight={500} fontSize={"sm"} color={"gray.500"}>
              Status
            </Text>
            <Text fontSize={"sm"} color={"gray.500"}>
              Running
            </Text>
          </>
        }
      >
        <Box as={FaGithub} />
        <Text textDecor={"underline"}>awesome-kardinal-example</Text>
      </TableCell>
      <Td verticalAlign={"baseline"}>
        <Flex alignItems="flex-start" gap={1} height={"100%"}>
          <Chip icon={FaStripeS} colorScheme="purple">
            Stripe
          </Chip>
          <Chip icon={FaAmazon} colorScheme="blue">
            Amazon RDS
          </Chip>
        </Flex>
      </Td>
      <Td pr={4} py={0} verticalAlign={"baseline"}>
        <Flex gap={1} justifyContent={"flex-end"}>
          <IconButton icon={<FiPlay />} aria-label="Play" variant="ghost" />
          <IconButton icon={<FiEdit />} aria-label="Edit" variant="ghost" />
          <IconButton
            icon={<FiTrash />}
            color={"red"}
            aria-label="Delete"
            variant="ghost"
            onClick={() => deleteExampleFlow()}
          />
        </Flex>
      </Td>
    </Tr>
  );
};

export default Row;
