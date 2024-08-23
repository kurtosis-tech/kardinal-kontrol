import { Stack, Text, Input, InputProps } from "@chakra-ui/react";

export interface Props extends InputProps {
  id: string;
  label: string;
  placeholder?: string;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

const TextInput = ({
  value,
  onChange,
  placeholder,
  id,
  label,
  ...props
}: Props) => {
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
        {...props}
      />
    </Stack>
  );
};

export default TextInput;
