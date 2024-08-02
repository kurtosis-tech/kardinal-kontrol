import { Button as ChakraButton, ButtonProps } from "@chakra-ui/react";
import { Link as ReactRouterLink } from "react-router-dom";
interface Props extends ButtonProps {
  to?: string;
}

const Button = ({ isDisabled, to, children, ...props }: Props) => (
  <ChakraButton
    height={10}
    colorScheme="orange"
    borderRadius={8}
    isDisabled={isDisabled}
    fontSize={"15px"}
    fontWeight={500}
    letterSpacing={"0.46px"}
    to={to}
    as={to != null ? ReactRouterLink : "button"}
    {...props}
  >
    {children}
  </ChakraButton>
);
export default Button;
