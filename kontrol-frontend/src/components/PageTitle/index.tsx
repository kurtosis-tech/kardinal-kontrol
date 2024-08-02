import { Spacer, Stack, Text } from "@chakra-ui/react";
import { ReactNode } from "react";
import Button from "@/components/Button";
import { FiPlus } from "react-icons/fi";

interface Props {
  title: string;
  children: ReactNode;
  buttonText?: string;
  onClick?: () => void;
}

const PageTitle = ({ title, children, buttonText, onClick }: Props) => {
  return (
    <Stack direction={"row"} alignItems={"center"}>
      <Stack gap={0}>
        <Text
          as="h1"
          fontSize="3xl"
          fontWeight={500}
          mb={4}
          color={"gray.800"}
          m={0}
        >
          {title}
        </Text>
        <Text fontSize="lg" color="gray.500" fontWeight={400}>
          {children}
        </Text>
      </Stack>
      <Spacer />
      {buttonText != null && (
        <Button onClick={onClick} gap={2}>
          <FiPlus /> {buttonText}
        </Button>
      )}
    </Stack>
  );
};

export default PageTitle;
