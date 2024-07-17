import { FiInfo } from "react-icons/fi";

import { Stack, Flex, Text, Select, Tooltip, Icon } from "@chakra-ui/react";

interface Option {
  label: string;
  value: string;
}

interface Props {
  id: string;
  label: string;
  tooltip?: string;
  placeholder?: string;
  options: Option[];
  onChange: (e: React.ChangeEvent<HTMLSelectElement>) => void;
  value: string;
}
const SelectInput = ({
  id,
  label,
  tooltip,
  placeholder,
  options,
  onChange,
  value,
}: Props) => {
  return (
    <Stack flex={1}>
      <Flex alignItems="center">
        <Text mr={2} htmlFor={id} as="label" m={0}>
          {label}
        </Text>
        {tooltip && (
          <Tooltip label={tooltip}>
            <span>
              <Icon as={FiInfo} boxSize={4} />
            </span>
          </Tooltip>
        )}
      </Flex>
      <Select
        id={id}
        placeholder={placeholder}
        borderColor={"gray.100"}
        borderRadius={"12px"}
        height={"50px"}
        onChange={onChange}
      >
        {options.map((option) => (
          <option value={option.value} selected={option.value === value}>
            {option.label}
          </option>
        ))}
      </Select>
    </Stack>
  );
};

export default SelectInput;
