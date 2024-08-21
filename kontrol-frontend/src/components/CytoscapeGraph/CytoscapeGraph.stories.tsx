import type { Meta, StoryObj } from "@storybook/react";

import CytoscapeGraph from ".";
import { mockResponse } from "./mocks";
type PropsAndCustomArgs = React.ComponentProps<typeof CytoscapeGraph> & {
  graphSize?: string;
};

const meta: Meta<PropsAndCustomArgs> = {
  component: CytoscapeGraph,
};

export default meta;
type Story = StoryObj<typeof CytoscapeGraph>;

export const OBDemo: Story = {
  args: {
    elements: mockResponse,
  },
};
