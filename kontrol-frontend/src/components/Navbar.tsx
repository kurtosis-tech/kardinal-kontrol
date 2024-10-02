import {
  Box,
  Flex,
  Avatar,
  Text,
  Link,
  Image,
  Spacer,
  AvatarBadge,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  Button,
} from "@chakra-ui/react";
import { useNavigate } from "react-router-dom";

const Navbar = () => {
  const navigate = useNavigate();

  const navigateToPlaceholder = () => {
    navigate("profile");
  };

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

        <Flex alignItems="center">
          <Menu>
            <MenuButton as={Button} p={0} bg={"white"}>
              <Avatar size="sm">
                <AvatarBadge boxSize="1.25em" bg="green.500" />
              </Avatar>
            </MenuButton>
            <MenuList>
              <MenuItem onClick={navigateToPlaceholder}>Profile</MenuItem>
              <MenuItem onClick={navigateToPlaceholder}>Settings</MenuItem>
              <MenuItem onClick={navigateToPlaceholder}>Log in</MenuItem>
            </MenuList>
          </Menu>
        </Flex>
      </Flex>
    </Box>
  );
};

export default Navbar;
