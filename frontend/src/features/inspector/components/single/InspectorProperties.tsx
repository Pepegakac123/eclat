import { useState, useEffect } from "react";
import { Textarea } from "@heroui/input";
import { Star, FileText } from "lucide-react";
import { Asset } from "@/types/api";
import { useAssetMutation } from "../../hooks/useAsset";

interface InspectorPropertiesProps {
  asset: Asset;
}

export const InspectorProperties = ({ asset }: InspectorPropertiesProps) => {
  const { patch } = useAssetMutation(asset.id);

  const [localDescription, setLocalDescription] = useState(
    asset.description || "",
  );

  useEffect(() => {
    setLocalDescription(asset.description || "");
  }, [asset.id, asset.description]);

  const [hoverRating, setHoverRating] = useState(0);

  // --- HANDLERS ---

  const handleDescriptionSave = () => {
    if (localDescription === (asset.description || "")) return;
    patch({ description: localDescription });
  };

  const handleRatingClick = (newRating: number) => {
    // Jeśli klikniesz w tę samą ocenę, to ją wyzeruj (toggle)
    const finalRating = asset.rating === newRating ? 0 : newRating;
    patch({ rating: finalRating });
  };

  return (
    <div className="p-4 flex flex-col gap-5">
      {/* 1. RATING SECTION */}
      <div className="flex flex-col gap-1">
        <span className="text-[10px] uppercase font-bold text-default-400 tracking-wider flex items-center gap-1">
          Rating
        </span>
        <div
          className="flex items-center gap-1 w-fit"
          onMouseLeave={() => setHoverRating(0)}
        >
          {[1, 2, 3, 4, 5].map((starValue) => {
            const isFilled =
              hoverRating > 0
                ? starValue <= hoverRating
                : starValue <= asset.rating;

            return (
              <button
                key={starValue}
                type="button"
                className="transition-transform hover:scale-110 focus:outline-none hover:cursor-pointer"
                onMouseEnter={() => setHoverRating(starValue)}
                onClick={() => handleRatingClick(starValue)}
              >
                <Star
                  size={20}
                  fill={isFilled ? "currentColor" : "none"}
                  className={`transition-colors ${
                    isFilled ? "text-primary" : "text-default-300"
                  }`}
                />
              </button>
            );
          })}
        </div>
      </div>

      {/* 2. DESCRIPTION SECTION */}
      <div className="flex flex-col gap-1">
        <span className="text-[10px] uppercase font-bold text-default-400 tracking-wider flex items-center gap-1">
          <FileText size={10} /> Description
        </span>
        <Textarea
          placeholder="Add a description..."
          minRows={2}
          maxRows={8}
          variant="faded"
          size="sm"
          value={localDescription}
          onValueChange={setLocalDescription}
          onBlur={handleDescriptionSave}
          classNames={{
            input: "text-small",
            inputWrapper:
              "bg-default-50 border-default-200 hover:border-default-300 transition-colors",
          }}
        />
      </div>
    </div>
  );
};
