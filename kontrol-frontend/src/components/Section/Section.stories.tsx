import type { Meta, StoryObj } from "@storybook/react";

import Section from ".";
import { Box } from "@chakra-ui/react";

const meta: Meta<typeof Section> = {
  component: Section,
};

export default meta;
type Story = StoryObj<typeof Section>;

export const Example: Story = {
  args: {
    title: "Maturity gate",
    children: <Box height={24} />,
  },
};
