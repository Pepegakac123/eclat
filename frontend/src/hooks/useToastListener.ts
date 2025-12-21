// hooks/useToastListener.ts
import { useEffect } from "react";
import { EventsOn } from "@wailsjs/runtime/runtime";
import { addToast } from "@heroui/toast";

export interface BackendToast {
  type: "info" | "success" | "warning" | "error";
  title: string;
  message: string;
}
type HeroUIColor = "success" | "danger" | "warning" | "primary" | "default";

export const useToastListener = () => {
  useEffect(() => {
    const stopListener = EventsOn("toast", (data: BackendToast) => {
      // Adapter / Mapper
      // Zamieniamy "info" na "primary", "error" na "danger"
      const colorMap: Record<string, HeroUIColor> = {
        info: "primary",
        success: "success",
        warning: "warning",
        error: "danger",
      };

      addToast({
        title: data.title,
        description: data.message,
        color: colorMap[data.type] || "default",
        timeout: 4000,
      });
    });

    return () => stopListener();
  }, []);
};
