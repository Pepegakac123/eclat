import { useQuery } from "@tanstack/react-query";
import { GetLibraryStats } from "../../../../wailsjs/go/app/AssetService";

export const useLibraryStats = () => {
  return useQuery({
    queryKey: ["library-stats"],
    queryFn: GetLibraryStats,
    refetchOnWindowFocus: true,
  });
};
