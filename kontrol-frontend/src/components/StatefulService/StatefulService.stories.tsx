import type { Meta, StoryObj } from "@storybook/react";

import StatefulService from ".";

const meta: Meta<typeof StatefulService> = {
  component: StatefulService,
};

export default meta;
type Story = StoryObj<typeof StatefulService>;

export const Example: Story = {
  args: {
    type: "rds",
  },
};
