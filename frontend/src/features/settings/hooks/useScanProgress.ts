import { useEffect, useState } from "react";
import * as signalR from "@microsoft/signalr";
import apiReq from "@/lib/axios";
import { useQueryClient } from "@tanstack/react-query";
import { API_BASE_URL } from "@/config/constants";

interface ScanState {
  isScanning: boolean;
  progress: number;
  message: string;
  total: number;
  current: number;
}

// export const useScanProgress = () => {
//   const queryClient = useQueryClient();
//   const [scanState, setScanState] = useState<ScanState>({
//     isScanning: false, // Domyślnie false, ale zaraz to sprawdzimy
//     progress: 0,
//     message: "",
//     total: 0,
//     current: 0,
//   });

//   useEffect(() => {
//     const fetchInitialStatus = async () => {
//       try {
//         const { data } = await apiReq.get<{ isScanning: boolean }>(
//           "/scanner/status",
//         );
//         if (data.isScanning) {
//           setScanState((prev) => ({
//             ...prev,
//             isScanning: true,
//             message: "Resuming scanner connection...",
//           }));
//         }
//       } catch (err) {
//         console.error("Failed to fetch scanner status", err);
//       }
//     };

//     fetchInitialStatus();
//   }, []);

//   // 2. SIGNALR (Nasłuchiwanie)
//   useEffect(() => {
//     const cleanBaseUrl = API_BASE_URL.replace(/\/$/, "");
//     const hubUrl = `${cleanBaseUrl}/hubs/scan`;
//     const connection = new signalR.HubConnectionBuilder()
//       .withUrl(hubUrl)
//       .configureLogging(signalR.LogLevel.Error) // Mniej logów w konsoli
//       .withAutomaticReconnect()
//       .build();

//     // HANDLER 1: Tylko zmiana stanu (Start/Stop)
//     connection.on("ReceiveScanStatus", (status: string) => {
//       const isNowScanning = status === "scanning";
//       if (!isNowScanning) {
//         console.log(
//           "✅ Scan finished (Idle signal received). Refreshing data...",
//         );
//         queryClient.invalidateQueries({ queryKey: ["assets"] });
//         queryClient.invalidateQueries({ queryKey: ["sidebar-stats"] });
//         queryClient.invalidateQueries({ queryKey: ["colors"] });
//       }
//       setScanState((prev) => {
//         // Logika "Detect Edge": Jeśli wcześniej skanował, a teraz przestał -> SUKCES
//         // if (prev.isScanning && !isNowScanning) {
//         //    addToast({
//         //      title: "Scan Finished",
//         //      description: "Library has been updated.",
//         //      color: "success",
//         //    });
//         // }
//         return {
//           ...prev,
//           isScanning: isNowScanning,
//           // Jeśli startujemy, resetujemy progress. Jeśli kończymy, zostawiamy 100% na chwilę
//           progress: isNowScanning ? 0 : 100,
//         };
//       });
//     });

//     // HANDLER 2: Tylko aktualizacja paska postępu
//     connection.on(
//       "ReceiveProgress",
//       (msg: string, total: number, current: number) => {
//         setScanState((prev) => ({
//           ...prev,
//           message: msg,
//           total: total,
//           current: current,
//           progress: total > 0 ? (current / total) * 100 : 0,
//         }));
//       },
//     );

//     connection.start().catch((err) => console.error("SignalR Error:", err));

//     return () => {
//       connection.stop();
//     };
//   }, []);

//   return scanState;
// };

// Tymczasowy mock dla Wailsa
export const useScanProgress = () => {
  // Na razie zwracamy stan "nie skanuje"
  return {
    isScanning: false,
    progress: 0,
    message: "",
    total: 0,
    current: 0,
  };
};
