import { useMutation, useQueryClient, useQuery } from "@tanstack/react-query";
import { assetService } from "@/services/assetService";
import { Asset } from "@/types/api";
import { addToast } from "@heroui/toast";

export const useAsset = (assetId: number | null | undefined) => {
  return useQuery({
    queryKey: ["asset", assetId],
    queryFn: () => assetService.getById(assetId!),
    // Fetchuj tylko jak mamy ID
    enabled: !!assetId,
    staleTime: 1000 * 60 * 5, // 5 minut
  });
};

export const useAssetMutation = (assetId: number) => {
  const queryClient = useQueryClient();

  const patchMutation = useMutation({
    mutationFn: (updates: Partial<Asset>) =>
      assetService.patch(assetId, updates),

    onMutate: async (updates) => {
      await queryClient.cancelQueries({ queryKey: ["asset", assetId] });
      const previousAsset = queryClient.getQueryData<Asset>(["asset", assetId]);

      if (previousAsset) {
        queryClient.setQueryData<Asset>(["asset", assetId], {
          ...previousAsset,
          ...updates,
        });
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
