import { VStack, Center, Text, Icon, Image } from "@chakra-ui/react";
import Button from "@/components/Button";
import { ElementType, ReactNode } from "react";
import { BiPlus } from "react-icons/bi";
import illustration from "./illustration.svg";
import bg from "./bg.svg";

interface Props {
  children: ReactNode;
  icon: ElementType;
  buttonText: string;
  buttonTo: string;
  buttonIcon?: null | ElementType;
}

const EmptyState = ({
  buttonIcon,
  buttonText,
  buttonTo,
  children,
  icon,
}: Props) => {
  return (
    <Center
      h="100%"
      backgroundImage={`url(${bg})`}
      backgroundPosition={"center center"}
      backgroundRepeat={"no-repeat"}
      backgroundSize={"cover"}
      bgColor={"white"}
    >
      <VStack
        bg="white"
        alignItems="center"
        gap={4}
        p={2}
        borderRadius="lg"
        textAlign="center"
        bgColor={"white"}
        maxW={394}
      >
        <VStack>
          <Image
            src={illustration}
            alt="gradient illustration"
            role="presentation"
            mb={"-24px"}
          />
          <Center
            borderRadius="100%"
            height={"52px"}
            width={"52px"}
            bgColor={"white"}
            filter={
              "drop-shadow(1px 1px 4px rgba(252, 160, 97, 0.10)) drop-shadow(4px 6px 7px rgba(252, 160, 97, 0.09)) drop-shadow(9px 12px 9px rgba(252, 160, 97, 0.05)) drop-shadow(17px 22px 11px rgba(252, 160, 97, 0.01)) drop-shadow(26px 35px 12px rgba(252, 160, 97, 0.00))"
            }
          >
            <Icon as={icon} size={24} boxSize={6} color="orange.500" />
          </Center>
        </VStack>
        <Text fontWeight={"normal"}>{children}</Text>
        <Button
          leftIcon={
            <Icon
              as={buttonIcon ?? BiPlus}
              size={24}
              boxSize={6}
              color="white"
            />
          }
          to={buttonTo}
        >
          {buttonText}
        </Button>
      </VStack>
    </Center>
  );
};

export default EmptyState;
