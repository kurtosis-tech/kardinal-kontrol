import type { Meta, StoryObj } from "@storybook/react";
import { Button } from "@chakra-ui/react";

import DeleteTemplateModal from ".";

const meta: Meta<typeof DeleteTemplateModal> = {
  component: DeleteTemplateModal,
};

export default meta;
type Story = StoryObj<typeof DeleteTemplateModal>;

export const Example: Story = {
  args: {
    children: <Button>Open DeleteTemplateModal</Button>,
    templateId: "example-123",
  },
};
