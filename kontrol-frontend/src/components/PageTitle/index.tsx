import { Stack, Text } from "@chakra-ui/react";
import { ReactNode } from "react";

interface Props {
  title: string;
  children: ReactNode;
}
const PageTitle = ({ title, children }: Props) => {
  return (
    <Stack>
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
      <Text fontSize="lg" color="gray.500">
        {children}
      </Text>
    </Stack>
  );
};

export default PageTitle;
