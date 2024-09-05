import type { Meta, StoryObj } from "@storybook/react";
import { Button } from "@chakra-ui/react";

import CreateFlowModal from ".";

const meta: Meta<typeof CreateFlowModal> = {
  component: CreateFlowModal,
};

export default meta;
type Story = StoryObj<typeof CreateFlowModal>;

export const Example: Story = {
  args: {
    children: <Button>Open CreateFlowModal</Button>,
    templateId: "example-123",
  },
};
