// To musi odpowiadać strukturze ToastField z Go
export interface ToastMessage {
  type: "info" | "success" | "warning" | "error"; // Enum zamiast string
  title: string;
  message: string;
}
