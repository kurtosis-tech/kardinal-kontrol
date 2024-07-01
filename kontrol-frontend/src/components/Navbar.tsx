import { Box, Flex, Avatar, Text, Link, Image, Spacer } from "@chakra-ui/react";

const Navbar = () => {
  return (
    <Box
      bg="white"
      px={4}
      boxShadow="sm"
      as="nav"
      height={"64px"}
      borderBottomWidth="1px"
      borderBottomColor="gray.100"
    >
      <Flex h={16} alignItems="center">
        {/* Logo */}
        <Link href="/" display="flex" alignItems="center">
          <Image
            src="/logo.png"
            alt="Kardinal Logo"
            borderRadius="full"
            boxSize="40px"
          />
          <Text ml={2} fontSize="xl">
            Kardinal
          </Text>
        </Link>

        <Spacer />

        {/* User Profile */}
        <Flex alignItems="center">
          <Avatar
            size="sm"
            name="Charlie Brown"
            src="https://via.placeholder.com/150"
          />
          <Box ml={2}>
            <Text>Charlie Brown</Text>
            <Text fontSize="sm" color="gray.500">
              ACME Corporation
            </Text>
          </Box>
        </Flex>
      </Flex>
    </Box>
  );
};

export default Navbar;
