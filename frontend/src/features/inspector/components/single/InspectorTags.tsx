import { useState, KeyboardEvent } from "react";
import { Chip } from "@heroui/chip";
import { Input } from "@heroui/input";
import { Tag as TagIcon, Plus } from "lucide-react";
import { app } from "@wailsjs/go/models";
import { useAssetTagsMutation } from "../../hooks/useAssetTagsMutation";

interface InspectorTagsProps {
  asset: app.AssetDetails;
}

export const InspectorTags = ({ asset }: InspectorTagsProps) => {
  const [inputValue, setInputValue] = useState("");
  const { updateTags, isUpdating } = useAssetTagsMutation(asset.id);
  const currentTagNames = asset.tags || [];

  // --- HANDLER: DODAWANIE TAGU ---
  const handleAddTag = () => {
    const trimmedInput = inputValue.trim().toLowerCase();

    if (!trimmedInput) return;
    if (currentTagNames.includes(trimmedInput)) {
      setInputValue("");
      return;
    }

    const newTagsList = [...currentTagNames, trimmedInput];
    updateTags(newTagsList);
    setInputValue("");
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleAddTag();
    }
  };

  // --- HANDLER: USUWANIE TAGU ---
  const handleRemoveTag = (tagToRemove: string) => {
    const newTagsList = currentTagNames.filter((name) => name !== tagToRemove);
    updateTags(newTagsList);
  };

  return (
    <div className="p-4 flex flex-col gap-3">
      {/* NAGŁÓWEK */}
      <div className="flex items-center gap-2 text-default-500 px-1">
        <TagIcon size={14} />
        <span className="text-xs font-semibold uppercase tracking-wider">
          Tags
        </span>
      </div>

      <Input
        placeholder="Add tag..."
        value={inputValue}
        onValueChange={setInputValue}
        onKeyDown={handleKeyDown}
        isDisabled={isUpdating}
        size="sm"
        radius="md"
        variant="flat"
        startContent={<Plus size={14} className="text-default-400" />}
        classNames={{
          input: "text-small",
          inputWrapper:
            "bg-default-100/50 border-transparent hover:bg-default-200 transition-colors h-8 min-h-0",
        }}
      />

      {/* CHIPS CLOUD */}
      <div className="flex flex-wrap gap-2 min-h-[2rem] animate-appearance-in">
        {asset.tags && asset.tags.length > 0 ? (
          asset.tags.map((tag) => (
            <Chip
              key={tag}
              onClose={() => handleRemoveTag(tag)}
              variant="flat"
              size="sm"
              className="bg-primary/10 border border-primary/20 text-primary hover:bg-primary/20 transition-colors"
              classNames={{
                base: "h-6 px-1",
                content: "text-[12px] font-semibold p-1",
                closeButton: "text-primary/60 hover:text-primary",
              }}
            >
              #{tag}
            </Chip>
          ))
        ) : (
          <div className="flex items-center justify-center w-full py-2 border border-dashed border-default-200 rounded-md bg-default-50/50">
            <span className="text-[10px] text-default-400 italic">
              No tags yet
            </span>
          </div>
        )}
      </div>
    </div>
  );
};

