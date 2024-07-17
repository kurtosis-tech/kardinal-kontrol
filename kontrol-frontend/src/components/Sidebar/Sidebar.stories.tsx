import type { Meta, StoryObj } from "@storybook/react";
import { withRouter } from "storybook-addon-remix-react-router";

import Sidebar from ".";

const meta: Meta<typeof Sidebar> = {
  decorators: [withRouter()],
  component: Sidebar,
};

export default meta;
type Story = StoryObj<typeof Sidebar>;

export const Example: Story = {
  args: {},
};
