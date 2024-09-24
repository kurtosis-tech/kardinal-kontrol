import { extendTheme, type ThemeConfig } from "@chakra-ui/react";
import { MultiSelectTheme } from "chakra-multiselect";

const config: ThemeConfig = {
  initialColorMode: "light",
  useSystemColorMode: false,
};

export const colorOverrides: Record<string, Record<string, string>> = {
  gray: {
    "50": "#F9FAFB",
    "100": "#F3F4F6",
    "200": "#E5E7EB",
    "300": "#D1D5DB",
    "400": "#9CA3AF",
    "500": "#6B7280",
    "600": "#4B5563",
    "700": "#374151",
    "800": "#1F2937",
    "900": "#111827",
    "950": "#030712",
  },
  orange: {
    "500": "#FF602C",
    "100": "#FCA0610D",
  },
  blue: {
    "50": "#F6FAFF",
    "300": "#58A6FF",
    "500": "#2170CB",
  },
  purple: {
    "50": "#F7F7FF",
    "300": "#B5B1FF",
    "500": "#635BFF",
  },
};

// extend the theme
const theme = extendTheme({
  config,
  colors: colorOverrides,
  components: {
    MultiSelect: {
      ...MultiSelectTheme,
      baseStyle: (props: Record<string, unknown>) => {
        const baseStyles = MultiSelectTheme.baseStyle(props);
        return {
          ...baseStyles,
          input: {
            ...baseStyles.input,
            flexGrow: 1,
          },
          defaultProps: {
            size: "lg",
            w: "100%",
          },
          list: {
            ...baseStyles.list,
            borderRadius: "12px",
          },
          control: {
            ...baseStyles.control,
            borderRadius: "12px",
            height: "48px",
            flexGrow: 0,
          },
          actionGroup: {
            ...baseStyles.actionGroup,
            flexGrow: 0,
          },
        };
      },
    },
  },
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
        height: "100vh",
        "font-size": "16px",
      },
      "#root": {
        height: "100vh",
      },
      body: {
        fontSize: "14px", // Set your desired default font size
        color: "gray.500",
        bg: "white",
        fontWeight: 400,
        height: "100vh",
      },
    },
  },
});

export default theme;
