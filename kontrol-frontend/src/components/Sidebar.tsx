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
  return (
    <Link href={href} _hover={{ textDecor: "none" }} w="full">
      <HStack spacing="3">
        <Icon as={icon} boxSize="6" />
        <Text fontSize="lg" fontWeight="bold">
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
    >
      <VStack align="start" spacing="4">
        <SidebarLink icon={FiTable} href="#">
          Dashboard
        </SidebarLink>
        <SidebarLink icon={FiShield} href="#">
          Maturity gates
        </SidebarLink>
        <SidebarLink icon={FiGitPullRequest} href="#">
          Flows
        </SidebarLink>
        <SidebarLink icon={FiRepeat} href="#">
          Traffic configuration
        </SidebarLink>
        <SidebarLink icon={FiDatabase} href="#">
          Data configuration
        </SidebarLink>
      </VStack>
      <Spacer />

      <Flex mt="auto" pt="5">
        <IconButton
          icon={<FiChevronLeft />}
          aria-label="Collapse sidebar"
          variant="ghost"
          fontSize="20px"
        />
        <Text ml="2" fontSize="lg" color="gray.500">
          Collapse sidebar
        </Text>
      </Flex>
    </Flex>
  );
};

export default Sidebar;
