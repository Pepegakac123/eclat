import { Button } from "@heroui/button";
import { SearchX, FilterX, FolderSearch, Settings } from "lucide-react";
import { useGalleryStore } from "../stores/useGalleryStore";
import { useNavigate } from "react-router-dom";

interface NoResultsProps {
  variant?: "no-matches" | "empty-library";
}

export const NoResults = ({ variant = "no-matches" }: NoResultsProps) => {
  const resetFilters = useGalleryStore((state) => state.resetFilters);
  const navigate = useNavigate();

  // --- WARIANT 1: Pusta biblioteka (Onboarding) ---
  if (variant === "empty-library") {
    return (
      <div className="flex flex-col items-center justify-center w-full h-[60vh] text-center p-4 animate-appearance-in">
        <div className="relative mb-6">
          <div className="w-24 h-24 bg-primary/10 rounded-full flex items-center justify-center">
            <FolderSearch size={48} className="text-primary" />
          </div>
        </div>

        <h3 className="text-xl font-bold text-default-900 mb-2">
          Your library is empty
        </h3>
        <p className="text-default-500 max-w-xs mb-8 mx-auto">
          It looks like you haven't added any folders to scan yet. Let's set up
          your library.
        </p>

        <Button
          color="primary"
          variant="shadow"
          startContent={<Settings size={18} />}
          onPress={() => navigate("/settings")}
          className="font-medium"
        >
          Go to Settings
        </Button>
      </div>
    );
  }

  // --- WARIANT 2: Brak wyników wyszukiwania (Standard) ---
  return (
    <div className="flex flex-col items-center justify-center w-full h-[60vh] text-center p-4 animate-appearance-in">
      <div className="relative mb-6">
        <div className="w-24 h-24 bg-default-100 rounded-full flex items-center justify-center">
          <SearchX size={48} className="text-default-400" />
        </div>
        <div className="absolute -bottom-2 -right-2 bg-content1 p-2 rounded-full shadow-lg border border-default-100">
          <FilterX size={20} className="text-danger" />
        </div>
      </div>

      <h3 className="text-xl font-bold text-default-900 mb-2">
        No assets found
      </h3>
      <p className="text-default-500 max-w-xs mb-8 mx-auto">
        Couldn't find any assets matching your current filters. Try adjusting
        your search criteria.
      </p>

      <Button
        color="primary"
        variant="flat"
        startContent={<FilterX size={18} />}
        onPress={resetFilters}
        className="font-medium"
      >
        Clear all filters
      </Button>
    </div>
  );
};
