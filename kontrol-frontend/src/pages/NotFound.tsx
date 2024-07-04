import { Box, Heading, Text } from "@chakra-ui/react";

const NotFound = () => {
  return (
    <Box textAlign="center" py={10} px={6}>
      <Heading display="inline-block" as="h2" size="2xl">
        Oops!
      </Heading>
      <Text fontSize="18px" mt={3} mb={2}>
        That is a broken link
      </Text>
      <Text color={"gray.500"} mb={6}>
        Ensure the path you are accessing is valid and that the path contains
        your cluster UUID, for example:
        <br /> <code>/eceb88f5-0d83-400e-ae45-3273347f2c24/flows</code>
      </Text>
    </Box>
  );
};

export default NotFound;
