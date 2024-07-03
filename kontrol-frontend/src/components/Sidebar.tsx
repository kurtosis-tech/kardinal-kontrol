import { Link as ReactRouterLink } from "react-router-dom";
import { useMatch } from "react-router-dom";

import {
  Flex,
  Icon,
  Link,
  Text,
  VStack,
  HStack,
  IconButton,
  Spacer,
} from "@chakra-ui/react";
import { ElementType, ReactNode } from "react";
import {
  FiChevronLeft,
  FiDatabase,
  FiGitPullRequest,
  FiRepeat,
  FiShield,
  FiTable,
} from "react-icons/fi";

const SidebarLink = ({
  icon,
  href,
  children,
}: {
  icon: ElementType;
  href: string;
  children: ReactNode;
}) => {
  const match = useMatch({
    path: href,
    caseSensitive: true,
    end: true, // exact match
  });
  const color = match ? "gray.900" : "gray.500";
  return (
    <Link
      as={ReactRouterLink}
      to={href}
      _hover={{ textDecor: "none" }}
      w="full"
    >
      <HStack spacing="3">
        <Icon as={icon} boxSize="5" color={color} />
        <Text fontSize="sm" fontWeight="normal" color={color}>
          {children}
        </Text>
      </HStack>
    </Link>
  );
};

const Sidebar = () => {
  return (
    <Flex
      as="nav"
      direction="column"
      h="100%"
      w="250px"
      bg="white"
      boxShadow="md"
      p="5"
      borderRightWidth="1px"
      borderRightColor="gray.100"
    >
      <VStack align="start" spacing="4">
        <SidebarLink icon={FiTable} href="">
          Dashboard
        </SidebarLink>
        <SidebarLink icon={FiShield} href="maturity-gates">
          Maturity gates
        </SidebarLink>
        <SidebarLink icon={FiGitPullRequest} href="flows">
          Flows
        </SidebarLink>
        <SidebarLink icon={FiRepeat} href="traffic-configuration">
          Traffic configuration
        </SidebarLink>
        <SidebarLink icon={FiDatabase} href="data-configuration">
          Data configuration
        </SidebarLink>
      </VStack>
      <Spacer />

      <Flex mt="auto" pt="5" alignItems="center">
        <IconButton
          icon={<FiChevronLeft />}
          aria-label="Collapse sidebar"
          variant="ghost"
          fontSize="20px"
          color="gray.500"
        />
        <Text ml="2" fontSize="sm" color="gray.500" fontWeight={"normal"}>
          Collapse sidebar
        </Text>
      </Flex>
    </Flex>
  );
};

export default Sidebar;
