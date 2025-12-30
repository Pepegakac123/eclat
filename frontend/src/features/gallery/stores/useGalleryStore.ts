import { create } from "zustand";
import { devtools } from "zustand/middleware";

export interface GalleryFilters {
  searchQuery: string; // C#: FileName
  tags: string[]; // C#: Tags
  matchAllTags: boolean; // C#: MatchAll
  fileTypes: string[]; // C#: FileType
  colors: string[]; // C#: DominantColors

  ratingRange: [number, number]; // C#: RatingMin, RatingMax (Slider UI zwraca tablicę)
  dateRange: {
    // C#: DateFrom, DateTo
    from: string | null;
    to: string | null;
  };

  // Wymiary (Opcjonalne)
  widthRange: [number, number]; // C#: MinWidth, MaxWidth
  heightRange: [number, number]; // C#: MinHeight, MaxHeight
  fileSizeRange: [number, number];
  hasAlpha: boolean | null; // C#: HasAlphaChannel (null = wszystko, true = z, false = bez)
}

export type SortOption = "dateadded" | "filename" | "filesize" | "lastmodified";

interface GalleryState {
  // --- UI State ---
  zoomLevel: number;
  viewMode: "grid" | "masonry";
  pageSize: number;

  // --- Data State ---
  filters: GalleryFilters;
  sortOption: SortOption;
  sortDesc: boolean; // C#: SortDesc

  // Stany Assetów
  selectedAssetIds: Set<number>;
  lastSelectedAssetId: number | null;
  filteredCount: number | null;

  // --- Actions ---
  setZoomLevel: (zoom: number) => void;
  setViewMode: (mode: "grid" | "masonry") => void;
  setPageSize: (size: number) => void;
  setFilteredCount: (count: number | null) => void;

  // Update filtrów (Partial pozwala aktualizować tylko jedno pole np. tylko tags)
  setFilters: (newFilters: Partial<GalleryFilters>) => void;

  // Helpersy do sortowania
  setSortOption: (option: SortOption) => void;
  toggleSortDirection: () => void;

  // Reset
  resetFilters: () => void;

  selectAsset: (id: number, multi: boolean) => void;
  setSelection: (ids: number[]) => void;
  clearSelection: () => void;
}

const DEFAULT_FILTERS: GalleryFilters = {
  searchQuery: "",
  tags: [],
  matchAllTags: true,
  fileTypes: [],
  colors: [],
  ratingRange: [0, 5],
  dateRange: { from: null, to: null },
  widthRange: [0, 8192],
  heightRange: [0, 8192],
  fileSizeRange: [0, 4096],
  hasAlpha: null,
};

export const useGalleryStore = create<GalleryState>()(
  devtools((set) => ({
    // Initial State
    zoomLevel: 250,
    viewMode: "grid",
    pageSize: 20,
    filters: DEFAULT_FILTERS,
    sortOption: "dateadded", // Default z C#
    sortDesc: true, // Default z C# (OrderByDescending)
    selectedAssetIds: new Set<number>(),
    lastSelectedAssetId: null,
    filteredCount: null,

    // Actions
    setZoomLevel: (zoom) => set({ zoomLevel: zoom }),
    setViewMode: (mode) => set({ viewMode: mode }),
    setPageSize: (size) => set({ pageSize: size }),
    setFilteredCount: (count) => set({ filteredCount: count }),

    setFilters: (newFilters) =>
      set((state) => ({
        filters: { ...state.filters, ...newFilters },
      })),

    setSortOption: (option) => set({ sortOption: option }),

    toggleSortDirection: () => set((state) => ({ sortDesc: !state.sortDesc })),

    resetFilters: () => set({ filters: DEFAULT_FILTERS }),

    selectAsset: (id, multi) => {
      if (multi) {
        set((state) => ({
          selectedAssetIds: new Set([...state.selectedAssetIds, id]),
          lastSelectedAssetId: id,
        }));
      } else {
        set({ selectedAssetIds: new Set([id]), lastSelectedAssetId: id });
      }
    },

    setSelection: (ids) =>
      set({
        selectedAssetIds: new Set(ids),
        lastSelectedAssetId: ids[0] || null,
      }),

    clearSelection: () =>
      set({ selectedAssetIds: new Set(), lastSelectedAssetId: null }),
  })),
);
