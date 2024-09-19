// styled wrapper for Chakra UI Table
import {
  Table as CTable,
  Thead as CThead,
  Tbody as CTbody,
  Tr as CTr,
  Th as CTh,
  Td as CTd,
  TableContainer as CTableContainer,
  TableColumnHeaderProps,
  TableProps,
  TableHeadProps,
  TableBodyProps,
  TableRowProps,
  TableCellProps,
  TableContainerProps,
} from "@chakra-ui/react";

export const Table = (props: TableProps) => (
  <CTable variant="simple" {...props} />
);

export const Thead = (props: TableHeadProps) => <CThead {...props} />;

export const Tbody = (props: TableBodyProps) => <CTbody {...props} />;

export const Tr = (props: TableRowProps) => <CTr {...props} />;

export const Th = (
  props: TableColumnHeaderProps & { isFirst?: boolean; isLast?: boolean },
) => (
  <CTh
    bg={"gray.100"}
    px={3}
    py={2}
    borderTopLeftRadius={props.isFirst ? 6 : 0}
    borderBottomLeftRadius={props.isFirst ? 6 : 0}
    borderTopRightRadius={props.isLast ? 6 : 0}
    borderBottomRightRadius={props.isLast ? 6 : 0}
    color={"gray.700"}
    fontWeight={500}
    fontSize={12}
    fontStyle={"normal"}
    textTransform={"none"}
    border={"none"}
    {...props}
  />
);

export const Td = (props: TableCellProps) => (
  <CTd
    px={3}
    py={2}
    borderColor={"gray.100"}
    color={"gray.900"}
    fontSize={12}
    fontWeight={400}
    {...props}
  />
);

export const TableContainer = (props: TableContainerProps) => (
  <CTableContainer
    px={2}
    py={2}
    background={"white"}
    borderRadius={12}
    {...props}
  />
);
