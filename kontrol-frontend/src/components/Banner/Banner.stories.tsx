import type { Meta, StoryObj } from "@storybook/react";

import Banner from ".";

const meta: Meta<typeof Banner> = {
  component: Banner,
};

export default meta;
type Story = StoryObj<typeof Banner>;

export const Example: Story = {
  args: {},
};
