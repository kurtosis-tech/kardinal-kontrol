import type { Preview } from "@storybook/react";
import theme from "../src/theme";

const preview: Preview = {
  parameters: {
    chakra: {
      theme,
    },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
};

export default preview;
