import { Outlet } from "react-router-dom";
import Navbar from "@/components/Navbar";
import Sidebar from "@/components/Sidebar";
import { Grid, GridItem } from "@chakra-ui/react";
import { useNavigationContext } from "@/contexts/NavigationContext";

const App = () => {
  const { isSidebarCollapsed } = useNavigationContext();
  return (
    <Grid
      templateAreas={`"nav nav"
                      "side main"`}
      gridTemplateRows={"64px 1fr"}
      gridTemplateColumns={`${isSidebarCollapsed ? "54px" : "250px"} 1fr`}
      color="blackAlpha.700"
      h="100%"
    >
      <GridItem area={"nav"}>
        <Navbar />
      </GridItem>
      <GridItem area={"side"} display={"flex"}>
        <Sidebar />
      </GridItem>
      <GridItem area={"main"} bg={"#FAFAFA"} p={8} pb={16}>
        <Outlet />
      </GridItem>
    </Grid>
  );
};

export default App;
