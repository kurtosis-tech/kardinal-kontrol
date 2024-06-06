// 1. import `extendTheme` function
import { extendTheme, type ThemeConfig } from "@chakra-ui/react";

// 2. Add your color mode config
const config: ThemeConfig = {
  initialColorMode: "light",
  useSystemColorMode: false,
};

// 3. Create a custom color palette
const colors = {
  gray: {
    "50": "#f2f0f0",
    "100": "#dad9d7",
    "200": "#c0bfbf",
    "300": "#a7a6a6",
    "400": "#8d8d8d",
    "500": "#747373",
    "600": "#5A5A59",
    "700": "#41403F",
    "800": "#282625",
    "900": "#1B1A19",
  },
};

// 4. extend the theme
const theme = extendTheme({ config, colors });

export default theme;
