import type { Meta, StoryObj } from "@storybook/react";

import TextInput from "./TextInput";

const meta: Meta<typeof TextInput> = {
  component: TextInput,
};

export default meta;
type Story = StoryObj<typeof TextInput>;

export const Example: Story = {
  args: {
    id: "example-a11y-id",
    value: "Example value",
    onChange: () => {},
  },
};
