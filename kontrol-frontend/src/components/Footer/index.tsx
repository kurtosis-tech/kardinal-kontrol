import { Flex } from "@chakra-ui/react";
import Button from "@/components/Button";
import { Link } from "react-router-dom";
import { useFlowsContext } from "@/contexts/FlowsContext";

const Footer = () => {
  const { createExampleFlow } = useFlowsContext();
  return (
    <Flex justifyContent={"flex-end"}>
      <Button
        as={Link}
        to={"../"}
        onClick={() => {
          createExampleFlow();
        }}
      >
        Create flow configuration
      </Button>
    </Flex>
  );
};

export default Footer;
