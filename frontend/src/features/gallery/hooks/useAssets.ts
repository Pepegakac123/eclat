import {
  useQuery,
  keepPreviousData,
  useMutation,
  useInfiniteQuery,
} from "@tanstack/react-query";
import { GetAssets, GetAvailableColors } from "../../../../wailsjs/go/app/AssetService";
import { OpenInDefaultApp, OpenInExplorer } from "../../../../wailsjs/go/app/App";
import { app } from "../../../../wailsjs/go/models";
...
  // 2. AKCJA: OTWIERANIE FOLDERU (Mutation)
  const openExplorerMutation = useMutation({
    mutationFn: async (filePath: string) => {
      return OpenInExplorer(filePath);
    },
    onSuccess: () => {
      // Opcjonalnie: Toast sukcesu, ale zazwyczaj okno otwiera się po prostu
      console.log("Explorer opened successfully");
    },
    onError: (error: any) => {
      addToast({
        title: "Błąd Systemu",
        description: "Nie udało się otworzyć folderu.",
        color: "danger",
      });
    },
  });

  const openDefaultAppMutation = useMutation({
    mutationFn: async (filePath: string) => {
      return OpenInDefaultApp(filePath);
    },
    onError: (error: any) => {
      addToast({
        title: "System Error",
        description: "Failed to open file in default application.",
        color: "danger",
      });
    },
  });

  return {
    ...getAssetsQuery,
    openExplorer: openExplorerMutation.mutate,
    openDefaultApp: openDefaultAppMutation.mutate,
  };
};

export const useColors = () => {
  const colorsQuery = useQuery({
    queryKey: ["colors"],
    queryFn: GetAvailableColors,
  });

  return colorsQuery.data;
};
