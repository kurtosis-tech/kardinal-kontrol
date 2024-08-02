import type { Meta, StoryObj } from "@storybook/react";

import EmptyState from ".";
import { FiShield } from "react-icons/fi";

const meta: Meta<typeof EmptyState> = {
  component: EmptyState,
};

export default meta;
type Story = StoryObj<typeof EmptyState>;

export const Example: Story = {
  args: {
    icon: FiShield,
    buttonText: "Create New Flow Configuration",
    children: "No active flow configurations yet",
  },
};
