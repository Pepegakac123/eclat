import React, { useState } from "react";
import { Button } from "@heroui/button";
import { ScrollShadow } from "@heroui/scroll-shadow";
import { Tabs, Tab } from "@heroui/tabs";
import { Textarea } from "@heroui/input";
import { Divider } from "@heroui/divider";
import { Tooltip } from "@heroui/tooltip";
import { Chip } from "@heroui/chip";
import { Spinner } from "@heroui/spinner";
import {
  Heart,
  ExternalLink,
  Layers,
  BoxSelect,
  FileText,
  Tag as TagIcon,
  FolderOpen,
  XCircle,
} from "lucide-react";
import { useGalleryStore } from "@/features/gallery/stores/useGalleryStore";
import { useAsset } from "../hooks/useAsset";
// IMPORT NOWEGO KOMPONENTU
import { InspectorHeader } from "./single/InspectorHeader";
import { InspectorProperties } from "./single/InspectorProperties";
import { InspectorTags } from "./single/InspectorTags";
import { InspectorTabs } from "./single/InspectorTabs";
import { InspectorFooter } from "./single/InspectorFooter";
import { InspectorLibraryStats } from "./InspectorLibraryStats";

// --- HELPER: SECTION HEADER ---
const SectionHeader = ({
  title,
  icon: Icon,
}: {
  title: string;
  icon?: any;
}) => (
  <div className="flex items-center gap-2 text-default-500 mb-2 px-1">
    {Icon && <Icon size={14} />}
    <span className="text-xs font-semibold uppercase tracking-wider">
      {title}
    </span>
  </div>
);

// --- HELPER: DETAIL ROW ---
const DetailRow = ({
  label,
  value,
}: {
  label: string;
  value: string | number;
}) => (
  <div className="flex justify-between py-1.5 border-b border-default-100/50 last:border-0">
    <span className="text-tiny text-default-500">{label}</span>
    <span className="text-tiny text-default-900 font-medium truncate max-w-[60%]">
      {value}
    </span>
  </div>
);

export const InspectorPanel = () => {
  const [activeTab, setActiveTab] = useState<string>("details");

  const selectedAssetIds = useGalleryStore((state) => state.selectedAssetIds);
  const selectionCount = selectedAssetIds.size;
  const isMultiSelect = selectionCount > 1;

  const activeAssetId =
    selectionCount === 1 ? selectedAssetIds.values().next().value : null;

  const { data: asset, isLoading, isError } = useAsset(activeAssetId);

  // --- BRAK SELEKCJI ---
  if (selectionCount === 0) {
    return <InspectorLibraryStats />;
  }

  // --- MULTI SELECT ---
  if (isMultiSelect) {
    return (
      <div className="h-full w-full flex flex-col bg-content1">
        <div className="p-6 border-b border-default-100 flex flex-col items-center justify-center gap-3 bg-default-50/50">
          <div className="w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center text-primary shadow-sm">
            <BoxSelect size={24} />
          </div>
          <div className="text-center">
            <h3 className="text-md font-bold text-default-900">
              {selectionCount} items
            </h3>
            <p className="text-tiny text-default-500">Batch selection active</p>
          </div>
        </div>
        <div className="p-8 text-center text-default-400 text-tiny">
          Batch editing coming soon...
        </div>
      </div>
    );
  }

  // --- LOADING ---
  if (isLoading) {
    return (
      <div className="h-full w-full flex items-center justify-center">
        <Spinner size="lg" color="primary" />
      </div>
    );
  }

  // --- ERROR ---
  if (isError || !asset) {
    return (
      <div className="h-full w-full flex items-center justify-center text-danger">
        <p>Failed to load asset details.</p>
      </div>
    );
  }

  // --- SINGLE ASSET RENDER ---
  return (
    <div className="h-full w-full flex flex-col bg-content1">
      {/* 1. Header */} {/* SEKCJA EDITABLE WŁAŚCIWOŚCI */}
      <InspectorHeader asset={asset} />
      {/* 2. SCROLLABLE CONTENT */}
      <ScrollShadow className="flex-1 min-h-0 flex flex-col custom-scrollbar">
        {/* SEKCJA NON-EDITABLE WŁAŚCIWOŚCI */}
        <InspectorProperties asset={asset} />
        <Divider className="opacity-50" />
        {/* SEKCJA TAGÓW */}
        <InspectorTags asset={asset} />
        <Divider className="opacity-50" />

        {/* TABS */}
        <InspectorTabs asset={asset} />
        <Divider className="opacity-50" />
      </ScrollShadow>
      {/* 4. STICKY FOOTER ACTIONS */}
      <InspectorFooter asset={asset} />
    </div>
  );
};
