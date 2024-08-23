import Modal from "@/components/Modal";
import { ReactElement } from "react";

interface Props {
  onConfirm: () => void;
  children: ReactElement<{ onClick: () => void }>;
  templateId: string;
}

const DeleteTemplateModal = ({ onConfirm, children, templateId }: Props) => {
  return (
    <Modal
      header={`Delete Template: ${templateId}?`}
      bodyText={
        "Are you sure you want to delete this flow template? You cannot undo this action."
      }
      onCancelText="Go back"
      onConfirmText="Delete Flow Template"
      onConfirm={onConfirm}
      target={children}
    ></Modal>
  );
};

export default DeleteTemplateModal;
