import type { Meta, StoryObj } from "@storybook/react";

import Text from "./Text";

const meta: Meta<typeof Text> = {
  component: Text,
};

export default meta;
type Story = StoryObj<typeof Text>;

export const Example: Story = {
  args: {
    id: "example-a11y-id",
    value: "Example value",
    onChange: () => {},
  },
};
