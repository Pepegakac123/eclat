import { useState } from "react";
import { Tabs, Tab } from "@heroui/tabs";
import { app } from "@wailsjs/go/models";
import { TabDetails } from "./tabs/TabDetails";
import { TabCollections } from "./tabs/TabCollections";
import { TabVersions } from "./tabs/TabVersions";

export const InspectorTabs = ({ asset }: { asset: app.AssetDetails }) => {
  const [selected, setSelected] = useState("details");

  return (
    <div className="flex-1 flex flex-col w-full">
      <Tabs
        fullWidth
        size="sm"
        variant="underlined"
        aria-label="Asset inspector tabs"
        selectedKey={selected}
        onSelectionChange={(k) => setSelected(k as string)}
        classNames={{
          tabList: "p-0 border-b border-default-100 w-full",
          cursor: "w-full bg-primary h-[2px]",
          tab: "h-9 px-0",
          tabContent:
            "text-tiny font-medium group-data-[selected=true]:text-primary text-default-500",
          panel: "p-4", // Padding dla zawartości
        }}
      >
        <Tab key="details" title="Details">
          <TabDetails asset={asset} />
        </Tab>
        <Tab key="collections" title="Collections">
          <TabCollections asset={asset} />
        </Tab>
        <Tab key="versions" title="Versions">
          <TabVersions />
        </Tab>
      </Tabs>
    </div>
  );
};

