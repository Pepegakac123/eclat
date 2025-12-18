import {
  GetFolders,
  AddFolder,
  DeleteFolder,
  ValidatePath,
  UpdateFolderStatus,
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

  // --- QUERY: Foldery ---
  const foldersQuery = useQuery({
    queryKey: ["scan-folders"],
    queryFn: GetFolders,
  });

  // --- MUTATION: Dodaj Folder ---
  const addFolderMutation = useMutation({
    mutationFn: AddFolder,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["scan-folders"] });
      addToast({
        title: "Success",
        description: "New folder linked to library.",
        color: "success",
        severity: "success",
        variant: "flat",
        timeout: 3000,
      });
    },
    onError: (error: any) => {
      addToast({
        title: "Action Failed",
        description: error || "Could not add folder",
        color: "danger",
        severity: "danger",
        variant: "flat",
      });
    },
  });

  // --- MUTATION: Usuń Folder ---
  const deleteFolderMutation = useMutation({
    mutationFn: DeleteFolder,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["scan-folders"] });
    },
    onError: (error: any) => {
      addToast({
        title: "Error",
        description: error || "Delete failed",
        color: "danger",
        severity: "danger",
        variant: "flat",
      });
    },
  });

  // --- MUTATION: Status (Switch) ---
  const updateStatusMutation = useMutation({
    mutationFn: ({ id, isActive }: { id: number; isActive: boolean }) =>
      UpdateFolderStatus(id, isActive),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["scan-folders"] });
    },
    onError: (error: any) => {
      addToast({ title: "Error", description: error, color: "danger" });
    },
  });

  // --- MUTATION: Start Skanowania ---
  const startScanMutation = useMutation({
    mutationFn: StartScan,
    onError: (error: any) => {
      addToast({ title: "Scan Error", description: error, color: "danger" });
    },
  });

  // --- QUERY & MUTATION: Extensions ---
  const extensionsQuery = useQuery({
    queryKey: ["allowed-extensions"],
    queryFn: async () => {
      const config = await GetConfig();
      return config.allowedExtensions || [];
    },
  });

  const addExtensionMutation = useMutation({
    mutationFn: (ext: string) => AddExtensions([ext]),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["allowed-extensions"] });
      addToast({
        title: "Saved",
        description: "Extension added",
        color: "success",
      });
    },
    onError: (err: any) => {
      addToast({ title: "Error", description: err, color: "danger" });
    },
  });

  const removeExtensionMutation = useMutation({
    mutationFn: RemoveExtension,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["allowed-extensions"] });
      addToast({
        title: "Removed",
        description: "Extension removed",
        color: "warning",
      });
    },
  });

  // --- Wrapper dla Validate Path (UI oczekuje Promise<{isValid}>) ---
  const validatePathWrapper = async (path: string) => {
    const isValid = await ValidatePath(path);
    return { isValid };
  };

  const paletteQuery = useQuery({
    queryKey: ["predefined-palette"],
    queryFn: GetPredefinedPalette, // Zwraca Promise<PaletteColor[]>
    staleTime: Infinity, // To się nigdy nie zmienia
  });

  return {
    folders: foldersQuery.data || [],
    isLoading: foldersQuery.isLoading,
    addFolder: addFolderMutation.mutateAsync,
    deleteFolder: deleteFolderMutation.mutateAsync,
    updateFolderStatus: updateStatusMutation.mutateAsync,
    validatePath: validatePathWrapper,
    isValidating: false, // Wails działa lokalnie błyskawicznie

    startScan: startScanMutation.mutateAsync,
    isStartingScan: startScanMutation.isPending,
    palette: paletteQuery.data || [],
    extensions: extensionsQuery.data || [],
    isLoadingExtensions: extensionsQuery.isLoading,
    addExtension: addExtensionMutation.mutateAsync,
    removeExtension: removeExtensionMutation.mutateAsync,

    openInExplorer: OpenInExplorer,
  };
};
