import type { Meta, StoryObj } from "@storybook/react";

import Table from ".";

const meta: Meta<typeof Table> = {
  component: Table,
};

export default meta;
type Story = StoryObj<typeof Table>;

export const Example: Story = {
  args: {},
};
