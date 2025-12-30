import { GetSidebarStats } from "../../../../wailsjs/go/app/AssetService";
import { useQuery } from "@tanstack/react-query";

export const useAssetsStats = () => {
  const getSidebarStats = useQuery({
    queryKey: ["sidebar-stats"],
    queryFn: GetSidebarStats,
  });
  return {
    sidebarStats: getSidebarStats.data,
    isLoading: getSidebarStats.isLoading,
  };
};
