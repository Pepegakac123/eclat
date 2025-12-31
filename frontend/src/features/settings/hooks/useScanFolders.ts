import { addToast } from "@heroui/toast";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	AddExtensions,
	GetConfig,
	GetPredefinedPalette,
	RemoveExtension,
	StartScan,
} from "@wailsjs/go/scanner/Scanner";
import {
	AddFolder,
	DeleteFolder,
	GetFolders,
	OpenFolderPicker,
	OpenInExplorer,
	UpdateFolderStatus,
	ValidatePath,
} from "@wailsjs/go/settings/SettingsService";

export const useScanFolders = () => {
	const queryClient = useQueryClient();

	const foldersQuery = useQuery({
		queryKey: ["scan-folders"],
		queryFn: GetFolders,
		refetchOnWindowFocus: true,
	});

	const addFolderMutation = useMutation({
		mutationFn: AddFolder,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["scan-folders"] });
		},
	});

	// 3. Usuwanie folderu
	const deleteFolderMutation = useMutation({
		mutationFn: DeleteFolder,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["scan-folders"] });
		},
	});

	const updateStatusMutation = useMutation({
		mutationFn: async ({ id, isActive }: { id: number; isActive: boolean }) => {
			return await UpdateFolderStatus(id, isActive);
		},
		onSuccess: (_, { isActive }) => {
			queryClient.invalidateQueries({ queryKey: ["scan-folders"] });
		},
		onError: (error: any) => {
			addToast({
				title: "Update Failed",
				description: error || "Failed to update folder status",
				color: "danger",
			});
		},
	});

	// Wrapper na ValidatePath (konwersja bool -> object dla UI)
	const validatePathWrapper = async (path: string) => {
		try {
			const isValid = await ValidatePath(path);
			return {
				isValid: isValid,
				message: isValid ? "" : "Path does not exist or is invalid",
			};
		} catch (_e) {
			return { isValid: false, message: "Validation error" };
		}
	};

	// Wrapper na OpenInExplorer (żeby obsłużyć ewentualne błędy w konsoli)
	const openInExplorerWrapper = async (path: string) => {
		try {
			await OpenInExplorer(path);
		} catch (e) {
			console.error("Failed to open explorer:", e);
			addToast({
				title: "Error",
				description: "Could not open explorer",
				color: "warning",
			});
		}
	};

	const startScanMutation = useMutation({
		mutationFn: StartScan,
		onSuccess: (result) => {
			// Tu można dodać obsługę wyniku skanowania, jeśli zwracasz statystyki
			console.log("Scan finished:", result);
			// Po skanowaniu warto odświeżyć listę folderów (np. daty ostatniego skanu)
			queryClient.invalidateQueries({ queryKey: ["scan-folders"] });
			queryClient.invalidateQueries({ queryKey: ["sidebar-stats"] });
			queryClient.invalidateQueries({ queryKey: ["asset"] });
		},
	});

	const extensionsQuery = useQuery({
		queryKey: ["allowed-extensions"],
		queryFn: async () => (await GetConfig()).allowedExtensions || [],
	});

	const addExtensionMutation = useMutation({
		mutationFn: (ext: string) => AddExtensions([ext]),
		onSuccess: (_, ext) => {
			queryClient.invalidateQueries({ queryKey: ["allowed-extensions"] });
			addToast({
				title: "Extension Added",
				description: `Added ${ext} to allowed list`,
				color: "success",
			});
		},
		onError: (error: any) => {
			addToast({
				title: "Error",
				description: error || "Failed to add extension",
				color: "danger",
			});
		},
	});

	const removeExtensionMutation = useMutation({
		mutationFn: RemoveExtension,
		onSuccess: (_, ext) => {
			queryClient.invalidateQueries({ queryKey: ["allowed-extensions"] });
			addToast({
				title: "Extension Removed",
				description: `Removed ${ext} from allowed list`,
				color: "default",
			});
		},
		onError: (error: any) => {
			addToast({
				title: "Error",
				description: error || "Failed to remove extension",
				color: "danger",
			});
		},
	});

	const paletteQuery = useQuery({
		queryKey: ["predefined-palette"],
		queryFn: GetPredefinedPalette,
		staleTime: Infinity,
	});
	const handleOpenPicker = async () => {
		try {
			const path = await OpenFolderPicker();
			return path;
		} catch (e) {
			console.error("Picker error:", e);
			return "";
		}
	};

	return {
		// --- Folders ---
		folders: foldersQuery.data || [],
		isLoading: foldersQuery.isLoading,
		addFolder: addFolderMutation.mutateAsync,
		deleteFolder: deleteFolderMutation.mutateAsync,
		updateFolderStatus: updateStatusMutation.mutateAsync,

		// --- Helpers ---
		validatePath: validatePathWrapper,
		openInExplorer: openInExplorerWrapper,
		openFolderPicker: handleOpenPicker,
		isValidating: false,

		// --- Scanner & Config ---
		startScan: startScanMutation.mutateAsync,
		isStartingScan: startScanMutation.isPending,
		palette: paletteQuery.data || [],
		extensions: extensionsQuery.data || [],
		isLoadingExtensions: extensionsQuery.isLoading,
		addExtension: addExtensionMutation.mutateAsync,
		removeExtension: removeExtensionMutation.mutateAsync,
	};
};
