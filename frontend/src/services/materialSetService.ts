import {
  CreateMaterialSetRequest,
  MaterialSet,
  UpdateMaterialSetRequest, // Teraz to zadziała
} from "@/types/api";

export const materialSetService = {
  // ZMIANA: Zwracamy tablicę, nie PagedResponse
  getAll: async (): Promise<MaterialSet[]> => {
    return [];
  },

  getById: async (id: string): Promise<MaterialSet> => {
    return {
      id: Number(id),
      name: "Mock Collection",
      assets: [],
      dateAdded: new Date().toISOString(), // Zgodne z interfejsem
      lastModified: new Date().toISOString(),
      totalAssets: 0,
      isDeleted: false,
    } as MaterialSet;
  },

  create: async (data: CreateMaterialSetRequest): Promise<MaterialSet> => {
    console.log("Mock create:", data);
    return {
      id: Math.floor(Math.random() * 1000),
      name: data.name,
      description: data.description,
      totalAssets: 0,
      dateAdded: new Date().toISOString(),
      lastModified: new Date().toISOString(),
      isDeleted: false,
      assets: [],
    } as MaterialSet;
  },

  update: async (id: string, data: MaterialSet): Promise<MaterialSet> => {
    console.log("Mock update:", id, data);
    return data;
  },

  delete: async (id: string): Promise<void> => {
    console.log("Mock delete:", id);
  },

  // Te metody były wywoływane, ale brakowało ich w interfejsie hooka w pewnym momencie
  // Dodajemy je dla pewności
  addAssetToSet: async (setId: number, assetId: number): Promise<void> => {
    console.log("Mock add asset:", setId, assetId);
  },
  removeAssetFromSet: async (setId: number, assetId: string): Promise<void> => {
    console.log("Mock remove asset:", setId, assetId);
  },
};
