import { assetService } from "@/services/assetService";
import { useQuery } from "@tanstack/react-query";

export const useAssetsStats = () => {
  const getSidebarStats = useQuery({
    queryKey: ["sidebar-stats"],
    queryFn: () => assetService.getSidebarStats(),
  });
  return {
    sidebarStats: getSidebarStats.data,
    isLoading: getSidebarStats.isLoading,
  };
};
