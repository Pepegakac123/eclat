import { useState } from "react"; // Dodaj useState
import { Button } from "@heroui/button";
import { Tooltip } from "@heroui/tooltip";
import { Heart, FolderOpen, ExternalLink, PlusCircle } from "lucide-react";
import { Asset } from "@/types/api";
import { useAssetActions } from "../../hooks/useAssetActions";
import { AddToCollectionModal } from "./AddToCollectionModal";

interface InspectorFooterProps {
  asset: Asset;
}

export const InspectorFooter = ({ asset }: InspectorFooterProps) => {
  const { toggleFavorite, openInExplorer, openInProgram } = useAssetActions(
    asset.id,
  );

  const [isAddModalOpen, setIsAddModalOpen] = useState(false);

  return (
    <>
      <div className="flex-none p-3 border-t border-default-200 bg-content1 flex gap-2 items-center">
        <Tooltip content="Open File" closeDelay={0}>
          <Button
            isIconOnly
            variant="light"
            size="sm"
            className="text-default-500 hover:text-default-900"
            onPress={() => openInProgram(asset.filePath)}
          >
            <ExternalLink size={18} />
          </Button>
        </Tooltip>

        <Tooltip content="Show in Explorer" closeDelay={0}>
          <Button
            isIconOnly
            variant="light"
            size="sm"
            className="text-default-500 hover:text-default-900"
            onPress={() => openInExplorer(asset.filePath)}
          >
            <FolderOpen size={18} />
          </Button>
        </Tooltip>

        <div className="flex-1" />

        {/* ADD TO COLLECTION BUTTON */}
        <Tooltip content="Add to Collection" closeDelay={0}>
          <Button
            isIconOnly
            variant="light"
            size="sm"
            className="text-default-500 hover:text-primary"
            onPress={() => setIsAddModalOpen(true)}
          >
            <PlusCircle size={18} />
          </Button>
        </Tooltip>

        <Tooltip
          content={asset.isFavorite ? "Unfavorite" : "Favorite"}
          closeDelay={0}
        >
          <Button
            isIconOnly
            variant={asset.isFavorite ? "flat" : "light"}
            color={asset.isFavorite ? "danger" : "default"}
            size="sm"
            className={
              asset.isFavorite
                ? "text-danger bg-danger/10"
                : "text-default-500 hover:text-danger/70"
            }
            onPress={() => toggleFavorite()}
          >
            <Heart
              size={18}
              fill={asset.isFavorite ? "currentColor" : "none"}
            />
          </Button>
        </Tooltip>
      </div>

      {/*RENDER MODALA */}
      <AddToCollectionModal
        isOpen={isAddModalOpen}
        onOpenChange={setIsAddModalOpen}
        asset={asset}
      />
    </>
  );
};
