import { Tag, Text } from "@chakra-ui/react";
import { ReactNode } from "react";

interface Props {
  children: ReactNode;
  icon: ReactNode;
}

const Chip = ({ children, icon }: Props) => {
  return (
    <Tag
      color={"gray.500"}
      backgroundColor={"gray.50"}
      borderRadius={"16px"}
      height={"32px"}
      px={3}
      py={4}
      gap={1}
    >
      {icon}
      <Text as="span" size={"sm"}>
        {children}
      </Text>
    </Tag>
  );
};
export default Chip;
