import { useQuery } from "@tanstack/react-query";
import { GetAll } from "../../../../wailsjs/go/app/TagService";

export const useTags = () => {
  const listTags = useQuery({
    queryKey: ["tags"],
    queryFn: GetAll,
  });
  return {
    tags: listTags.data,
    isLoading: listTags.isLoading,
  };
};
