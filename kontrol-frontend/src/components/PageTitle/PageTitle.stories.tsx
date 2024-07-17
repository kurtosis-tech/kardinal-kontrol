import type { Meta, StoryObj } from "@storybook/react";

import PageTitle from ".";

const meta: Meta<typeof PageTitle> = {
  component: PageTitle,
};

export default meta;
type Story = StoryObj<typeof PageTitle>;

export const Example: Story = {
  args: {
    title: "Create new flow configuration",
    children: "Update traffic control and data isolation details below",
  },
};
