import { addToast } from "@heroui/toast";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
	OpenInDefaultApp,
	OpenInExplorer,
} from "../../../../wailsjs/go/app/App";
import {
	DeleteAssetsPermanently,
	RenameAsset,
	SetAssetHidden,
	ToggleAssetFavorite,
} from "../../../../wailsjs/go/app/AssetService";
import { app } from "../../../../wailsjs/go/models";

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

	// 2. TOGGLE HIDDEN
	const hiddenMutation = useMutation({
		mutationFn: (isHidden: boolean) => SetAssetHidden(assetId, isHidden),

		onMutate: async (newHiddenStatus) => {
			await queryClient.cancelQueries({ queryKey: ["asset", assetId] });

			const previousAsset = queryClient.getQueryData<app.AssetDetails>([
				"asset",
				assetId,
			]);
			if (previousAsset) {
				const newAsset = new app.AssetDetails({
					...previousAsset,
					isHidden: newHiddenStatus,
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
				description: "Failed to update hidden status",
				color: "danger",
			});
		},

		onSuccess: (_data, newHiddenStatus) => {
			addToast({
				title: newHiddenStatus ? "Hidden" : "Unhidden",
				description: newHiddenStatus
					? "Asset has been hidden"
					: "Asset is visible again",
				color: "success",
			});
			queryClient.invalidateQueries({ queryKey: ["assets"] });
			queryClient.invalidateQueries({ queryKey: ["sidebar-stats"] });
		},
	});

	// 3. OPEN IN EXPLORER
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

	// 4. OPEN IN PROGRAM (Default App)
	const programMutation = useMutation({
		mutationFn: (path: string) => OpenInDefaultApp(path),
		onError: () => {
			addToast({
				title: "System Error",
				description: "Could not open file.",
				color: "danger",
			});
		},
	});

	// 5. DELETE ASSET
	const deleteMutation = useMutation({
		mutationFn: () => DeleteAssetsPermanently([assetId]),
		onSuccess: () => {
			addToast({
				title: "Success",
				description: "Asset deleted permanently",
				color: "success",
			});
			queryClient.invalidateQueries({ queryKey: ["assets"] });
			queryClient.invalidateQueries({ queryKey: ["sidebar-stats"] });
		},
		onError: (err) => {
			addToast({
				title: "Error",
				description: "Failed to delete asset: " + err,
				color: "danger",
			});
		},
	});

	// 6. RENAME ASSET
	const renameMutation = useMutation({
		mutationFn: (newName: string) => RenameAsset(assetId, newName),
		onSuccess: () => {
			addToast({
				title: "Success",
				description: "Asset renamed successfully",
				color: "success",
			});
			queryClient.invalidateQueries({ queryKey: ["assets"] });
			queryClient.invalidateQueries({ queryKey: ["asset", assetId] });
		},
		onError: (err) => {
			addToast({
				title: "Error",
				description: "Failed to rename asset: " + err,
				color: "danger",
			});
		},
	});

	return {
		toggleFavorite: favoriteMutation.mutate,
		toggleHidden: hiddenMutation.mutate,
		openInExplorer: explorerMutation.mutate,
		openInProgram: programMutation.mutate,
		deleteAsset: deleteMutation.mutate,
		renameAsset: renameMutation.mutate,
		isTogglingFav: favoriteMutation.isPending,
		isTogglingHidden: hiddenMutation.isPending,
		isDeleting: deleteMutation.isPending,
		isRenaming: renameMutation.isPending,
	};
};
