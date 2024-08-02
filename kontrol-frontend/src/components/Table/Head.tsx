import { Thead, Tr, Th } from "@chakra-ui/react";

const Head = () => {
  return (
    <Thead background={"gray.50"}>
      <Tr>
        <Th
          width={12}
          textTransform={"none"}
          fontSize={"sm"}
          fontWeight={500}
          pl={0}
        >
          {/* empty */}
        </Th>
        <Th textTransform={"none"} fontSize={"sm"} fontWeight={500} pl={0}>
          Flow template
        </Th>
        <Th textTransform={"none"} fontSize={"sm"} fontWeight={500}>
          Service
        </Th>
        <Th textTransform={"none"} fontSize={"sm"} fontWeight={500}>
          Data layer
        </Th>
        <Th
          textTransform={"none"}
          fontSize={"sm"}
          fontWeight={500}
          textAlign={"right"}
        >
          Actions
        </Th>
      </Tr>
    </Thead>
  );
};

export default Head;
