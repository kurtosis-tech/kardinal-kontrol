import type { Meta, StoryObj } from "@storybook/react";

import Table from ".";

const meta: Meta<typeof Table> = {
  component: Table,
};

export default meta;
type Story = StoryObj<typeof Table>;

export const Example: Story = {
  args: {
    templates: [
      {
        description: "asdfasdf",
        name: "asdfasdf",
        "template-id": "template-5nf8a4z5hv",
      },
      {
        description: "Cool",
        name: "Hello demo",
        "template-id": "template-dpotdpd9b6",
      },
      {
        description: "123123",
        name: "123123",
        "template-id": "template-s0muzww1lg",
      },
    ],
  },
};
