import type { Meta, StoryObj } from "@storybook/react";

import { FaStripeS, FaAmazon } from "react-icons/fa";

import Chip from ".";

const meta: Meta<typeof Chip> = {
  component: Chip,
};

export default meta;
type Story = StoryObj<typeof Chip>;

export const Short: Story = {
  args: {
    icon: <FaStripeS />,
    children: "Stripe",
  },
};
export const Long: Story = {
  args: {
    icon: <FaAmazon />,
    children: "Amazon Relational Database Service",
  },
};
