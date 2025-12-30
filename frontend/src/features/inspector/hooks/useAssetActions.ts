import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ToggleAssetFavorite } from "../../../../wailsjs/go/app/AssetService";
import {
  OpenInExplorer,
  OpenInDefaultApp,
} from "../../../../wailsjs/go/app/App";
import { app } from "../../../../wailsjs/go/models";
import { addToast } from "@heroui/toast";

export const useAssetActions = (assetId: number) => {
  const queryClient = useQueryClient();

  //  TOGGLE FAVORITE
  const favoriteMutation = useMutation({
    mutationFn: () => ToggleAssetFavorite(assetId),

    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: ["asset", assetId] });

      const previousAsset = queryClient.getQueryData<app.AssetDetails>([
        "asset",
        assetId,
      ]);
      if (previousAsset) {
        const newAsset = new app.AssetDetails({
          ...previousAsset,
          isFavorite: !previousAsset.isFavorite,
        });
        queryClient.setQueryData<app.AssetDetails>(
          ["asset", assetId],
          newAsset,
        );
      }

      return { previousAsset };
    },

    onError: (_err, _vars, context) => {
      if (context?.previousAsset) {
        queryClient.setQueryData(["asset", assetId], context.previousAsset);
      }
      addToast({
        title: "Error",
        description: "Failed to toggle favorite",
        color: "danger",
      });
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["asset", assetId] });
      queryClient.invalidateQueries({ queryKey: ["assets"] });
      queryClient.invalidateQueries({ queryKey: ["favorites"] });
      queryClient.invalidateQueries({ queryKey: ["sidebar-stats"] });
    },
  });

  // 2. OPEN IN EXPLORER
  const explorerMutation = useMutation({
    mutationFn: (path: string) => OpenInExplorer(path),
    onError: () => {
      addToast({
        title: "System Error",
        description: "Could not open explorer.",
        color: "danger",
      });
    },
  });

  // 3. OPEN IN PROGRAM (Default App)
  const programMutation = useMutation({
    mutationFn: (path: string) => OpenInDefaultApp(path),
    onError: () => {
      addToast({
        title: "System Error",
        description: "Could not open file.",
        color: "danger",
      });
    },
    onSuccess: () => {
      addToast({
        title: "Success",
        description: "File is opening it may take a second.",
        color: "success",
      });
    },
  });

  return {
    toggleFavorite: favoriteMutation.mutate,
    openInExplorer: explorerMutation.mutate,
    openInProgram: programMutation.mutate,
    isTogglingFav: favoriteMutation.isPending,
  };
};
