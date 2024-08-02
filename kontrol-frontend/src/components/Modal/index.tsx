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

const Modal = () => {
  const { isOpen, onOpen, onClose } = useDisclosure();

  return (
    <>
      <Button onClick={onOpen}>Open Modal</Button>

      <ChakraModal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent borderRadius={"12px"} textAlign={"center"}>
          <ModalHeader color="gray.800" fontSize={"md"} pb={0} pt={8}>
            Delete this flow?
          </ModalHeader>
          <ModalBody pt={0} px={8} mt={2}>
            <Text fontSize={"md"}>
              Are you sure you want to delete this flow?
              <br />
              You cannot undo this action.
            </Text>
          </ModalBody>

          <ModalFooter pb={8}>
            <Flex justifyContent={"center"} width={"100%"}>
              <Button variant="ghost" mr={3} onClick={onClose}>
                Go back
              </Button>
              <Button colorScheme="red" onClick={onClose}>
                Delete flow
              </Button>
            </Flex>
          </ModalFooter>
        </ModalContent>
      </ChakraModal>
    </>
  );
};

export default Modal;
