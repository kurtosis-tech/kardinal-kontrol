import { Center, Text } from "@chakra-ui/react";
import VisxDemo from "./components/VisxDemo";

function App() {
  return (
    <Center p={24} height={"100vh"} flexDir={"column"}>
      <Center width={800} height={800}>
        <VisxDemo width={800} height={600} />
      </Center>
      <Text fontSize={"3xl"}>"Stonks only go up"</Text>
      <Text fontSize={"3xl"}>- Tim Apple</Text>
    </Center>
  );
}

export default App;
