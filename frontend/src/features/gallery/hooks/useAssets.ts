import {
  useQuery,
  keepPreviousData,
  useMutation,
  useInfiniteQuery,
} from "@tanstack/react-query";
import { assetService } from "@/services/assetService";
import { AssetQueryParams } from "@/types/api";
import { UI_CONFIG } from "@/config/constants";
import { addToast } from "@heroui/toast";

type GalleryMode = keyof typeof UI_CONFIG.GALLERY.AllowedDisplayContentModes;

export const useAssets = (
  mode: GalleryMode,
  params: AssetQueryParams,
  collectionId?: number,
) => {
  const getAssetsQuery = useInfiniteQuery({
    queryKey: ["assets", mode, params, collectionId],
    initialPageParam: 1,
    queryFn: async ({ pageParam = 1 }) => {
      const currentParams = { ...params, pageNumber: pageParam as number };
      switch (mode) {
        case UI_CONFIG.GALLERY.AllowedDisplayContentModes.favorites:
          return assetService.getFavorites(currentParams);
        case UI_CONFIG.GALLERY.AllowedDisplayContentModes.trash:
          return assetService.getTrashed(currentParams);
        case UI_CONFIG.GALLERY.AllowedDisplayContentModes.collection:
          if (!collectionId) throw new Error("Brak ID kolekcji");
          return assetService.getAssetsForMaterialSet(
            collectionId,
            currentParams,
          );
        case UI_CONFIG.GALLERY.AllowedDisplayContentModes.uncategorized:
          return assetService.getUncategorizedAssets(currentParams); // TODO: Dodać filtr uncategorized
        default:
          return assetService.getAll(currentParams);
      }
    },
    getNextPageParam: (lastPage) => {
      if (lastPage.hasNextPage) {
        return lastPage.currentPage + 1;
      }
      return undefined;
    },
    placeholderData: keepPreviousData,
    enabled: mode === "collection" ? !!collectionId : true,
    staleTime: 1000 * 60 * 1,
  });

  // 2. AKCJA: OTWIERANIE FOLDERU (Mutation)
  const openExplorerMutation = useMutation({
    mutationFn: (filePath: string) => assetService.openInExplorer(filePath),
    onSuccess: () => {
      // Opcjonalnie: Toast sukcesu, ale zazwyczaj okno otwiera się po prostu
      console.log("Explorer opened successfully");
    },
    onError: (error: any) => {
      addToast({
        title: "Błąd Systemu",
        description:
          error.response?.data?.message || "Nie udało się otworzyć folderu.",
        color: "danger",
      });
    },
  });

  return {
    ...getAssetsQuery,
    openExplorer: openExplorerMutation.mutate,
  };
};
export const useColors = () => {
  const colorsQuery = useQuery({
    queryKey: ["colors"],
    queryFn: assetService.getColorsList,
  });

  return colorsQuery.data;
};
