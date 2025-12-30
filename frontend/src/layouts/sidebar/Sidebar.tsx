import { Button } from "@heroui/button";
import { ScrollShadow } from "@heroui/scroll-shadow";
import {
  Home,
  Settings,
  Layers,
  Heart,
  Box,
  Shapes,
  EyeOff,
} from "lucide-react";
import { SidebarSection } from "./SidebarSection";
import { SidebarItem } from "./SidebarItem";
import { TagFilter } from "./TagFilter";
import { Skeleton } from "@heroui/skeleton";
import { useNavigate } from "react-router-dom";
import { CircularProgress } from "@heroui/progress";
import { useScanProgress } from "@/features/settings/hooks/useScanProgress";
import { useAssets } from "@/features/gallery/hooks/useAssets";
import { useAssetsStats } from "@/features/gallery/hooks/useAssetsStats";
import { SidebarCollections } from "./SidebarCollections";
import { SidebarFilters } from "@/features/gallery/components/SidebarFilters";
/*


TODO: [FEATURE] Smart Collections & Drag-Drop
- UI Guidelines (Sekcja 6.2): Dodać sekcję "Smart Collections" (zapisane filtry z bazy).
- UI Guidelines (Sekcja 6.2): Obsłużyć Drag & Drop - przeciąganie assetu z Gridu na nazwę Kolekcji w Sidebarze powinno go do niej dodać.

*/

export const Sidebar = () => {
  const navigate = useNavigate();
  const { isScanning, progress } = useScanProgress();
  const { sidebarStats } = useAssetsStats();
  return (
    <aside className="h-full w-full flex flex-col border-r border-default-200 bg-content1/50 backdrop-blur-xl">
      {/* LOGO */}
      <div className="flex h-16 flex-shrink-0 items-center px-5 border-b border-default-100/50">
        <div className="flex items-center gap-3 text-xl font-bold tracking-tight select-none">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary ring-1 ring-primary/20">
            <Layers size={20} />
          </div>
          <span className="text-foreground font-mono text-xl font-extrabold tracking-tight">
            Eclat
          </span>
        </div>
      </div>

      {/* NAVIGATION LIST */}
      <ScrollShadow className="flex-1 py-4 px-3 custom-scrollbar">
        {/* LIBRARY */}
        <SidebarSection title="Library">
          <SidebarItem
            icon={Home}
            label="All Assets"
            to="/"
            count={sidebarStats?.totalAssets || 0}
          />
          <SidebarItem
            icon={Heart}
            label="Favorites"
            to="/favorites"
            count={sidebarStats?.totalFavorites || 0}
          />
          <SidebarItem
            icon={Box}
            label="Uncategorized"
            to="/uncategorized"
            count={sidebarStats?.totalUncategorized || 0}
          />
          <SidebarItem
            icon={EyeOff}
            label="Hidden"
            to="/hidden"
            count={sidebarStats?.totalHidden || 0}
          />
        </SidebarSection>

        {/* COLLECTIONS */}
        <SidebarCollections />

        {/* TAGS */}
        <SidebarSection title="Tags">
          <TagFilter />
        </SidebarSection>

        {/* Filters */}
        <SidebarSection title="Filters">
          <SidebarFilters />
        </SidebarSection>
      </ScrollShadow>

      {/* FOOTER */}
      <div className="flex-shrink-0 border-t border-default-200 p-3 bg-content1 flex items-center gap-2">
        <Button
          variant="light"
          // Zmieniamy w-full na flex-1, żeby przycisk zajął dostępne miejsce, ale zostawił trochę dla kółka
          className="flex-1 justify-start gap-3 px-3 text-default-500 hover:text-foreground"
          startContent={<Settings size={18} />}
          onPress={() => navigate("/settings")}
        >
          Settings
        </Button>

        {/* Renderujemy Kółko tylko gdy skanuje */}
        {isScanning && (
          <div className="flex flex-col items-center justify-center pr-1">
            <CircularProgress
              aria-label="Scanning..."
              size="sm" // Małe kółko
              value={progress}
              color="primary"
              showValueLabel={true} // Pokaże % w środku kółka
              classNames={{
                svg: "w-8 h-8 drop-shadow-md",
                indicator: "stroke-primary",
                track: "stroke-default-300/20",
                value: "text-[10px] font-semibold text-default-500",
              }}
            />
          </div>
        )}
      </div>
    </aside>
  );
};
