import {
  Modal as ChakraModal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  Button,
  Text,
  useDisclosure,
  Flex,
  ModalFooter,
} from "@chakra-ui/react";
import { cloneElement, ReactElement, ReactNode } from "react";

interface Props {
  header: string;
  bodyText?: string;
  target: ReactElement<{ onClick: () => void }>;
  children?: ReactNode;
  onConfirm?: () => void;
  onConfirmText?: string;
  onCancel?: () => void;
  onCancelText?: string;
  wide?: boolean;
}

const Modal = ({
  header,
  bodyText,
  children,
  target,
  onConfirm,
  onConfirmText,
  onCancel,
  onCancelText,
  wide,
}: Props) => {
  const { isOpen, onOpen, onClose } = useDisclosure();

  return (
    <>
      {cloneElement(target, { onClick: onOpen })}
      <ChakraModal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent
          borderRadius={"12px"}
          textAlign={"center"}
          maxWidth={wide ? 680 : undefined}
        >
          <ModalHeader color="gray.800" fontSize={"md"} pb={0} pt={8}>
            {header}
          </ModalHeader>
          <ModalBody pt={0} px={8} mt={2}>
            {bodyText != null && <Text fontSize={"md"}>{bodyText}</Text>}
            {children}
          </ModalBody>

          <ModalFooter pb={8}>
            <Flex justifyContent={"center"} width={"100%"}>
              <Button
                variant="ghost"
                mr={3}
                onClick={() => {
                  onCancel && onCancel();
                  onClose();
                }}
              >
                {onCancelText}
              </Button>
              <Button
                colorScheme="red"
                onClick={() => {
                  onConfirm && onConfirm();
                  onClose();
                }}
              >
                {onConfirmText}
              </Button>
            </Flex>
          </ModalFooter>
        </ModalContent>
      </ChakraModal>
    </>
  );
};

export default Modal;
