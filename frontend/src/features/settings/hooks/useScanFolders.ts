import { scannerService } from "@/services/scannerService";
import { ScanFolder } from "@/types/api";
import { addToast } from "@heroui/toast";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CheckCircle2, XCircle, FolderPlus, Trash2 } from "lucide-react";

export const useScanFolders = () => {
  const queryClient = useQueryClient();

  const foldersQuery = useQuery({
    queryKey: ["scan-folders"],
    queryFn: scannerService.getFolders,
  });

  const addFolderMutation = useMutation({
    mutationFn: scannerService.addFolder,
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
      const msg = error.response?.data?.message || "Unknown error";
      addToast({
        title: "Action Failed",
        description: msg,
        color: "danger",
        severity: "danger",
        variant: "flat",
      });
    },
  });

  const deleteFolderMutation = useMutation({
    mutationFn: scannerService.deleteFolder,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["scan-folders"] });
      addToast({
        title: "Folder Removed",
        description: "It will no longer be scanned. Rescan will be required.",
        color: "warning",
        severity: "warning",
        variant: "flat",
      });
    },
    onError: (error: any) => {
      addToast({
        title: "Could not delete folder",
        description:
          error.response?.data?.message || "Could not remove the folder.",
        color: "danger",
        severity: "danger",
        variant: "flat",
      });
    },
  });
  const updateStatusMutation = useMutation({
    mutationFn: ({ id, isActive }: { id: number; isActive: boolean }) =>
      scannerService.updateFolderStatus(id, isActive),

    // Dzieje się ZANIM request pójdzie do API
    onMutate: async ({ id, isActive }) => {
      // A. Anulujemy wszelkie odświeżanie w tle, żeby nie nadpisało nam UI
      await queryClient.cancelQueries({ queryKey: ["scan-folders"] });

      // B. Robimy "snapshot" obecnego stanu (na wypadek błędu trzeba cofnąć)
      const previousFolders = queryClient.getQueryData<ScanFolder[]>([
        "scan-folders",
      ]);

      // C. Ręcznie modyfikujemy cache. UI odświeży się NATYCHMIAST.
      queryClient.setQueryData<ScanFolder[]>(["scan-folders"], (old) => {
        if (!old) return [];
        return old.map((folder) =>
          folder.id === id ? { ...folder, isActive: isActive } : folder,
        );
      });

      // Zwracamy kontekst do użycia w razie błędu
      return { previousFolders };
    },
    onError: (err, newVars, context) => {
      // Przywracamy stan sprzed kliknięcia (Rollback)
      if (context?.previousFolders) {
        queryClient.setQueryData(["scan-folders"], context.previousFolders);
      }
      addToast({
        title: "Action Failed",
        description: "Could not update folder status. Changes reverted.",
        color: "danger",
        severity: "danger",
        variant: "flat",
      });
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["scan-folders"] });
    },

    onSuccess: (_, { isActive }) => {
      const action = isActive ? "Activated" : "Deactivated";
      addToast({
        title: `Folder ${action}`,
        description: `Scanning is now ${isActive ? "enabled" : "disabled"} for this path.`,
        color: isActive ? "success" : "warning",
        severity: isActive ? "success" : "warning",
        timeout: 2000,
      });
    },
  });
  const validateMutation = useMutation({
    mutationFn: scannerService.validatePath,
  });
  const startScanMutation = useMutation({
    mutationFn: scannerService.startScan,
    onSuccess: () => {
      addToast({
        title: "Scanner Started",
        description: "The background process has begun.",
        color: "success",
        severity: "success",
        variant: "flat",
      });
      queryClient.invalidateQueries({ queryKey: ["scanner-status"] });
    },
    onError: (error: any) => {
      addToast({
        title: "Scan Failed to Start",
        description:
          error.response?.data?.message || "Is the scanner already running?",
        color: "danger",
        severity: "danger",
        variant: "flat",
      });
    },
  });
  const extensionsQuery = useQuery({
    queryKey: ["allowed-extensions"],
    queryFn: scannerService.getAllowedExtensions,
  });

  const updateExtensionsMutation = useMutation({
    mutationFn: scannerService.updateAllowedExtensions,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["allowed-extensions"] });
      addToast({
        title: "Settings Saved",
        description: "File type filters updated.",
        color: "success",
      });
    },
    onError: (error: any) => {
      addToast({
        title: "Error",
        description:
          error.response?.data?.message || "Failed to update extensions.",
        color: "danger",
      });
    },
  });

  const openExplorerMutation = useMutation({
    mutationFn: scannerService.openInExplorer,
    onError: (error: any) => {
      addToast({
        title: "System Error",
        description:
          error.response?.data?.message || "Could not open explorer.",
        color: "danger",
      });
    },
  });
  return {
    folders: foldersQuery.data,
    isLoading: foldersQuery.isLoading,
    addFolder: addFolderMutation.mutateAsync,
    deleteFolder: deleteFolderMutation.mutateAsync,
    validatePath: validateMutation.mutateAsync,
    isValidating: validateMutation.isPending,
    updateFolderStatus: updateStatusMutation.mutateAsync,
    startScan: startScanMutation.mutateAsync,
    isStartingScan: startScanMutation.isPending,
    extensions: extensionsQuery.data || [],
    isLoadingExtensions: extensionsQuery.isLoading,
    updateExtensions: updateExtensionsMutation.mutateAsync,
    openInExplorer: openExplorerMutation.mutateAsync,
  };
};
