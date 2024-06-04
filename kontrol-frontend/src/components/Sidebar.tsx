import { Box, Link, VStack } from "@chakra-ui/react";

const Sidebar = () => {
  return (
    <Box
      height="100vh"
      backgroundColor="gray.50"
      padding="20px"
      boxShadow="2xl"
    >
      <VStack spacing="10px" align="flex-start">
        <Link href="#dashboard" _hover={{ textDecoration: "underline" }}>
          Dashboard
        </Link>
        <Link href="#settings" _hover={{ textDecoration: "underline" }}>
          Settings
        </Link>
        <Link href="#profile" _hover={{ textDecoration: "underline" }}>
          Profile
        </Link>
      </VStack>
    </Box>
  );
};

export default Sidebar;
