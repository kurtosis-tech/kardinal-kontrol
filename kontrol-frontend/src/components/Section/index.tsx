import { Box, Flex, Text } from "@chakra-ui/react";
import { ReactNode } from "react";

interface Props {
  title?: string;
  children?: ReactNode;
}
const Section = ({ title, children }: Props) => {
  return (
    <>
      <Text
        fontSize="sm"
        fontWeight="bold"
        mb={4}
        as={"h3"}
        textTransform={"uppercase"}
      >
        {title}
      </Text>

      <Box
        borderWidth="1px"
        borderRadius="12px"
        p={4}
        background={"white"}
        borderColor={"gray.100"}
      >
        <Flex direction={{ base: "column", md: "row" }} gap={4}>
          {children}
        </Flex>
      </Box>
    </>
  );
};

export default Section;
