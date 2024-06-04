import { Box, Flex, Center, Text } from "@chakra-ui/react";
import VisxDemo from "./components/VisxDemo";
import Sidebar from "./components/Sidebar";
import Content from "./components/Content";

function App() {
  return (
    <Flex display="grid" gridTemplateColumns={"220px 1fr"}>
      <Sidebar />
      <Box flex="1" display="flex" flexDirection="column">
        <Content>
          <Center p={24} flexDir={"column"}>
            <Center width={800} height={800}>
              <VisxDemo width={800} height={600} />
            </Center>
            <Text fontSize={"3xl"}>"Stonks only go up"</Text>
            <Text fontSize={"3xl"}>- Tim Apple</Text>
          </Center>
        </Content>
      </Box>
    </Flex>
  );
}

export default App;
