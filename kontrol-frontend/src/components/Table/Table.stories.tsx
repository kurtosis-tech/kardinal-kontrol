import type { Meta, StoryObj } from "@storybook/react";

import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
} from "@/components/Table";
import { NodeVersion } from "@/types";

const meta: Meta = {
  component: Table,
};

export default meta;
type Story = StoryObj;

const versions: NodeVersion[] = [
  { flowId: "flow-1", imageTag: "v1", isBaseline: true },
  { flowId: "flow-2", imageTag: "v2", isBaseline: false },
];

export const Example: Story = {
  render: () => (
    <TableContainer>
      <Table>
        <Thead>
          <Tr>
            <Th>Flow ID</Th>
            <Th>Image Tag</Th>
          </Tr>
        </Thead>
        <Tbody>
          {versions.map(({ flowId, imageTag }) => {
            return (
              <Tr key={flowId}>
                <Td>{flowId}</Td>
                <Td>{imageTag}</Td>
              </Tr>
            );
          })}
        </Tbody>
      </Table>
    </TableContainer>
  ),
};
