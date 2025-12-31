import { Button } from "@heroui/button";
import { FolderOpen, X, PlusCircle, Shapes, Copy } from "lucide-react";
import { Tooltip } from "@heroui/tooltip";
import { app } from "@wailsjs/go/models";
import { useMaterialSets } from "@/layouts/sidebar/hooks/useMaterialSets";
import { ClipboardSetText } from "../../../../../../wailsjs/runtime/runtime";
import { addToast } from "@heroui/toast";

export const TabCollections = ({ asset }: { asset: app.AssetDetails }) => {
  const { removeAssetFromSet } = useMaterialSets();

  const collections = asset.materialSets || [];

  const handleCopyName = async (name: string) => {
    await ClipboardSetText(name);
    addToast({
      title: "Copied!",
      description: `Collection name "${name}" copied to clipboard`,
      color: "success",
      variant: "flat",
      timeout: 1500,
    });
  };

  if (collections.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-8 gap-2 text-default-400">
        <FolderOpen size={24} className="opacity-20" />
        <span className="text-small font-medium text-default-600">
          Not in any collection
        </span>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-2 p-1">
      {collections.map((col) => (
        <div
          key={col.id}
          className="flex items-center justify-between p-2 rounded-medium bg-default-50 border border-default-200 group hover:border-default-300 transition-colors"
        >
          <div 
            className="flex items-center gap-2 overflow-hidden cursor-pointer flex-1"
            onClick={() => handleCopyName(col.name)}
          >
            <Shapes
              size={14}
              style={{ color: col.customColor || undefined }}
              className={
                !col.customColor
                  ? "text-primary flex-shrink-0"
                  : "flex-shrink-0"
              }
            />
            <span className="text-small text-default-700 truncate group-hover:text-primary transition-colors">
              {col.name}
            </span>
            <Copy size={10} className="opacity-0 group-hover:opacity-100 transition-opacity text-default-400" />
          </div>

          <Tooltip content="Remove from collection" closeDelay={0}>
            <button
              type="button"
              onClick={() =>
                removeAssetFromSet({ setId: col.id, assetId: asset.id })
              }
              className="text-default-400 hover:text-danger opacity-0 group-hover:opacity-100 transition-opacity p-1 ml-2"
            >
              <X size={14} />
            </button>
          </Tooltip>
        </div>
      ))}
    </div>
  );
};

