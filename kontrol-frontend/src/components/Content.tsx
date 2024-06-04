import { PropsWithChildren } from "react";
import { Box } from "@chakra-ui/react";

const Content = ({ children }: PropsWithChildren) => {
  return (
    <Box flex="1" padding="20px">
      {children}
    </Box>
  );
};

export default Content;
