import { UI_CONFIG } from '@/config/constants';
import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: UI_CONFIG.QUERY.STALE_TIME, 
      gcTime: UI_CONFIG.QUERY.GC_TIME, // lifetime danych w pamięci
      retry: UI_CONFIG.QUERY.STALE_TIME,
      refetchOnWindowFocus: false,
    },
  },
});