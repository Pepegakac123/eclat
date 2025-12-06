// frontend/src/services/assetService.ts

import apiReq from "@/lib/axios";
import {
  Asset,
  AssetQueryParams,
  PagedResponse,
  SidebarStats,
} from "@/types/api";

// Helper do pustej odpowiedzi stronicowanej
const emptyPagedResponse = <T>(): PagedResponse<T> => ({
  items: [],
  totalItems: 0,
  currentPage: 1,
  pageSize: 20,
  totalPages: 0,
  hasNextPage: false,
  hasPreviousPage: false,
});

export const assetService = {
  getAll: async (params: AssetQueryParams): Promise<PagedResponse<Asset>> => {
    // MOCK DLA WAILSA
    return emptyPagedResponse<Asset>();
    // const response = await apiReq.get("/assets", { params });
    // return response.data;
  },
  getFavorites: async (
    params: AssetQueryParams,
  ): Promise<PagedResponse<Asset>> => {
    return emptyPagedResponse<Asset>();
    // const response = await apiReq.get("/assets/favorites", { params });
    // return response.data;
  },
  getTrashed: async (
    params: AssetQueryParams,
  ): Promise<PagedResponse<Asset>> => {
    return emptyPagedResponse<Asset>();
    // const response = await apiReq.get("/assets/deleted", { params });
    // return response.data;
  },
  getUncategorizedAssets: async (
    params: AssetQueryParams,
  ): Promise<PagedResponse<Asset>> => {
    return emptyPagedResponse<Asset>();
    // const response = await apiReq.get("/assets/uncategorized", { params });
    // return response.data;
  },
  getColorsList: async (): Promise<string[]> => {
    // MOCK - naprawia błąd w SidebarFilters
    return [];
    // const response = await apiReq.get("/assets/colors");
    // return response.data;
  },
  getSidebarStats: async (): Promise<SidebarStats> => {
    return {
      totalAssets: 0,
      totalFavorites: 0,
      totalUncategorized: 0,
      totalTrashed: 0,
    };
  },

  // ... reszta metod (getById, patch, updateTags, etc.) może zostać bez zmian na razie,
  // bo one są wywoływane dopiero po kliknięciu w asset (którego nie ma).
  // Ale upewnij się, że nie ma błędów składniowych w reszcie pliku.
  getById: async (id: number): Promise<Asset> => {
    const response = await apiReq.get(`/assets/${id}`);
    return response.data;
  },
  patch: async (id: number, updates: Partial<Asset>): Promise<Asset> => {
    const response = await apiReq.patch<Asset>(`/assets/${id}`, updates);
    return response.data;
  },
  updateTags: async (id: number, tagsNames: string[]): Promise<void> => {
    await apiReq.post(`/assets/${id}/tags`, { tagsNames });
  },
  getAssetsForMaterialSet: async (
    setId: number,
    params: AssetQueryParams,
  ): Promise<PagedResponse<Asset>> => {
    return emptyPagedResponse<Asset>(); // To też warto zaślepić
    // const response = await apiReq.get(`/materialsets/${setId}/assets`, { params });
    // return response.data;
  },
  addAssetToMaterialSet: async (
    setId: number,
    assetId: number,
  ): Promise<void> => {
    await apiReq.post(`/materialsets/${setId}/assets/${assetId}`);
  },
  removeAssetFromMaterialSet: async (
    setId: number,
    assetId: string,
  ): Promise<void> => {
    await apiReq.delete(`/materialsets/${setId}/assets/${assetId}`);
  },
  toggleFavorite: async (id: number): Promise<void> => {
    await apiReq.patch(`/assets/${id}/toggle-favorite`);
  },
  openInExplorer: async (path: string) => {
    return apiReq.post("/system/open-in-explorer", { path });
  },
  openInProgram: async (path: string) => {
    return apiReq.post("/system/open-in-program", { path });
  },
};
