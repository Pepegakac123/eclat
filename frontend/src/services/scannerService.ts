import { ScanFolder } from "@/types/api";

export const scannerService = {
  // --- Metody skanera ---
  getStatus: async () => {
    return {
      isScanning: false, // Zgodne z useScanProgress
      progress: 0,
      message: "Idle",
      total: 0,
      current: 0,
    };
  },

  startScan: async (): Promise<void> => {
    console.log("Mock start scan");
  },

  // --- Metody folderów (Settings) ---
  getFolders: async (): Promise<ScanFolder[]> => {
    return [];
  },

  addFolder: async (path: string): Promise<ScanFolder> => {
    console.log("Mock add folder:", path);
    // Zwracamy pełny obiekt zgodny z interfejsem ScanFolder
    return {
      id: Math.floor(Math.random() * 1000),
      path,
      isActive: true,
      isDeleted: false,
      lastScanned: new Date().toISOString(),
    };
  },

  deleteFolder: async (id: number): Promise<void> => {
    console.log("Mock delete folder:", id);
  },

  updateFolderStatus: async (
    id: number,
    isActive: boolean,
  ): Promise<ScanFolder> => {
    console.log("Mock update status:", id, isActive);
    return {
      id,
      path: "mock/path",
      isActive,
      isDeleted: false,
    };
  },

  validatePath: async (path: string): Promise<{ isValid: boolean }> => {
    console.log("Mock validate:", path);
    return { isValid: true };
  },

  // --- Metody rozszerzeń ---
  getAllowedExtensions: async (): Promise<string[]> => {
    return [".jpg", ".png", ".obj", ".fbx"]; // Przykładowe dane, żeby UI nie było puste
  },

  updateAllowedExtensions: async (extensions: string[]): Promise<void> => {
    console.log("Mock update extensions:", extensions);
  },

  // --- Systemowe ---
  openInExplorer: async (path: string): Promise<void> => {
    console.log("Mock open explorer:", path);
  },
};
