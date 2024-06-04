// 1. import `extendTheme` function
import { extendTheme, type ThemeConfig } from "@chakra-ui/react";

// 2. Add your color mode config
const config: ThemeConfig = {
  initialColorMode: "light",
  useSystemColorMode: false,
};

// 3. Create a custom color palette
const colors = {
  // gray: {
  //   50: "#929290",
  //   100: "#898987",
  //   200: "#7f7f7d",
  //   300: "#666664",
  //   400: "#5c5c59",
  //   500: "#434340",
  //   600: "#393937",
  //   700: "#20202d",
  //   800: "#161613",
  //   900: "#0d0d0b",
  // },
  gray: {
    "50": "#F2F2F2",
    "100": "#DBDBDB",
    "200": "#C4C4C4",
    "300": "#ADADAD",
    "400": "#969696",
    "500": "#808080",
    "600": "#666666",
    "700": "#4D4D4D",
    "800": "#333333",
    "900": "#1A1A1A",
  },
};

// 4. extend the theme
const theme = extendTheme({ config, colors });

export default theme;
