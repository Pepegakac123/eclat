import { useState, useMemo } from "react";
import {
  Modal,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
} from "@heroui/modal";
import { Button } from "@heroui/button";
import { Input } from "@heroui/input";
import { ScrollShadow } from "@heroui/scroll-shadow";
import { Search, Plus, Check, Shapes, FolderPlus } from "lucide-react";
import { MaterialSet } from "@/types/api"; // Upewnij się, że masz import MaterialSet
import { app } from "@wailsjs/go/models";
import { useMaterialSets } from "@/layouts/sidebar/hooks/useMaterialSets";
import {
  MaterialSetForm,
  MaterialSetFormModal,
} from "@/layouts/sidebar/MaterialSetFormModal";

interface AddToCollectionModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  asset: app.AssetDetails;
}

export const AddToCollectionModal = ({
  isOpen,
  onOpenChange,
  asset,
}: AddToCollectionModalProps) => {
  const { materialSets, addAssetToSet, createMaterialSet } = useMaterialSets();

  const [searchQuery, setSearchQuery] = useState("");
  const [loadingSetId, setLoadingSetId] = useState<number | null>(null);

  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isCreatingSet, setIsCreatingSet] = useState(false);

  const filteredSets = useMemo(() => {
    return materialSets.filter((set) =>
      set.name.toLowerCase().includes(searchQuery.toLowerCase()),
    );
  }, [materialSets, searchQuery]);

  const assetSetIds = useMemo(() => {
    return (asset.materialSets || []).map((s) => s.id);
  }, [asset.materialSets]);

  const handleAdd = async (setId: number) => {
    setLoadingSetId(setId);
    try {
      await addAssetToSet({ setId, assetId: asset.id });
    } finally {
      setLoadingSetId(null);
    }
  };

  const handleCreateSet = async (
    data: MaterialSetForm,
    onCloseForm: () => void,
  ) => {
    setIsCreatingSet(true);
    try {
      const newSet = await createMaterialSet(data);
      if (newSet && newSet.id) {
        await addAssetToSet({ setId: newSet.id, assetId: asset.id });
      }
      onCloseForm();
      // setSearchQuery("") - to zostanie obsłużone w onOpenChange
    } catch (error) {
      console.error("Failed to create set", error);
    } finally {
      setIsCreatingSet(false);
    }
  };

  // Logika zamykania modala tworzenia
  const handleCreateModalOpenChange = (open: boolean) => {
    setIsCreateModalOpen(open);

    // Jeśli zamykamy modal (czy to przez Cancel, X, czy po sukcesie)
    // resetujemy wyszukiwarkę w głównym oknie
    if (!open) {
      setSearchQuery("");
    }
  };

  return (
    <>
      <Modal
        isOpen={isOpen}
        onOpenChange={onOpenChange}
        scrollBehavior="inside"
        backdrop="blur"
        size="md"
      >
        <ModalContent>
          {(onClose) => (
            <>
              <ModalHeader className="flex flex-col gap-1">
                Add to Collection
                <span className="text-tiny font-normal text-default-400">
                  Select collections for{" "}
                  <span className="font-mono text-foreground">
                    {asset.fileName}
                  </span>
                </span>
              </ModalHeader>

              <ModalBody className="pt-0">
                <Input
                  placeholder="Search collections..."
                  startContent={
                    <Search size={16} className="text-default-400" />
                  }
                  value={searchQuery}
                  onValueChange={setSearchQuery}
                  variant="faded"
                  size="sm"
                  classNames={{ inputWrapper: "bg-default-100" }}
                />

                <ScrollShadow className="h-[300px] mt-2">
                  <div className="flex flex-col gap-1">
                    {filteredSets.length > 0 ? (
                      filteredSets.map((set) => {
                        const isAlreadyAdded = assetSetIds.includes(set.id);
                        const isLoading = loadingSetId === set.id;

                        return (
                          <div
                            key={set.id}
                            className="flex items-center justify-between p-2 rounded-lg hover:bg-default-100 transition-colors border border-transparent hover:border-default-200"
                          >
                            <div className="flex items-center gap-3 overflow-hidden">
                              <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary">
                                <Shapes
                                  size={16}
                                  style={{
                                    color: set.customColor || undefined,
                                  }}
                                  className={
                                    !set.customColor ? "text-primary" : ""
                                  }
                                />
                              </div>
                              <span className="text-small text-default-700 truncate">
                                {set.name}
                              </span>
                            </div>

                            {isAlreadyAdded ? (
                              <Button
                                size="sm"
                                isIconOnly
                                variant="flat"
                                color="success"
                                className="bg-success/10 text-success cursor-default rounded-full"
                              >
                                <Check size={16} />
                              </Button>
                            ) : (
                              <Button
                                size="sm"
                                isIconOnly
                                variant="light"
                                className="text-default-400 hover:text-primary hover:bg-primary/10 rounded-full"
                                onPress={() => handleAdd(set.id)}
                                isLoading={isLoading}
                              >
                                {!isLoading && <Plus size={18} />}
                              </Button>
                            )}
                          </div>
                        );
                      })
                    ) : (
                      <div className="flex flex-col items-center justify-center py-8 text-default-400 gap-2">
                        <p>No collections found.</p>
                        {/* Przycisk tworzenia z nazwą z wyszukiwarki */}
                        <Button
                          size="sm"
                          variant="flat"
                          onPress={() => setIsCreateModalOpen(true)}
                        >
                          Create "{searchQuery}"
                        </Button>
                      </div>
                    )}
                  </div>
                </ScrollShadow>
              </ModalBody>

              <ModalFooter className="flex justify-between items-center">
                <Button
                  variant="light"
                  color="primary"
                  startContent={<FolderPlus size={18} />}
                  onPress={() => setIsCreateModalOpen(true)}
                >
                  New Collection
                </Button>

                <Button variant="light" onPress={onClose}>
                  Done
                </Button>
              </ModalFooter>
            </>
          )}
        </ModalContent>
      </Modal>

      <MaterialSetFormModal
        isOpen={isCreateModalOpen}
        onOpenChange={handleCreateModalOpenChange}
        mode="create"
        isLoading={isCreatingSet}
        onSubmit={handleCreateSet}
        // Przekazujemy searchQuery jako initialData (tylko nazwę).
        // Rzutujemy na 'any' lub 'MaterialSet', bo brakuje nam ID,
        // ale formularz obsłuży to poprawnie (użyje defaultów dla reszty).
        initialData={
          searchQuery ? ({ name: searchQuery } as MaterialSet) : undefined
        }
      />
    </>
  );
};

