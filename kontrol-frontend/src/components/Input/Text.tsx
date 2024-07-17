import { Stack, Text, Input } from "@chakra-ui/react";

interface Props {
  id: string;
  placeholder?: string;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

const TextInput = ({ value, onChange, placeholder, id }: Props) => {
  return (
    <Stack flex={1}>
      <Text mb={2} as="label" htmlFor={id} m={0}>
        Name
      </Text>
      <Input
        id={id}
        borderColor={"gray.100"}
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
