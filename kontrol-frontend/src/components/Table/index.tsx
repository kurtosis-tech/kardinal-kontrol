import { Box, Table, Tbody } from "@chakra-ui/react";
import Hat from "./Hat";
import Head from "./Head";
import Row from "./Row";
import { useState } from "react";

const rows = ["a"];

const FlowConfigurationTable = () => {
  const [expandedRowId, setExpandedRowId] = useState<string | null>(null);

  const handleExpandRow = (rowId: string | null) => {
    setExpandedRowId(rowId);
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
          {rows.map((rowId) => (
            <Row
              id={rowId}
              isExpanded={expandedRowId === rowId}
              onExpandRow={handleExpandRow}
            />
          ))}
        </Tbody>
      </Table>
    </Box>
  );
};

export default FlowConfigurationTable;
