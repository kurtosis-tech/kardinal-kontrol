import type { Meta, StoryObj } from "@storybook/react";
import { Box } from "@chakra-ui/react";

import Layout from ".";

const meta: Meta<typeof Layout> = {
  component: Layout,
};

export default meta;
type Story = StoryObj<typeof Layout>;

export const Example: Story = {
  args: {
    children: [
      <Box backgroundColor={"gray.200"} flex={1}>
        Section A
      </Box>,
      <Box backgroundColor={"gray.200"} flex={1}>
        Section B
      </Box>,
      <Box backgroundColor={"gray.200"} flex={1}>
        Section C
      </Box>,
      <Box backgroundColor={"gray.200"} flex={1}>
        Section D
      </Box>,
    ],
  },
};
