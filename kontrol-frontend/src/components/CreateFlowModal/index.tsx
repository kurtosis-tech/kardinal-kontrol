import { Flex, Stack } from "@chakra-ui/react";
import Modal from "@/components/Modal";
import Input from "@/components/Input";
import { ReactElement } from "react";

interface Props {
  onConfirm: () => void;
  children: ReactElement<{ onClick: () => void }>;
  templateId: string;
}
const CreateFlowModal = ({ children, onConfirm, templateId }: Props) => {
  return (
    <Modal
      header={`Create Flow from Template: ${templateId}`}
      bodyText={
        "Configure the dev flow template paramters below to create a new flow."
      }
      onCancelText="Go back"
      onConfirmText="Create Flow"
      onConfirm={onConfirm}
      target={children}
      wide
    >
      <Stack textAlign={"left"} mt={4}>
        <Flex gap={2}>
          <Input.Text
            label="Service"
            id="description"
            value={"frontend"}
            placeholder="e.g. my-service:latest"
            onChange={() => {}}
            isDisabled
          />
          <Input.Text
            label="Image"
            id="description"
            value={"kurtosistech/frontend:demo-frontend"}
            placeholder="e.g. my-service:latest"
            onChange={() => {}}
          />
        </Flex>
        <Flex gap={2}>
          <Input.Text
            label="Service"
            id="description"
            value={"cartservice"}
            placeholder="e.g. my-service:latest"
            onChange={() => {}}
            isDisabled
          />
          <Input.Text
            label="Image"
            id="description"
            value={"kurtosistech/frontend:demo-cartservice"}
            placeholder="e.g. my-service:latest"
            onChange={() => {}}
          />
        </Flex>
        <Flex gap={2}>
          <Input.Text
            label="Service"
            id="description"
            value={"productcatalogservice"}
            placeholder="e.g. my-service:latest"
            onChange={() => {}}
            isDisabled
          />
          <Input.Text
            label="Image"
            id="description"
            value={"kurtosistech/frontend:demo-productcatalogservice"}
            placeholder="e.g. my-service:latest"
            onChange={() => {}}
          />
        </Flex>
      </Stack>
    </Modal>
  );
};

export default CreateFlowModal;
