import type { Meta, StoryObj } from "@storybook/react";
import { FiGitMerge } from "react-icons/fi";

import ComingSoon from ".";

const meta: Meta<typeof ComingSoon> = {
  component: ComingSoon,
};

export default meta;
type Story = StoryObj<typeof ComingSoon>;

export const Example: Story = {
  args: {
    icon: FiGitMerge,
    title: "Flows are coming soon",
    children:
      "We’re working on getting this functionality up and running. We’ll let you know when it’s ready for you!",
  },
};
