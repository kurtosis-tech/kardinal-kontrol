import {
  Modal as ChakraModal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  Button,
  Text,
  useDisclosure,
} from "@chakra-ui/react";

interface Props {
  error: Error;
  resetErrorBoundary: () => void;
}

const Fallback = ({ error, resetErrorBoundary }: Props) => {
  // Call resetErrorBoundary() to reset the error boundary and retry the render.
  console.debug({ error });
  const { onClose } = useDisclosure();

  return (
    <ChakraModal isOpen={error != null} onClose={onClose}>
      <ModalOverlay />
      <ModalContent maxW={"xlg"} mx={20}>
        <ModalHeader>Oopsie, something went wrong!</ModalHeader>
        <ModalBody>
          <pre style={{ color: "red" }}>{error.message}</pre>
          <pre style={{ display: "flex", flexDirection: "column" }}>
            {error.stack?.split("\n").map((line) => {
              return (
                <code
                  style={{
                    color: line.includes("/src/") ? "black" : "darkgray",
                  }}
                >
                  {line}
                </code>
              );
            })}
          </pre>
        </ModalBody>

        <ModalFooter>
          <Button onClick={resetErrorBoundary}>Close</Button>
        </ModalFooter>
      </ModalContent>
    </ChakraModal>
  );
};

export default Fallback;
