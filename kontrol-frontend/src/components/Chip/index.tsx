import { Icon, Tag, Text } from "@chakra-ui/react";
import { ElementType, ReactNode } from "react";

type ColorScheme = "blue" | "purple";

interface Props {
  children: ReactNode;
  icon: ElementType;
  colorScheme: ColorScheme;
}

const Chip = ({ children, icon, colorScheme }: Props) => {
  return (
    <Tag
      backgroundColor={`${colorScheme}.50`}
      borderRadius={"16px"}
      border={"1px solid"}
      borderColor={`${colorScheme}.300`}
      height={"32px"}
      px={3}
      py={4}
      gap={1}
    >
      <Icon
        as={icon}
        size={"16px"}
        boxSize={"16px"}
        color={`${colorScheme}.500`}
      />
      <Text
        as="span"
        size={"12px"}
        fontWeight={600}
        letterSpacing={"0.96px"}
        textTransform={"uppercase"}
        color={"gray.700"}
      >
        {children}
      </Text>
    </Tag>
  );
};
export default Chip;
