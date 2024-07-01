import { Outlet } from "react-router-dom";
import Navbar from "./components/Navbar";
import Sidebar from "./components/Sidebar";
import { Grid, GridItem } from "@chakra-ui/react";

const App = () => {
  return (
    <Grid
      templateAreas={`"nav nav"
                      "side main"`}
      gridTemplateRows={"64px 1fr"}
      gridTemplateColumns={"250px 1fr"}
      h="100vh"
      gap="1"
      color="blackAlpha.700"
      fontWeight="bold"
    >
      <GridItem area={"nav"}>
        <Navbar />
      </GridItem>
      <GridItem area={"side"}>
        <Sidebar />
      </GridItem>
      <GridItem area={"main"} bg="gray.50" p="5">
        <Outlet />
      </GridItem>
    </Grid>
  );
};

export default App;
