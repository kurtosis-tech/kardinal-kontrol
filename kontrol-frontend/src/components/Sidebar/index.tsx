import { Link as ReactRouterLink } from "react-router-dom";
import { useMatch } from "react-router-dom";
import { useNavigationContext } from "@/contexts/NavigationContext";

import {
  Flex,
  Icon,
  Link,
  Text,
  VStack,
  HStack,
  Button,
  IconButton,
  Spacer,
} from "@chakra-ui/react";
import { ElementType } from "react";
import {
  FiChevronLeft,
  FiChevronRight,
  FiDatabase,
  FiGitPullRequest,
  FiRepeat,
  FiShield,
  FiTable,
} from "react-icons/fi";

interface SidebarLinkProps {
  icon: ElementType;
  href: string;
  children: string;
  isCollapsed: boolean;
}

const SidebarLink = ({
  icon,
  href,
  children,
  isCollapsed,
}: SidebarLinkProps) => {
  const match = useMatch({
    path: href,
    caseSensitive: true,
    end: true, // exact match
  });
  const color = match ? "gray.900" : "gray.500";
  return isCollapsed ? (
    <IconButton
      variant={"ghost"}
      aria-label={children}
      as={ReactRouterLink}
      icon={<Icon as={icon} boxSize="5" color={color} />}
    />
  ) : (
    <Link
      as={ReactRouterLink}
      to={href}
      _hover={{ textDecor: "none" }}
      w="full"
      mb={"3px"}
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
  const { isSidebarCollapsed, setIsSidebarCollapsed } = useNavigationContext();
  return (
    <VStack
      as="nav"
      w="250px"
      bg="white"
      boxShadow="md"
      p={isSidebarCollapsed ? "6px" : 4}
      flex={1}
      alignItems={"flex-start"}
      borderRightWidth="1px"
      borderRightColor="gray.100"
    >
      <VStack
        align="start"
        spacing={isSidebarCollapsed ? 0 : 4}
        pt={isSidebarCollapsed ? "1px" : 0}
      >
        <SidebarLink icon={FiTable} href="" isCollapsed={isSidebarCollapsed}>
          Dashboard
        </SidebarLink>
        <SidebarLink
          icon={FiShield}
          href="maturity-gates"
          isCollapsed={isSidebarCollapsed}
        >
          Maturity gates
        </SidebarLink>
        <SidebarLink
          icon={FiGitPullRequest}
          href="flows"
          isCollapsed={isSidebarCollapsed}
        >
          Flows
        </SidebarLink>
        <SidebarLink
          icon={FiRepeat}
          href="traffic-configuration"
          isCollapsed={isSidebarCollapsed}
        >
          Traffic configuration
        </SidebarLink>
        <SidebarLink
          icon={FiDatabase}
          href="data-configuration"
          isCollapsed={isSidebarCollapsed}
        >
          Data configuration
        </SidebarLink>
      </VStack>
      <Spacer />

      <Flex mt="auto" pt="5" alignItems="center">
        {isSidebarCollapsed ? (
          <IconButton
            icon={<FiChevronRight size={20} />}
            color="gray.500"
            variant={"ghost"}
            aria-label="Expand sidebar"
            onClick={() => setIsSidebarCollapsed(false)}
          />
        ) : (
          <Button
            aria-label="Collapse sidebar"
            variant="ghost"
            color="gray.500"
            fontWeight={400}
            gap={2}
            pl={0}
            onClick={() => setIsSidebarCollapsed(true)}
          >
            <FiChevronLeft />
            Collapse sidebar
          </Button>
        )}
      </Flex>
    </VStack>
  );
};

export default Sidebar;
