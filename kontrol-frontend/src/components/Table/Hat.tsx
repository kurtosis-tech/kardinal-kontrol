import { FiGrid, FiColumns, FiMoreHorizontal } from "react-icons/fi";
import { ReactNode } from "react";
import { Flex, Text, IconButton } from "@chakra-ui/react";

const Hat = ({ children }: { children?: ReactNode }) => {
  return (
    <Flex
      justifyContent="space-between"
      alignItems="center"
      p={4}
      pl={10}
      borderBottomWidth="1px"
      borderColor="gray.200"
    >
      <Text fontSize="md" fontWeight="semibold" ml={2}>
        {children}
      </Text>
      <Flex justifyContent={"flex-end"} gap={1}>
        <IconButton
          icon={<FiColumns />}
          aria-label="List view"
          variant="ghost"
        />
        <IconButton icon={<FiGrid />} aria-label="Grid view" variant="ghost" />
        <IconButton
          icon={<FiMoreHorizontal />}
          aria-label="More options"
          variant="ghost"
        />
      </Flex>
    </Flex>
  );
};

export default Hat;
