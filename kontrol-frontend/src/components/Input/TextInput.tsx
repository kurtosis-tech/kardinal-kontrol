import { Stack, Text, Input } from "@chakra-ui/react";

export interface Props {
  id: string;
  label: string;
  placeholder?: string;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

const TextInput = ({ value, onChange, placeholder, id, label }: Props) => {
  return (
    <Stack flex={1}>
      <Text mb={2} as="label" htmlFor={id} m={0} fontWeight={400}>
        {label}
      </Text>
      <Input
        id={id}
        borderColor={"gray.200"}
        color={"gray.800"}
        borderRadius={"12px"}
        height={"50px"}
        placeholder={placeholder}
        value={value}
        onChange={onChange}
      />
    </Stack>
  );
};

export default TextInput;
