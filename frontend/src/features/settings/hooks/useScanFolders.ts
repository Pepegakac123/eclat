import {
  GetFolders,
  AddFolder,
  DeleteFolder,
  UpdateFolderStatus,
  ValidatePath,
  OpenInExplorer,
  OpenFolderPicker,
} from "@wailsjs/go/settings/SettingsService";

import {
  GetConfig,
  AddExtensions,
  RemoveExtension,
  GetPredefinedPalette,
  StartScan,
} from "@wailsjs/go/scanner/Scanner";

import { addToast } from "@heroui/toast";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

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
      addToast({
        title: "Success",
        description: "Folder added to library",
        color: "success",
      });
    },
  });

  // 3. Usuwanie folderu
  const deleteFolderMutation = useMutation({
    mutationFn: DeleteFolder,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["scan-folders"] });
      addToast({
        title: "Deleted",
        description: "Folder removed from library",
        color: "default",
      });
    },
  });

  const updateStatusMutation = useMutation({
    mutationFn: async ({ id, isActive }: { id: number; isActive: boolean }) => {
      return await UpdateFolderStatus(id, isActive);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["scan-folders"] });
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
    } catch (e) {
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
    },
  });

  const extensionsQuery = useQuery({
    queryKey: ["allowed-extensions"],
    queryFn: async () => (await GetConfig()).allowedExtensions || [],
  });

  const addExtensionMutation = useMutation({
    mutationFn: (ext: string) => AddExtensions([ext]),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["allowed-extensions"] }),
  });

  const removeExtensionMutation = useMutation({
    mutationFn: RemoveExtension,
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["allowed-extensions"] }),
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
