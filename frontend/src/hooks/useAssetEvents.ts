import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { EventsOn } from "../../wailsjs/runtime/runtime";

export const useAssetEvents = () => {
  const queryClient = useQueryClient();

  useEffect(() => {
    const cleanupAssets = EventsOn("assets:changed", (data) => {
      console.log("⚡ Asset change detected:", data);
      queryClient.invalidateQueries({ queryKey: ["assets"] });
      queryClient.invalidateQueries({ queryKey: ["sidebar-stats"] });
      queryClient.invalidateQueries({ queryKey: ["folders"] });
      queryClient.invalidateQueries({ queryKey: ["library-stats"] });
      queryClient.invalidateQueries({ queryKey: ["colors"] });
    });

    const cleanupStatus = EventsOn("scan_status", (status) => {
      if (status === "idle") {
        queryClient.invalidateQueries({ queryKey: ["library-stats"] });
      }
    });

    return () => {
      if (cleanupAssets) cleanupAssets();
      if (cleanupStatus) cleanupStatus();
    };
  }, [queryClient]);
};
