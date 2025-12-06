import { Outlet } from "react-router-dom";
import { TopToolbar } from "@/features/gallery/components/TopToolbar";
import { InspectorPanel } from "@/features/inspector/components/InspectorPanel";
import { Sidebar } from "./sidebar/Sidebar";

export const MainLayout = () => {
  return (
    <div
      className="grid h-screen w-screen overflow-hidden bg-background text-foreground
              grid-cols-[300px_1fr_320px] grid-rows-[64px_1fr]"
    >
      {/* 1. LEWY SIDEBAR pełna wysokość) */}
      <aside className="row-span-2 h-full">
        <Sidebar />
      </aside>

      {/* 2. TOP TOOLBAR */}
      <header className="col-start-2 col-end-4 z-20">
        <TopToolbar />
      </header>

      {/* 3. MAIN CONTENT  */}
      <main className="relative col-start-2 col-end-3 row-start-2 overflow-hidden bg-background">
        <div className="h-full w-full overflow-y-auto p-6 custom-scrollbar">
          <Outlet />
        </div>
      </main>

      {/* 4. INSPECTOR (Prawy panel) */}
      {}
      <aside className="col-start-3 col-end-4 row-start-2 border-l border-default-200 bg-content1 h-full overflow-hidden relative flex flex-col">
        <InspectorPanel />
      </aside>
    </div>
  );
};
