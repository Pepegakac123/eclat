import { UI_CONFIG } from "@/config/constants";
import { addToast } from "@heroui/toast";
import { MutationCache, QueryCache, QueryClient } from "@tanstack/react-query";
// Funkcja pomocnicza do wyciągania treści błędu
const getErrorMessage = (error: any) => {
  if (typeof error === "string") return error;
  if (error instanceof Error) return error.message;
  return "An unexpected error occurred";
};
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: UI_CONFIG.QUERY.STALE_TIME,
      gcTime: UI_CONFIG.QUERY.GC_TIME, // lifetime danych w pamięci
      retry: UI_CONFIG.QUERY.STALE_TIME,
      refetchOnWindowFocus: false,
    },
  },
  // Obsługa błędów przy pobieraniu danych (np. lista folderów)
  queryCache: new QueryCache({
    onError: (error) => {
      console.error("Global Query Error:", error);
      addToast({
        title: "Data Fetch Error",
        description: getErrorMessage(error),
        color: "danger",
      });
    },
  }),
  // Obsługa błędów przy zmianach danych (np. dodawanie rozszerzenia)
  mutationCache: new MutationCache({
    onError: (error: any) => {
      // Ignorujemy błędy, które obsłużyliśmy ręcznie (opcjonalne)
      if (error.isHandled) return;

      console.error("Global Mutation Error:", error);
      addToast({
        title: "Operation Failed",
        description: getErrorMessage(error),
        color: "danger",
      });
    },
  }),
});
