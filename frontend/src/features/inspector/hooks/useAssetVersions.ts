import { useQuery } from "@tanstack/react-query";
import { GetAssetVersions } from "../../../../wailsjs/go/app/AssetService";

export const useAssetVersions = (assetId: number | null | undefined) => {
  return useQuery({
    queryKey: ["asset-versions", assetId],
    queryFn: () => GetAssetVersions(assetId!),
    enabled: !!assetId,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
};
