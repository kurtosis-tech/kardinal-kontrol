import type { Meta, StoryObj } from "@storybook/react";
import { Button } from "@chakra-ui/react";

import Modal from ".";

const meta: Meta<typeof Modal> = {
  component: Modal,
};

export default meta;
type Story = StoryObj<typeof Modal>;

export const Example: Story = {
  args: {
    header: "Delete this flow?",
    bodyText:
      "Are you sure you want to delete this flow? You cannot undo this action.",
    onCancel: () => console.log("Cancelled"),
    onCancelText: "Go back",
    onConfirm: () => console.log("Confirmed"),
    onConfirmText: "Delete Flow",
    children: <Button>Open Modal</Button>,
  },
};
