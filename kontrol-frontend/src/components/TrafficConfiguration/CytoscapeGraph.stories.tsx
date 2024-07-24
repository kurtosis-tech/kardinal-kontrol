import type { Meta, StoryObj } from "@storybook/react";

import CytoscapeGraph from "./CytoscapeGraph";
import {
  groupedByVersion,
  nestedSidecars,
  noDevFlows,
  groupedByService,
  coloredByEnv,
  devFlow,
} from "./mocks";
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

export const GroupedByService: Story = {
  args: {
    elements: groupedByService,
  },
};

export const GroupedByVersion: Story = {
  args: {
    elements: groupedByVersion,
  },
};

export const NestedSidecars: Story = {
  args: {
    elements: nestedSidecars,
  },
};

export const ColoredByEnv: Story = {
  args: {
    elements: coloredByEnv,
  },
};

export const DevFlow: Story = {
  args: {
    elements: devFlow,
    // layout: {
    //   name: "breadthfirst",
    //   directed: true,
    //   grid: true,
    //   roots: ["node[type='gateway']"],
    //   spacingFactor: 1,
    // },
    layout: {
      name: "dagre",
      // @ts-expect-error cytoscape types are not great
      spacingFactor: 1.6,
      nodeSep: 40,
      nodeDimensionsIncludeLabels: true,
      // ranker: "network-simplex",
      rankDir: "LR",
      align: "UL",
      transform: (_, pos: cytoscape.Position) => {
        const roundedPos = {
          ...pos,
          y: Math.round(Math.ceil(pos.y / 100)) * 100,
        };
        return roundedPos;
      },
    },
  },
};
