import { Button } from "@heroui/button";
import { FolderOpen, X, PlusCircle, Shapes } from "lucide-react";
import { Tooltip } from "@heroui/tooltip";
import { Asset } from "@/types/api";
import { useMaterialSets } from "@/layouts/sidebar/hooks/useMaterialSets"; // <-- IMPORT Z SIDEBARA

export const TabCollections = ({ asset }: { asset: Asset }) => {
  const { removeAssetFromSet } = useMaterialSets();

  const collections = asset.materialSets || [];

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
          <div className="flex items-center gap-2 overflow-hidden">
            <Shapes
              size={14}
              style={{ color: col.customColor || undefined }}
              className={
                !col.customColor
                  ? "text-primary flex-shrink-0"
                  : "flex-shrink-0"
              }
            />
            <span className="text-small text-default-700 truncate">
              {col.name}
            </span>
          </div>

          <Tooltip content="Remove from collection" closeDelay={0}>
            <button
              type="button"
              onClick={() =>
                removeAssetFromSet({ setId: col.id, assetId: asset.id })
              }
              className="text-default-400 hover:text-danger opacity-0 group-hover:opacity-100 transition-opacity p-1"
            >
              <X size={14} />
            </button>
          </Tooltip>
        </div>
      ))}
    </div>
  );
};
