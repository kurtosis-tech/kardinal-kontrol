import { Flex, Heading, Text } from "@chakra-ui/react";
import { ReactNode } from "react";

const Manage = ({ children }: { children?: ReactNode }) => {
  return (
    <Flex
      alignItems={"center"}
      justifyContent={"center"}
      height={"100%"}
      flexDir={"column"}
      gap={8}
    >
      <Heading>Manage</Heading>
      <Text fontSize="2xl">Deploy, manage, delete the created flow</Text>
      {children}
    </Flex>
  );
};

export default Manage;
