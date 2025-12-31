import React, { useState, useEffect } from "react";
import { Spinner } from "@heroui/spinner";
import { Database, FileImage, Clock, Server } from "lucide-react";
import { useLibraryStats } from "../hooks/useLibraryStats";
import { GetAppVersion } from "@wailsjs/go/app/App";

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

const StatItem = ({
  icon: Icon,
  label,
  value,
  subValue,
}: {
  icon: any;
  label: string;
  value: string | number;
  subValue?: string;
}) => (
  <div className="flex items-center gap-4 p-4 rounded-xl bg-default-50 border border-default-100/50 hover:bg-default-100 transition-colors">
    <div className="p-3 rounded-lg bg-primary/10 text-primary">
      <Icon size={24} />
    </div>
    <div className="flex flex-col">
      <span className="text-tiny text-default-500 uppercase font-bold tracking-wider">
        {label}
      </span>
      <span className="text-xl font-bold text-default-900">{value}</span>
      {subValue && (
        <span className="text-tiny text-default-400">{subValue}</span>
      )}
    </div>
  </div>
);

export const InspectorLibraryStats = () => {
  const { data: stats, isLoading, isError } = useLibraryStats();
  const [version, setVersion] = useState<string>("...");

  useEffect(() => {
    GetAppVersion().then(setVersion).catch(console.error);
  }, []);

  console.log(stats);
  if (isLoading) {
    return (
      <div className="h-full w-full flex items-center justify-center">
        <Spinner size="lg" color="primary" />
      </div>
    );
  }

  if (isError || !stats) {
    return (
      <div className="h-full w-full flex flex-col items-center justify-center text-default-300 gap-4 p-8 text-center">
        <Server size={48} className="opacity-20" />
        <p>Library stats unavailable</p>
      </div>
    );
  }

  const lastScanDate = stats.lastScan
    ? new Date(stats.lastScan).toLocaleDateString()
    : "Never";
  const lastScanTime = stats.lastScan
    ? new Date(stats.lastScan).toLocaleTimeString()
    : "";

  return (
    <div className="h-full w-full flex flex-col p-6 gap-6 overflow-y-auto">
      <div className="flex flex-col gap-2 mb-4">
        <h2 className="text-2xl font-bold text-default-900">Library Stats</h2>
        <p className="text-small text-default-500">
          Overview of your asset database
        </p>
      </div>

      <div className="flex flex-col gap-3">
        <StatItem
          icon={FileImage}
          label="Total Assets"
          value={stats.totalAssets.toLocaleString()}
          subValue="Indexed files"
        />

        <StatItem
          icon={Database}
          label="Storage Used"
          value={formatFileSize(stats.totalSize)}
          subValue="Total file size"
        />

        <StatItem
          icon={Clock}
          label="Last Scan"
          value={lastScanDate}
          subValue={lastScanTime}
        />
      </div>

      <div className="mt-auto p-4 rounded-xl bg-gradient-to-br from-default-100 to-default-50 border border-default-200">
        <p className="text-tiny text-center text-default-500">
          Eclat Asset Manager v{version}
        </p>
      </div>
    </div>
  );
};

