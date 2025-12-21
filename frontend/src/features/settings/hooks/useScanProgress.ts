import { useState, useEffect } from "react";
import { EventsOn } from "@wailsjs/runtime/runtime";

// Odzwierciedlenie ScanProgressDTO z Go
export interface ScanProgressData {
  current: number;
  total: number;
  lastFile: string;
}

interface ScanState {
  isScanning: boolean;
  progress: number;
  message: string;
  total: number;
  current: number;
}

export const useScanProgress = () => {
  const [scanState, setScanState] = useState<ScanState>({
    isScanning: false,
    progress: 0,
    message: "Idle",
    total: 0,
    current: 0,
  });

  useEffect(() => {
    // 1. Status Skanera
    const stopStatus = EventsOn("scan_status", (status: string) => {
      const isScanning = status === "scanning";
      setScanState((prev) => ({
        ...prev,
        isScanning: isScanning,
        message: isScanning ? "Scanning..." : "Idle",
        progress: isScanning ? prev.progress : 100,
      }));
    });

    // 2. Postęp Skanowania - Odbieramy obiekt ScanProgressData
    const stopProgress = EventsOn("scan_progress", (data: ScanProgressData) => {
      // Obliczanie procentów (zabezpieczenie przed dzieleniem przez 0)
      const percent =
        data.total > 0 ? Math.round((data.current / data.total) * 100) : 0;

      setScanState({
        isScanning: true,
        current: data.current,
        total: data.total,
        message: `Processing: ${data.lastFile}`, // Krótszy tekst
        progress: percent,
      });
    });

    return () => {
      stopStatus();
      stopProgress();
    };
  }, []);

  return scanState;
};
