import { useParams } from "react-router-dom";
import { useGalleryStore } from "../stores/useGalleryStore";
import { AssetCard } from "./AssetCard";
import { Spinner } from "@heroui/spinner";
import { UI_CONFIG } from "@/config/constants";
import { useEffect, useMemo, useRef } from "react";
import { useAssets } from "../hooks/useAssets";
import { useShallow } from "zustand/react/shallow";
import { useAssetsStats } from "../hooks/useAssetsStats";
import { NoResults } from "./NoResults";
import { app } from "../../../../wailsjs/go/models";

type DisplayContentMode =
  keyof typeof UI_CONFIG.GALLERY.AllowedDisplayContentModes;

interface GalleryGridProps {
  mode: DisplayContentMode;
}

export const GalleryGrid = ({ mode }: GalleryGridProps) => {
  const { collectionId } = useParams<{ collectionId: string }>();
  const parsedCollectionId = collectionId ? parseInt(collectionId) : undefined;
  const loadMoreRef = useRef<HTMLDivElement>(null);
  const { sidebarStats } = useAssetsStats();

  const {
    zoomLevel,
    viewMode,
    filters,
    sortOption,
    sortDesc,
    pageSize,
    selectedAssetIds,
    lastSelectedAssetId,
    selectAsset,
    setSelection,
    setFilteredCount,
  } = useGalleryStore(
    useShallow((state) => ({
      zoomLevel: state.zoomLevel,
      viewMode: state.viewMode,
      filters: state.filters,
      sortOption: state.sortOption,
      sortDesc: state.sortDesc,
      pageSize: state.pageSize,
      selectedAssetIds: state.selectedAssetIds,
      lastSelectedAssetId: state.lastSelectedAssetId,
      selectAsset: state.selectAsset,
      setSelection: state.setSelection,
      setFilteredCount: state.setFilteredCount,
    })),
  );

  const assetFilters = useMemo(() => {
    return new app.AssetQueryFilters({
      page: 1,
      pageSize: pageSize,
      searchQuery: filters.searchQuery || "",
      tags: filters.tags || [],
      matchAllTags: filters.matchAllTags,
      fileTypes: filters.fileTypes || [],
      colors: filters.colors || [],
      ratingRange: filters.ratingRange,
      widthRange: filters.widthRange,
      heightRange: filters.heightRange,
      fileSizeRange: filters.fileSizeRange, // Store is MB, Wails expects MB
      dateRange: filters.dateRange,
      hasAlpha: filters.hasAlpha === null ? undefined : filters.hasAlpha,
      onlyFavorites: false, // Override in useAssets based on mode
      isDeleted: false, // Override in useAssets based on mode
      isHidden: false,
      collectionId: parsedCollectionId,
      sortOption: sortOption,
      sortDesc: sortDesc,
    });
  }, [filters, sortOption, sortDesc, pageSize, parsedCollectionId]);

  const {
    data,
    isLoading,
    isError,
    error,
    openExplorer,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useAssets(mode, assetFilters);

  const handleAssetClick = (e: React.MouseEvent, assetId: number) => {
    // 1. SHIFT CLICK (Range Selection)
    if (e.shiftKey && lastSelectedAssetId !== null) {
      const lastIndex = allAssets.findIndex(
        (a) => a.id === lastSelectedAssetId,
      );
      const currentIndex = allAssets.findIndex((a) => a.id === assetId);

      if (lastIndex !== -1 && currentIndex !== -1) {
        const start = Math.min(lastIndex, currentIndex);
        const end = Math.max(lastIndex, currentIndex);

        // kawałek tablicy assetów, które są pomiędzy kliknięciami
        const rangeIds = allAssets.slice(start, end + 1).map((a) => a.id);

        setSelection(rangeIds);
        return;
      }
    }

    // 2. CTRL/CMD CLICK (Multi Toggle)
    const isMulti = e.ctrlKey || e.metaKey;
    selectAsset(assetId, isMulti);
  };

  useEffect(() => {
    if (data?.pages[0]) {
      setFilteredCount(data.pages[0].totalCount);
    }
  }, [data, setFilteredCount]);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        // Jeśli strażnik jest widoczny I mamy następną stronę I nie ładujemy jej teraz
        if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
          fetchNextPage();
        }
      },
      { threshold: 0.1, rootMargin: "200px" }, // Ładuj 200px przed końcem
    );

    if (loadMoreRef.current) {
      observer.observe(loadMoreRef.current);
    }

    return () => observer.disconnect();
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  // 5. Renderowanie Stanów
  if (isLoading) {
    return (
      <div className="flex h-full w-full items-center justify-center">
        <Spinner size="lg" label="Loading assets..." color="primary" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex h-full w-full flex-col items-center justify-center text-danger">
        <p className="text-xl font-bold">Błąd ładowania galerii</p>
        <p className="text-sm opacity-70">
          {(error as Error).message ||
            (error as any)?.response?.data?.message ||
            (error as any)?.response?.data?.error}
        </p>
      </div>
    );
  }

  const allAssets = data?.pages.flatMap((page) => page.items) || [];

  if (allAssets.length === 0) {
    const totalLibraryAssets = sidebarStats?.totalAssets || 0;
    return (
      <NoResults
        variant={totalLibraryAssets === 0 ? "empty-library" : "no-matches"}
      />
    );
  }
  console.log(allAssets[0]);
  return (
    <div className="h-full w-full">
      <div
        style={{ "--col-width": `${zoomLevel}px` } as React.CSSProperties}
        className="grid grid-cols-[repeat(auto-fill,minmax(var(--col-width),1fr))] gap-4 pb-20 p-4 select-none outline-none"
      >
        {allAssets.map((asset) => (
          <div key={asset.id} className="aspect-square">
            <AssetCard
              asset={asset}
              isSelected={selectedAssetIds.has(asset.id)}
              isBulkMode={selectedAssetIds.size > 1}
              onClick={(e) => handleAssetClick(e, asset.id)}
              onDoubleClick={() => openExplorer(asset.filePath)}
              explorerfn={openExplorer}
            />
          </div>
        ))}
      </div>
      {/* Ten element jest na samym dnie. Jak go widać -> fetchNextPage() */}
      <div
        ref={loadMoreRef}
        className="w-full h-20 flex items-center justify-center mt-4"
      >
        {isFetchingNextPage && (
          <Spinner size="md" color="default" label="Loading more..." />
        )}
        {!hasNextPage && allAssets.length > 0 && (
          <p className="text-tiny text-default-400">End of library</p>
        )}
      </div>
    </div>
  );
};
