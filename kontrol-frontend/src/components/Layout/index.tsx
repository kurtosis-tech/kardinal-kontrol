import { GridItem, Grid } from "@chakra-ui/react";
import { ReactNode } from "react";

interface Props {
  children: ReactNode[];
}

const Layout = ({ children }: Props) => {
  return (
    <Grid
      mt={4}
      templateAreas={`
        "title title"
        "a a"
        "b b"
        "c d"
        "footer footer"
      `}
      gridTemplateRows={"64px auto auto auto 40px"}
      gridTemplateColumns={"1fr 1fr"}
      h="100%"
      rowGap={8}
      columnGap={4}
      color="blackAlpha.700"
      fontWeight="bold"
    >
      {["title", "a", "b", "c", "d", "footer"].map((area, i) => (
        <GridItem area={area} display="flex" flexDirection={"column"}>
          {children[i]}
        </GridItem>
      ))}
    </Grid>
  );
};

export default Layout;
