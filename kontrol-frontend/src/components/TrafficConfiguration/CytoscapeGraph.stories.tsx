import type { Meta, StoryObj } from "@storybook/react";

import CytoscapeGraph from "./CytoscapeGraph";
import mockData from "./mock-data";

const meta: Meta<typeof CytoscapeGraph> = {
  component: CytoscapeGraph,
};

export default meta;
type Story = StoryObj<typeof CytoscapeGraph>;

export const Example: Story = {
  args: {
    elements: mockData,
  },
};
