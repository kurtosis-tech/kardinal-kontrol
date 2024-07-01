import { Flex, Center, Text, Icon } from "@chakra-ui/react";
import Button from "@/components/Button";
import { ElementType, ReactNode } from "react";

const ComingSoon = ({
  title,
  children,
  icon,
}: {
  title: string;
  children: ReactNode;
  icon: ElementType;
}) => {
  return (
    <Center h="100vh">
      <Flex
        bg="white"
        py={12}
        px={8}
        flexDirection="column"
        alignItems="center"
        gap={2}
        borderRadius="lg"
        boxShadow="lg"
        textAlign="center"
        maxW={502}
      >
        <Icon as={icon} size={48} boxSize={8} color="gray.500" />
        <Text fontSize="md" fontWeight="bold" color={"gray.900"}>
          {title}
        </Text>
        <Text fontWeight={"normal"}>{children}</Text>
        <Button>Go back</Button>
      </Flex>
    </Center>
  );
};

export default ComingSoon;
