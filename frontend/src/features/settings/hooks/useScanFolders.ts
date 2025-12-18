import {
  GetFolders,
  AddFolder,
  DeleteFolder,
  UpdateFolderStatus,
  ValidatePath,
  OpenInExplorer,
} from "@wailsjs/go/services/SettingsService";

import {
  GetConfig,
  AddExtensions,
  RemoveExtension,
  GetPredefinedPalette,
  StartScan,
} from "@wailsjs/go/services/Scanner";

import { addToast } from "@heroui/toast";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export const useScanFolders = () => {
  const queryClient = useQueryClient();

  // =========================================================
  // --- REAL BACKEND: FOLDERS (DATABASE) ---
  // =========================================================

  // 1. Pobieranie listy folderów
  const foldersQuery = useQuery({
    queryKey: ["scan-folders"],
    queryFn: GetFolders,
    // Opcjonalnie: odświeżaj co jakiś czas lub przy focusie okna
    refetchOnWindowFocus: true,
  });

  // 2. Dodawanie folderu
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
    onError: (err: any) => {
      addToast({
        title: "Error",
        description:
          err instanceof Error ? err.message : "Failed to add folder",
        color: "danger",
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
    onError: (err: any) => {
      addToast({
        title: "Error",
        description: "Could not delete folder",
        color: "danger",
      });
    },
  });

  // 4. Zmiana statusu (Active/Inactive)
  // Wails generuje funkcję przyjmującą osobne argumenty, a useMutation przyjmuje jeden obiekt.
  // Musimy to owinąć.
  const updateStatusMutation = useMutation({
    mutationFn: async ({ id, isActive }: { id: number; isActive: boolean }) => {
      return await UpdateFolderStatus(id, isActive);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["scan-folders"] });
    },
    onError: (err: any) => {
      addToast({
        title: "Error",
        description: "Could not update status",
        color: "danger",
      });
    },
  });

  // =========================================================
  // --- REAL BACKEND: HELPER METHODS ---
  // =========================================================

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

  // =========================================================
  // --- REAL BACKEND: SCANNER & CONFIG ---
  // =========================================================

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
