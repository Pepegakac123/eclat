import { tagService } from "@/services/tagService";
import { useQuery } from "@tanstack/react-query";

export const useTags = () => {
  const listTags = useQuery({
    queryKey: ["tags"],
    queryFn: tagService.getAll,
  });
  return {
    tags: listTags.data,
    isLoading: listTags.isLoading,
  };
};
