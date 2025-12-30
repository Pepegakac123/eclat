import { useEffect, useState } from "react";
import { Button, ButtonGroup } from "@heroui/button";
import {
  Dropdown,
  DropdownItem,
  DropdownMenu,
  DropdownTrigger,
} from "@heroui/dropdown";
import { Input } from "@heroui/input";
import { Slider } from "@heroui/slider";
import {
  Grid3X3,
  LayoutList,
  Search,
  ChevronDown,
  ArrowUpAZ,
  ArrowDownAZ,
  Minus,
  Plus,
  RefreshCcw,
  ListFilter,
} from "lucide-react";
import { useGalleryStore, SortOption } from "../stores/useGalleryStore";
import { useShallow } from "zustand/react/shallow";
import { UI_CONFIG } from "@/config/constants";
import { useLocation, useMatch } from "react-router-dom";
import { useMaterialSet } from "@/layouts/sidebar/hooks/useMaterialSets";
import { Skeleton } from "@heroui/skeleton";
import { useAssetsStats } from "../hooks/useAssetsStats";

export const TopToolbar = () => {
  const {
    zoomLevel,
    setZoomLevel,
    viewMode,
    setViewMode,
    sortOption,
    setSortOption,
    sortDesc,
    toggleSortDirection,
    filters,
    setFilters,
    resetFilters,
    pageSize,
    setPageSize,
    filteredCount,
  } = useGalleryStore(
    useShallow((state) => ({
      zoomLevel: state.zoomLevel,
      setZoomLevel: state.setZoomLevel,
      viewMode: state.viewMode,
      setViewMode: state.setViewMode,
      sortOption: state.sortOption,
      setSortOption: state.setSortOption,
      sortDesc: state.sortDesc,
      toggleSortDirection: state.toggleSortDirection,
      filters: state.filters,
      setFilters: state.setFilters,
      resetFilters: state.resetFilters,
      pageSize: state.pageSize,
      setPageSize: state.setPageSize,
      filteredCount: state.filteredCount,
    })),
  );

  const [searchValue, setSearchValue] = useState(filters.searchQuery);
  const location = useLocation();

  // Dane Kolekcji
  const collectionMatch = useMatch("/collections/:id");
  const collectionId = collectionMatch?.params.id;
  const { data: activeCollection, isLoading: isLoadingCollection } =
    useMaterialSet(collectionId);

  const { sidebarStats: stats, isLoading: isLoadingStats } = useAssetsStats();

  // --- LOGIKA TYTUŁU ---
  const getPageTitle = () => {
    // 1. Priorytet: Kolekcja
    if (collectionMatch) {
      if (isLoadingCollection)
        return <Skeleton className="h-7 w-48 rounded-lg" />;
      return activeCollection?.name || "Collection";
    }

    // 2. Reszta stron
    const path = location.pathname;
    if (path.startsWith("/favorites")) return "Favorites";
    if (path.startsWith("/trash")) return "Recycle Bin";
    if (path.startsWith("/uncategorized")) return "Uncategorized";

    return "All Assets";
  };

  // --- LOGIKA LICZNIKA ---
  const getPageCounter = () => {
    let count: number | undefined = 0;
    let loading = false;

    if (collectionMatch) {
      // Jesteśmy w kolekcji
      loading = isLoadingCollection;
      count = activeCollection?.totalAssets; // Pamiętaj o dodaniu tego pola w DTO backendu!
    } else {
      // Jesteśmy w widoku globalnym
      loading = isLoadingStats;
      const path = location.pathname;

      if (path.startsWith("/favorites")) count = stats?.totalFavorites;
      else if (path.startsWith("/trash")) count = stats?.totalTrash;
      else if (path.startsWith("/uncategorized"))
        count = stats?.totalUncategorized;
      else count = stats?.totalAssets; // Default: All Assets
    }

    if (loading) return <Skeleton className="h-6 w-16 rounded-full" />;
    const total = count ?? 0;
    const current = filteredCount ?? total;

    // Jeśli filtrowanie zmniejszyło liczbę wyników, pokaż "Aktualne / Wszystkie"
    // Ale tylko jeśli mamy załadowane statystyki (filteredCount !== null)
    const displayText =
      current !== total && filteredCount !== null
        ? `${current} / ${total}`
        : `${total}`;

    return (
      <span className="rounded-full bg-default-100 px-2.5 py-0.5 text-xs font-medium text-default-500">
        {displayText}
      </span>
    );
  };

  // Debounce dla wyszukiwania
  useEffect(() => {
    const handler = setTimeout(() => {
      if (searchValue !== filters.searchQuery) {
        setFilters({ searchQuery: searchValue });
      }
    }, 400);

    return () => clearTimeout(handler);
  }, [searchValue, setFilters, filters.searchQuery]);

  // Synchronizacja inputa ze storem (gdyby zmienił się z innej strony)
  useEffect(() => {
    setSearchValue(filters.searchQuery);
  }, [filters.searchQuery]);

  // Helper nazewnictwa sortowania
  const getSortLabel = (option: SortOption) => {
    switch (option) {
      case UI_CONFIG.GALLERY.AllowedSortOptions.dateadded:
        return "Date Added";
      case UI_CONFIG.GALLERY.AllowedSortOptions.filename:
        return "File Name";
      case UI_CONFIG.GALLERY.AllowedSortOptions.filesize:
        return "File Size";
      case UI_CONFIG.GALLERY.AllowedSortOptions.lastmodified:
        return "Last Modified";
      default:
        return option;
    }
  };

  return (
    <div className="sticky top-0 z-50 flex h-16 w-full items-center justify-between border-b border-default-200 bg-background/80 px-6 backdrop-blur-md">
      {/* SEKCJA A: TYTUŁ I LICZNIK */}
      <div className="flex items-center gap-4 w-fit">
        <h1 className="text-lg font-bold tracking-tight text-foreground truncate max-w-[300px]">
          {getPageTitle()}
        </h1>
        {getPageCounter()}
      </div>

      {/* SEKCJA B: WYSZUKIWANIE */}
      <div className="flex-1 max-w-xl px-6">
        <Input
          classNames={{
            base: "max-w-full h-10",
            mainWrapper: "h-full",
            input: "text-small",
            inputWrapper:
              "h-full font-normal text-default-500 bg-default-400/20 dark:bg-default-500/20",
          }}
          placeholder="Search by filename..."
          size="sm"
          startContent={<Search size={18} />}
          type="search"
          value={searchValue}
          onValueChange={setSearchValue}
          isClearable
          onClear={() => setSearchValue("")}
        />
      </div>

      {/* SEKCJA C: KONTROLA */}
      <div className="flex items-center gap-4">
        {/* 1. RESET FILTRÓW */}
        <Button
          isIconOnly
          variant="light"
          size="sm"
          onPress={resetFilters}
          title="Reset filters"
        >
          <RefreshCcw size={16} className="text-default-400" />
        </Button>

        <div className="h-6 w-px bg-default-300" />

        {/* 2. ZOOM SLIDER */}
        <div className="flex w-32 xl:w-48 items-center gap-2">
          <Slider
            size="sm"
            step={UI_CONFIG.GALLERY.STEP}
            color="primary"
            maxValue={UI_CONFIG.GALLERY.MAX_ZOOM}
            minValue={UI_CONFIG.GALLERY.MIN_ZOOM}
            aria-label="Thumbnail Size"
            value={zoomLevel}
            onChange={(v) => {
              const val = Array.isArray(v) ? v[0] : v;
              setZoomLevel(val);
            }}
            startContent={
              <button
                type="button"
                className="rounded-full p-1 text-default-400 outline-none transition-colors hover:cursor-pointer hover:text-primary focus-visible:ring-2 focus-visible:ring-primary"
                onClick={() =>
                  setZoomLevel(
                    Math.max(
                      UI_CONFIG.GALLERY.MIN_ZOOM,
                      zoomLevel - UI_CONFIG.GALLERY.STEP,
                    ),
                  )
                }
              >
                <Minus size={14} />
              </button>
            }
            endContent={
              <button
                type="button"
                className="rounded-full p-1 text-default-400 outline-none transition-colors hover:cursor-pointer hover:text-primary focus-visible:ring-2 focus-visible:ring-primary"
                onClick={() =>
                  setZoomLevel(
                    Math.min(
                      UI_CONFIG.GALLERY.MAX_ZOOM,
                      zoomLevel + UI_CONFIG.GALLERY.STEP,
                    ),
                  )
                }
              >
                <Plus size={14} />
              </button>
            }
          />
        </div>

        <div className="h-6 w-px bg-default-300" />

        {/* 3. SORTOWANIE */}
        <div className="flex items-center gap-1">
          <Dropdown>
            <DropdownTrigger>
              <Button
                variant="flat"
                size="sm"
                endContent={<ChevronDown size={16} />}
                className="text-default-600 capitalize min-w-[120px] justify-between hidden sm:flex"
              >
                {getSortLabel(sortOption)}
              </Button>
            </DropdownTrigger>
            <DropdownMenu
              aria-label="Sort options"
              disallowEmptySelection
              selectionMode="single"
              selectedKeys={new Set([sortOption])}
              onSelectionChange={(keys) => {
                const selected = Array.from(keys)[0] as SortOption;
                setSortOption(selected);
              }}
            >
              <DropdownItem
                key={UI_CONFIG.GALLERY.AllowedSortOptions.dateadded}
              >
                Date Added
              </DropdownItem>
              <DropdownItem key={UI_CONFIG.GALLERY.AllowedSortOptions.filename}>
                File Name
              </DropdownItem>
              <DropdownItem key={UI_CONFIG.GALLERY.AllowedSortOptions.filesize}>
                File Size
              </DropdownItem>
              <DropdownItem
                key={UI_CONFIG.GALLERY.AllowedSortOptions.lastmodified}
              >
                Last Modified
              </DropdownItem>
            </DropdownMenu>
          </Dropdown>

          <Button
            isIconOnly
            variant="flat"
            size="sm"
            onPress={toggleSortDirection}
            title={sortDesc ? "Descending" : "Ascending"}
          >
            {sortDesc ? <ArrowDownAZ size={18} /> : <ArrowUpAZ size={18} />}
          </Button>

          {/* PAGE SIZE SELECTOR */}
          <Dropdown>
            <DropdownTrigger>
              <Button
                variant="flat"
                size="sm"
                className="w-fit min-w-[60px]"
                startContent={
                  <ListFilter size={16} className="text-default-500" />
                }
              >
                {pageSize}
              </Button>
            </DropdownTrigger>
            <DropdownMenu
              aria-label="Page Size"
              disallowEmptySelection
              selectionMode="single"
              selectedKeys={new Set([pageSize.toString()])}
              onSelectionChange={(keys) => {
                const val = Number(Array.from(keys)[0]);
                setPageSize(val);
              }}
            >
              {[20, 40, 60, 80, 100].map((size) => (
                <DropdownItem key={size}>{size} items</DropdownItem>
              ))}
            </DropdownMenu>
          </Dropdown>
        </div>

        {/* 4. VIEW TOGGLE */}
        <ButtonGroup variant="flat" size="sm">
          <Button
            isIconOnly
            className={
              viewMode === "grid" ? "bg-default-300 text-foreground" : ""
            }
            onPress={() => setViewMode("grid")}
          >
            <Grid3X3 size={18} />
          </Button>
          <Button
            isIconOnly
            className={
              viewMode === "masonry" ? "bg-default-300 text-foreground" : ""
            }
            onPress={() => setViewMode("masonry")}
          >
            <LayoutList size={18} />
          </Button>
        </ButtonGroup>
      </div>
    </div>
  );
};
