import type { Meta, StoryObj } from "@storybook/react";

import Select from "./Select";

const meta: Meta<typeof Select> = {
  component: Select,
};

export default meta;
type Story = StoryObj<typeof Select>;

const options = [
  { label: "Option 1", value: "option1" },
  { label: "Option 2", value: "option2" },
];

export const Example: Story = {
  args: {
    id: "example-a11y-id",
    label: "Example label",
    options,
    value: options[1].value,
    onChange: () => {},
  },
};
