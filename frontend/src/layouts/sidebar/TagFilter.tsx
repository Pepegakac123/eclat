import { Autocomplete, AutocompleteItem } from "@heroui/autocomplete";
import { Chip } from "@heroui/chip";
import { Switch } from "@heroui/switch";
import { Tooltip } from "@heroui/tooltip";
import { Search, Hash, Info } from "lucide-react";
import { useGalleryStore } from "../../features/gallery/stores/useGalleryStore";
import { useMemo } from "react";
import { useTags } from "./hooks/useTags";
import { Skeleton } from "@heroui/skeleton";

export const TagFilter = () => {
  const filters = useGalleryStore((state) => state.filters);
  const setFilters = useGalleryStore((state) => state.setFilters);

  const selectedTags = filters.tags;
  const isStrictMode = filters.matchAllTags;

  const { tags: apiTags, isLoading } = useTags();

  const handleToggleTag = (tag: string) => {
    const newTags = selectedTags.includes(tag)
      ? selectedTags.filter((t) => t !== tag)
      : [...selectedTags, tag];

    setFilters({ tags: newTags });
  };

  const handleStrictModeChange = (isSelected: boolean) => {
    setFilters({ matchAllTags: isSelected });
  };

  const filteredAutocompleteItems = useMemo(() => {
    if (!apiTags || !Array.isArray(apiTags)) return [];
    return apiTags
      .filter((tag) => !selectedTags.includes(tag.name))
      .map((tag) => ({
        label: tag.name,
        value: tag.name,
      }));
  }, [apiTags, selectedTags]);
  // 10 tagów jako "Popular" (backend zwraca je posortowane)
  const popularTagsList = useMemo(() => {
    if (!apiTags || !Array.isArray(apiTags)) return [];
    return apiTags.slice(0, 10);
  }, [apiTags]);

  return (
    <div className="px-2 mb-6">
      <div className="flex items-center justify-between mb-3 px-1">
        <div className="flex items-center gap-2">
          <span className="text-[10px] font-semibold uppercase tracking-[0.15em] text-default-400/80">
            Filter Logic
          </span>
          <Tooltip
            content="If enabled, shows only assets containing ALL selected tags (AND logic). Otherwise shows assets with ANY of the tags (OR logic)."
            className="max-w-xs text-tiny"
          >
            <Info
              size={14}
              className="text-default-400/70 cursor-help hover:text-foreground transition-colors"
            />
          </Tooltip>
        </div>
        <Switch
          size="sm"
          isSelected={isStrictMode}
          onValueChange={handleStrictModeChange}
          aria-label="Strict Mode"
          color="primary"
        >
          <span className="text-[10px] text-default-500 font-medium">
            {isStrictMode ? "STRICT" : "LOOSE"}
          </span>
        </Switch>
      </div>

      <Autocomplete
        aria-label="Filter tags"
        placeholder="Add tag filter..."
        size="sm"
        variant="flat"
        radius="md"
        startContent={<Search size={16} className="text-default-400" />}
        items={filteredAutocompleteItems}
        isLoading={isLoading}
        selectedKey={null}
        isVirtualized={true}
        maxListboxHeight={250}
        itemHeight={32}
        onSelectionChange={(key) => {
          if (key) {
            handleToggleTag(key.toString());
          }
        }}
        classNames={{
          popoverContent: "bg-content1 border border-default-200 mb-6",
        }}
        inputProps={{
          classNames: {
            input: "text-sm",
            inputWrapper:
              "h-9 min-h-0 bg-default-100 group-data-[focus=true]:bg-default-200",
          },
        }}
      >
        {(item) => (
          <AutocompleteItem key={item.value} textValue={item.label}>
            <div className="flex items-center justify-between gap-2">
              <div className="flex items-center gap-2">
                <Hash size={14} className="text-default-400" />
                <span>{item.label}</span>
              </div>
            </div>
          </AutocompleteItem>
        )}
      </Autocomplete>

      {/* 3. SELECTED TAGS (ACTIVE FILTER CHIPS) */}
      {selectedTags.length > 0 && (
        <div className="flex flex-wrap gap-2 mb-6 mt-6 animate-appearance-in">
          {selectedTags.map((tag) => (
            <Chip
              key={tag}
              onClose={() => handleToggleTag(tag)}
              variant="solid"
              color="primary"
              size="sm"
              classNames={{
                base: "h-6 bg-primary shadow-md shadow-primary/20",
                content:
                  "text-[10px] font-bold px-1 text-white uppercase tracking-wider",
                closeButton: "text-white/70 hover:text-white transition-colors",
              }}
            >
              {tag}
            </Chip>
          ))}
          {/* Przycisk Clear All */}
          <button
            type="button"
            onClick={() => setFilters({ tags: [] })}
            className="text-[10px] text-default-400 hover:text-danger transition-colors underline decoration-dotted underline-offset-2 ml-1"
          >
            Clear all
          </button>
        </div>
      )}

      {/* 4. POPULAR TAGS (CLOUD) */}
      <div>
        <p className="text-[12px] font-semibold uppercase tracking-[0.15em] text-default-400/80 mb-3 mt-6 px-1">
          Popular Tags
        </p>

        {isLoading ? (
          <div className="flex flex-wrap gap-2">
            {[1, 2, 3].map((i) => (
              <Skeleton
                key={i}
                className="h-6 w-16 rounded-md bg-default-200/50"
              />
            ))}
          </div>
        ) : popularTagsList.length > 0 ? (
          <div className="flex flex-wrap gap-1.5">
            {popularTagsList.map((tag) => {
              const isActive = selectedTags.includes(tag.name);

              return (
                <Chip
                  key={tag.id}
                  size="sm"
                  variant="flat"
                  className={`
                          cursor-pointer transition-all duration-200 select-none border
                          ${
                            isActive
                              ? "bg-primary/10 border-primary/40 text-primary font-semibold shadow-[0_0_10px_rgba(var(--heroui-primary),0.15)]"
                              : "bg-default-100/50 border-transparent text-default-500 hover:bg-default-200 hover:text-default-700 hover:border-default-300"
                          }
                        `}
                  classNames={{
                    base: "h-6 px-0",
                    content: "text-[12px] px-2 flex items-center gap-1",
                  }}
                  onClick={() => handleToggleTag(tag.name)}
                >
                  {!isActive && <Hash size={10} className="opacity-40" />}
                  {tag.name}
                </Chip>
              );
            })}
          </div>
        ) : (
          <div className="flex items-center gap-2 px-1 py-4 text-default-400 border border-dashed border-default-200 rounded-lg justify-center bg-default-50/50">
            <Info size={14} className="opacity-50" />
            <span className="text-[10px] opacity-70">
              Library has no tags yet.
            </span>
          </div>
        )}
      </div>
    </div>
  );
};
