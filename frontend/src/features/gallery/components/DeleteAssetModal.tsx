import React from "react";
import {
  Modal,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
} from "@heroui/modal";
import { Button } from "@heroui/button";

interface DeleteAssetModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  fileName: string;
  isLoading?: boolean;
}

export const DeleteAssetModal = ({
  isOpen,
  onClose,
  onConfirm,
  fileName,
  isLoading,
}: DeleteAssetModalProps) => {
  return (
    <Modal isOpen={isOpen} onClose={onClose} backdrop="blur">
      <ModalContent>
        {(onClose) => (
          <>
            <ModalHeader className="flex flex-col gap-1">
              Delete Asset
            </ModalHeader>
            <ModalBody>
              <p>
                Are you sure you want to permanently delete <b>{fileName}</b>?
                This action cannot be undone and the file will be removed from
                your disk.
              </p>
            </ModalBody>
            <ModalFooter>
              <Button color="default" variant="light" onPress={onClose}>
                Cancel
              </Button>
              <Button
                color="danger"
                onPress={onConfirm}
                isLoading={isLoading}
              >
                Delete
              </Button>
            </ModalFooter>
          </>
        )}
      </ModalContent>
    </Modal>
  );
};
