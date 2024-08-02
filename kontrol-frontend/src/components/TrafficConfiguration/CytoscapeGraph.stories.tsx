import type { Meta, StoryObj } from "@storybook/react";

import CytoscapeGraph from "./CytoscapeGraph";
import { noDevFlows, devFlow, devFlow2 } from "./mocks";
type PropsAndCustomArgs = React.ComponentProps<typeof CytoscapeGraph> & {
  graphSize?: string;
};

const meta: Meta<PropsAndCustomArgs> = {
  component: CytoscapeGraph,
};

export default meta;
type Story = StoryObj<typeof CytoscapeGraph>;

export const NoDevFlows: Story = {
  args: {
    elements: noDevFlows,
  },
};

export const DevFlow: Story = {
  args: {
    elements: devFlow,
  },
};

export const DevFlow2: Story = {
  args: {
    elements: devFlow2,
  },
};
