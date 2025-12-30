import { useMutation, useQueryClient, useQuery } from "@tanstack/react-query";
import {
  GetAssetById,
  UpdateAssetMetadata,
} from "../../../../wailsjs/go/app/AssetService";
import { app } from "../../../../wailsjs/go/models";
import { addToast } from "@heroui/toast";

export const useAsset = (assetId: number | null | undefined) => {
  return useQuery({
    queryKey: ["asset", assetId],
    queryFn: () => GetAssetById(assetId!),
    // Fetchuj tylko jak mamy ID
    enabled: !!assetId,
    staleTime: 1000 * 60 * 5, // 5 minut
  });
};

export const useAssetMutation = (assetId: number) => {
  const queryClient = useQueryClient();

  const patchMutation = useMutation({
    mutationFn: (updates: Partial<app.AssetDetails>) => {
      const req = new app.UpdateAssetRequest({
        Description: updates.description,
        Rating: updates.rating,
        IsFavorite: updates.isFavorite,
      });
      return UpdateAssetMetadata(assetId, req);
    },

    onMutate: async (updates) => {
      await queryClient.cancelQueries({ queryKey: ["asset", assetId] });
      const previousAsset = queryClient.getQueryData<app.AssetDetails>([
        "asset",
        assetId,
      ]);

      if (previousAsset) {
        // Create a new instance or copy to avoid mutation issues, though strict mode warns
        // Wails classes have methods, so spreading might lose them if not careful,
        // but for data object usually fine. Ideally use new app.AssetDetails({...prev, ...updates})
        // But updates is Partial, so simple spread works for properties.
        const newAsset = new app.AssetDetails({
          ...previousAsset,
          ...updates,
        });
        queryClient.setQueryData<app.AssetDetails>(
          ["asset", assetId],
          newAsset,
        );
      }

      return { previousAsset };
    },

    onError: (_err, _updates, context) => {
      if (context?.previousAsset) {
        queryClient.setQueryData(["asset", assetId], context.previousAsset);
      }
      addToast({
        title: "Update Failed",
        description: "Could not update asset properties.",
        color: "danger",
      });
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["asset", assetId] });
      queryClient.invalidateQueries({ queryKey: ["assets"] });
    },
  });

  return {
    patch: patchMutation.mutate,
    isUpdating: patchMutation.isPending,
  };
};
