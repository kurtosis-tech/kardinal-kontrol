import type { Meta, StoryObj } from "@storybook/react";

import SelectInput from "./SelectInput";

const meta: Meta<typeof SelectInput> = {
  component: SelectInput,
};

export default meta;
type Story = StoryObj<typeof SelectInput>;

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
