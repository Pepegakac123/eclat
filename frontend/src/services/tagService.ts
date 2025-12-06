import apiReq from "@/lib/axios";
import { Tag } from "@/types/api";

export const tagService = {
  getAll: async (): Promise<Tag[]> => {
    // const response = await apiReq.get("/tags");
    // return response.data;
    // MOCK DLA WAILSA (Tymczasowy)
    return [];
  },
};
