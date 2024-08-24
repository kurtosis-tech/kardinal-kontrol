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
import { FaGithub } from "react-icons/fa";
import {
  FiGitBranch,
  FiEdit,
  FiTrash,
  FiChevronRight,
  FiChevronDown,
  FiPlay,
} from "react-icons/fi";
import { Template } from "@/types";
import DeleteTemplateModal from "@/components/DeleteTemplateModal";
import CreateFlowModal from "../CreateFlowModal";

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
  template: Template;
  id: string;
  isExpanded: boolean;
  onExpandRow: (rowIndex: string | null) => void;
  onDelete: (templateId: string) => void;
}

const Row = ({ template, id, isExpanded, onExpandRow, onDelete }: Props) => {
  const templateId = template["template-id"];
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
              Services
            </Text>
            <Text fontSize={"sm"} color={"gray.500"}>
              TODO
            </Text>
          </>
        }
      >
        <Box as={FiGitBranch} />
        <Link textDecor={"underline"}>{template.name}</Link>
      </TableCell>
      <TableCell
        isExpanded={isExpanded}
        expandedContent={
          <>
            <Text fontWeight={500} fontSize={"sm"} color={"gray.500"}>
              Status
            </Text>
            <Text fontSize={"sm"} color={"gray.500"}>
              Not running
            </Text>
          </>
        }
      >
        <Box as={FaGithub} />
        <Text textDecor={"underline"}>{templateId}</Text>
      </TableCell>
      <Td pr={4} py={0} verticalAlign={"baseline"}>
        <Flex gap={1} justifyContent={"flex-end"}>
          <CreateFlowModal
            onConfirm={() => onDelete(templateId)}
            templateId={templateId}
          >
            <IconButton icon={<FiPlay />} aria-label="Play" variant="ghost" />
          </CreateFlowModal>
          <IconButton icon={<FiEdit />} aria-label="Edit" variant="ghost" />
          <DeleteTemplateModal
            onConfirm={() => onDelete(templateId)}
            templateId={templateId}
          >
            <IconButton
              icon={<FiTrash />}
              color={"red"}
              aria-label="Delete"
              variant="ghost"
            />
          </DeleteTemplateModal>
        </Flex>
      </Td>
    </Tr>
  );
};

export default Row;
