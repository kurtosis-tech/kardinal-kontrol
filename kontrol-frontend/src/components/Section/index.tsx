import { Box, Text } from "@chakra-ui/react";
import { ReactNode } from "react";

interface Props {
  title?: string;
  children?: ReactNode;
  offsetTop?: number;
}
const Section = ({ title, children, offsetTop }: Props) => {
  return (
    <>
      {title != null ? (
        <Text
          fontSize="sm"
          fontWeight="bold"
          mb={4}
          as={"h3"}
          textTransform={"uppercase"}
        >
          {title}
        </Text>
      ) : offsetTop ? (
        <Box height={offsetTop} />
      ) : null}

      <Box
        borderWidth="1px"
        borderRadius="12px"
        p={4}
        background={"white"}
        borderColor={"gray.200"}
        alignSelf={"stretch"}
        width={"100%"}
        flex={1}
        maxH={"520px"}
      >
        {children}
      </Box>
    </>
  );
};

export default Section;
