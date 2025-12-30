import {
  GetAll,
  Create,
  Update,
  Delete,
  GetById,
} from "../../../../wailsjs/go/app/MaterialSetService";
import {
  AddAssetToMaterialSet,
  RemoveAssetFromMaterialSet,
} from "../../../../wailsjs/go/app/AssetService";
import { app } from "../../../../wailsjs/go/models";
import { addToast } from "@heroui/toast";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export const useMaterialSets = () => {
  const queryClient = useQueryClient();
  const QUERY_KEY = ["material-sets"];

  const listQuery = useQuery({
    queryKey: QUERY_KEY,
    queryFn: GetAll,
  });

  const createMutation = useMutation({
    mutationFn: (data: app.CreateMaterialSetRequest) => Create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY });
      queryClient.invalidateQueries({ queryKey: ["sidebar-stats"] });
      addToast({
        title: "Success",
        description: "Material Set created successfully.",
        color: "success",
        severity: "success",
        variant: "flat",
      });
    },
    onError: (error: any) => {
      addToast({
        title: "Creation Failed",
        description:
          error.message || "Could not create material set.",
        color: "danger",
        severity: "danger",
        variant: "flat",
      });
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: app.CreateMaterialSetRequest }) =>
      Update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY });
      queryClient.invalidateQueries({ queryKey: ["sidebar-stats"] });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEY, variables.id],
      });
      queryClient.invalidateQueries({
        queryKey: ["material-set", variables.id],
      });

      addToast({
        title: "Updated",
        description: "Material Set updated successfully.",
        color: "success",
        severity: "success",
        variant: "flat",
      });
    },
    onError: (error: any) => {
      addToast({
        title: "Update Failed",
        description:
          error.message || "Could not update material set.",
        color: "danger",
        severity: "danger",
        variant: "flat",
      });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: Delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY });
      queryClient.invalidateQueries({ queryKey: ["sidebar-stats"] });
      addToast({
        title: "Deleted",
        description: "Material Set has been removed.",
        color: "warning", // Warning pasuje do usuwania
        severity: "warning",
        variant: "flat",
      });
    },
    onError: (error: any) => {
      addToast({
        title: "Delete Failed",
        description:
          error.message || "Could not delete material set.",
        color: "danger",
        severity: "danger",
        variant: "flat",
      });
    },
  });
  const addAssetToSetMutation = useMutation({
    mutationFn: ({ setId, assetId }: { setId: number; assetId: number }) =>
      AddAssetToMaterialSet(setId, assetId),

    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["material-set", variables.setId.toString()],
      });
      queryClient.invalidateQueries({ queryKey: QUERY_KEY });
      queryClient.invalidateQueries({ queryKey: ["assets"] });
      queryClient.invalidateQueries({ queryKey: ["asset", variables.assetId] });
      queryClient.invalidateQueries({ queryKey: ["sidebar-stats"] });
    },
    onError: (error: any) => {
      addToast({
        title: "Error",
        description:
          error.message || "Failed to add to collection.",
        color: "danger",
      });
    },
  });
  const removeAssetFromSetMutation = useMutation({
    mutationFn: ({ setId, assetId }: { setId: number; assetId: number }) =>
      RemoveAssetFromMaterialSet(setId, assetId),

    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["material-set", variables.setId.toString()],
      });
      queryClient.invalidateQueries({ queryKey: QUERY_KEY });
      queryClient.invalidateQueries({ queryKey: ["assets"] });
      queryClient.invalidateQueries({ queryKey: ["asset", variables.assetId] });
      queryClient.invalidateQueries({ queryKey: ["sidebar-stats"] });

      addToast({
        title: "Removed",
        description: "Asset removed from collection.",
        color: "default",
        variant: "flat",
      });
    },
    onError: () => {
      addToast({
        title: "Error",
        description: "Failed to remove asset.",
        color: "danger",
      });
    },
  });

  return {
    // Data
    materialSets: listQuery.data || [],
    isLoading: listQuery.isLoading,
    isError: listQuery.isError,

    // Actions
    createMaterialSet: createMutation.mutateAsync,
    isCreating: createMutation.isPending,

    updateMaterialSet: updateMutation.mutateAsync,
    isUpdating: updateMutation.isPending,

    deleteMaterialSet: deleteMutation.mutateAsync,
    isDeleting: deleteMutation.isPending,

    // Asset Operations
    addAssetToSet: addAssetToSetMutation.mutateAsync,
    isAddingAsset: addAssetToSetMutation.isPending,
    removeAssetFromSet: removeAssetFromSetMutation.mutateAsync,
    isRemovingAsset: removeAssetFromSetMutation.isPending,
  };
};

export const useMaterialSet = (id: string | number | null | undefined) => {
  const numericId = typeof id === "string" ? parseInt(id, 10) : id;
  return useQuery({
    queryKey: ["material-set", id],
    queryFn: () => GetById(numericId!),
    enabled: !!numericId,
    staleTime: 1000 * 60 * 5,
  });
};
