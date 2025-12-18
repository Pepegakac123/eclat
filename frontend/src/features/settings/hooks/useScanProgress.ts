import { useState, useEffect } from "react";
import { EventsOn } from "@wailsjs/runtime/runtime";

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
    // 1. Status Skanera (Start/Stop)
    const stopStatus = EventsOn("scan_status", (status: string) => {
      const isScanning = status === "scanning";
      setScanState((prev) => ({
        ...prev,
        isScanning: isScanning,
        message: isScanning ? "Scanning..." : "Idle",
        progress: isScanning ? prev.progress : 100, // 100% na koniec
      }));
    });

    // 2. Postęp Skanowania (Co 10 plików)
    const stopProgress = EventsOn("scan_progress", (data: any) => {
      // data = {current, total, lastFile}
      const percent = data.total > 0 ? (data.current / data.total) * 100 : 0;
      setScanState({
        isScanning: true,
        current: data.current,
        total: data.total,
        message: `Scanning: ${data.lastFile}`,
        progress: percent,
      });
    });

    // Cleanup przy odmontowaniu komponentu
    return () => {
      stopStatus();
      stopProgress();
    };
  }, []);

  return scanState;
};
