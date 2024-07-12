import type { Meta, StoryObj } from "@storybook/react";

import CytoscapeGraph from "./CytoscapeGraph";
import { fullGraph, smallGraph } from "./mock-data";
type PropsAndCustomArgs = React.ComponentProps<typeof CytoscapeGraph> & {
  graphSize?: string;
};

const meta: Meta<PropsAndCustomArgs> = {
  component: CytoscapeGraph,
  render: ({ graphSize }) => {
    const elements = graphSize === "full" ? fullGraph : smallGraph;
    return <CytoscapeGraph elements={elements} />;
  },
  argTypes: {
    graphSize: {
      options: ["full", "small"],
      control: { type: "radio" },
    },
  },
};

export default meta;
type Story = StoryObj<typeof CytoscapeGraph>;

export const Example: Story = {};
