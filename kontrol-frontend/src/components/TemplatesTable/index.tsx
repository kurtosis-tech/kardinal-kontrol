import { Box, Table, Tbody } from "@chakra-ui/react";
import Hat from "./Hat";
import Head from "./Head";
import Row from "./Row";
import { useState } from "react";
import { Template } from "@/types";

interface Props {
  templates: Template[];
}
const FlowConfigurationTable = ({ templates }: Props) => {
  const [expandedRowId, setExpandedRowId] = useState<string | null>(null);

  const handleExpandRow = (rowId: string | null) => {
    setExpandedRowId(rowId);
  };

  const handleDelete = (templateId: string) => {
    console.log("DELETE TEMPLATE", templateId);
  };

  return (
    <Box
      borderWidth="1px"
      borderRadius="lg"
      borderColor="gray.200"
      overflow="hidden"
      background={"white"}
      mt={8}
    >
      <Hat>Flow configuration</Hat>
      <Table variant="simple">
        <Head />
        <Tbody>
          {templates.map((t) => (
            <Row
              template={t}
              id={t["template-id"]}
              isExpanded={expandedRowId === t["template-id"]}
              onExpandRow={handleExpandRow}
              onDelete={handleDelete}
            />
          ))}
        </Tbody>
      </Table>
    </Box>
  );
};

export default FlowConfigurationTable;
