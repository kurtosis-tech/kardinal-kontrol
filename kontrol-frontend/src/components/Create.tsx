import { Flex, Heading, Text } from "@chakra-ui/react";
import { ReactNode } from "react";

const Create = ({ children }: { children?: ReactNode }) => {
  return (
    <Flex
      alignItems={"center"}
      justifyContent={"center"}
      height={"100%"}
      flexDir={"column"}
      gap={8}
    >
      <Heading>Create</Heading>
      <Text fontSize="2xl">Create a new flow</Text>
      {children}
    </Flex>
  );
};

export default Create;
