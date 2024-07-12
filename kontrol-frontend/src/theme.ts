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
    "50": "#FAFBFC",
    "100": "#EAEBF0",
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
const theme = extendTheme({
  config,
  colors,
  fonts: {
    body: "'DM Sans', 'sans-serif'",
  },
  styles: {
    global: {
      ":root": {
        "--color-text": "#68727d",
        "--font-family": "'DM Sans', sans-serif",
      },
      "*": {
        "font-synthesis": "none",
        "text-rendering": "optimizeLegibility",
        "-webkit-font-smoothing": "antialiased",
        "-moz-osx-font-smoothing": "grayscale",
      },
      html: {
        "color-scheme": "light",
        "background-color": "#f9f9f9",
        "min-height": "100vh",
        "font-size": "16px",
      },
      "#root": {
        height: "100vh",
      },
      body: {
        fontSize: "14px", // Set your desired default font size
        color: "gray.800",
        bg: "white",
        fontWeight: 400,
      },
    },
  },
});

export default theme;
