import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { EventsOn } from "../../wailsjs/runtime/runtime";

export const useAssetEvents = () => {
  const queryClient = useQueryClient();

  useEffect(() => {
    const cleanup = EventsOn("assets:changed", (data) => {
      console.log("⚡ Asset change detected via Watcher:", data);
      queryClient.invalidateQueries({ queryKey: ["assets"] });
      queryClient.invalidateQueries({ queryKey: ["sidebar-stats"] });
      queryClient.invalidateQueries({ queryKey: ["folders"] });
    });

    return () => {
      if (cleanup) cleanup();
    };
  }, [queryClient]);
};
