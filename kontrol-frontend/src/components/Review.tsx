import { ReactNode } from "react";
import { Flex, Heading, Text } from "@chakra-ui/react";

const Review = ({ children }: { children?: ReactNode }) => {
  return (
    <Flex
      alignItems={"center"}
      justifyContent={"center"}
      height={"100%"}
      flexDir={"column"}
      gap={8}
    >
      <Heading>Review</Heading>
      <Text fontSize="2xl">Review the created flow here</Text>
      {children}
    </Flex>
  );
};

export default Review;
