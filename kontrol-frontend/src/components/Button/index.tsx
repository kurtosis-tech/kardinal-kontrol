import { Button as ChakraButton, ButtonProps } from "@chakra-ui/react";
const Button = (props: ButtonProps) => (
  <ChakraButton
    colorScheme="blackAlpha"
    bg={"gray.900"}
    borderRadius={8}
    color={"white"}
    {...props}
  >
    Go back
  </ChakraButton>
);
export default Button;
