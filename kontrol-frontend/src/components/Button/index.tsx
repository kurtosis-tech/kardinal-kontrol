import { Button as ChakraButton, ButtonProps } from "@chakra-ui/react";
const Button = ({ isDisabled, ...props }: ButtonProps) => (
  <ChakraButton
    colorScheme="blackAlpha"
    bg={isDisabled ? "gray.100" : "gray.900"}
    borderRadius={8}
    color={isDisabled ? "gray.800" : "white"}
    isDisabled={isDisabled}
    {...props}
  >
    Go back
  </ChakraButton>
);
export default Button;
