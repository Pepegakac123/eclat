import {
  useQuery,
  keepPreviousData,
  useMutation,
  useInfiniteQuery,
} from "@tanstack/react-query";
import { GetAssets, GetAvailableColors } from "../../../../wailsjs/go/app/AssetService";
import { OpenInDefaultApp, OpenInExplorer } from "../../../../wailsjs/go/app/App";
import { app } from "../../../../wailsjs/go/models";
import { UI_CONFIG } from "@/config/constants";
import { addToast } from "@heroui/toast";

type GalleryMode = keyof typeof UI_CONFIG.GALLERY.AllowedDisplayContentModes;

export const useAssets = (
  mode: GalleryMode,
  filters: app.AssetQueryFilters,
) => {
  const getAssetsQuery = useInfiniteQuery({
    queryKey: ["assets", mode, filters],
    initialPageParam: 1,
    queryFn: async ({ pageParam = 1 }) => {
      // Create a copy of filters to avoid mutating the original object
      const currentFilters = new app.AssetQueryFilters({ ...filters });
      currentFilters.page = pageParam as number;

      // Apply mode-specific overrides
      switch (mode) {
        case UI_CONFIG.GALLERY.AllowedDisplayContentModes.favorites:
          currentFilters.onlyFavorites = true;
          currentFilters.isDeleted = false;
          break;
        case UI_CONFIG.GALLERY.AllowedDisplayContentModes.trash:
          currentFilters.isDeleted = true;
          break;
        case UI_CONFIG.GALLERY.AllowedDisplayContentModes.collection:
          currentFilters.isDeleted = false;
          break;
        case UI_CONFIG.GALLERY.AllowedDisplayContentModes.uncategorized:
          currentFilters.isDeleted = false;
          currentFilters.onlyUncategorized = true;
          break;
        case UI_CONFIG.GALLERY.AllowedDisplayContentModes.hidden:
          currentFilters.isDeleted = false;
          currentFilters.isHidden = true;
          break;
        default:
          currentFilters.isDeleted = false;
          break;
      }

      return GetAssets(currentFilters);
    },
    getNextPageParam: (lastPage) => {
      // Calculate next page based on total count and page size
      if (
        lastPage.items &&
        lastPage.items.length > 0 &&
        lastPage.page * lastPage.pageSize < lastPage.totalCount
      ) {
        return lastPage.page + 1;
      }
      return undefined;
    },
    placeholderData: keepPreviousData,
    staleTime: 1000 * 60 * 1,
  });

  // 2. AKCJA: OTWIERANIE FOLDERU (Mutation)
  const openExplorerMutation = useMutation({
    mutationFn: async (filePath: string) => {
      return OpenInExplorer(filePath);
    },
    onSuccess: () => {
      // Opcjonalnie: Toast sukcesu, ale zazwyczaj okno otwiera się po prostu
      console.log("Explorer opened successfully");
    },
    onError: (error: any) => {
      addToast({
        title: "Błąd Systemu",
        description: "Nie udało się otworzyć folderu.",
        color: "danger",
      });
    },
  });

  const openDefaultAppMutation = useMutation({
    mutationFn: async (filePath: string) => {
      return OpenInDefaultApp(filePath);
    },
    onError: (error: any) => {
      addToast({
        title: "System Error",
        description: "Failed to open file in default application.",
        color: "danger",
      });
    },
  });

  return {
    ...getAssetsQuery,
    openExplorer: openExplorerMutation.mutate,
    openDefaultApp: openDefaultAppMutation.mutate,
  };
};

export const useColors = () => {
  const colorsQuery = useQuery({
    queryKey: ["colors"],
    queryFn: GetAvailableColors,
  });

  return colorsQuery.data;
};