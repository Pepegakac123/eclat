// Zamiast enuma:
export const UI_CONFIG = {
  GALLERY: {
    DEFAULT_ZOOM: 220,
    MIN_ZOOM: 220,
    MAX_ZOOM: 420,
    STEP: 50,
    AllowedDisplayContentModes: {
      default: "default",
      favorites: "favorites",
      uncategorized: "uncategorized",
      trash: "trash",
      collection: "collection",
      hidden: "hidden",
    },
    AllowedSortOptions: {
      dateadded: "dateadded",
      filename: "filename",
      filesize: "filesize",
      lastmodified: "lastmodified",
    },
    FilterOptions: {
      MAX_DIMENSION: 8160,
      MAX_FILE_SIZE: 4096,
    },
  },
  QUERY: {
    STALE_TIME: 1000 * 60 * 5, // 5 minut
    GC_TIME: 1000 * 60 * 15, // 15 minut
    RETRY_COUNT: 3,
  },
} as const;

export const PRESET_COLORS = [
  "#000000", // Black
  "#FFFFFF", // White
  "#808080", // Gray
  "#C0C0C0", // Silver
  "#FF0000", // Red
  "#800000", // Maroon
  "#FFFF00", // Yellow
  "#808000", // Olive
  "#00FF00", // Lime
  "#008000", // Green
  "#00FFFF", // Cyan
  "#008080", // Teal
  "#0000FF", // Blue
  "#000080", // Navy
  "#FF00FF", // Magenta
  "#800080", // Purple
];
export const ALLOWED_FILE_TYPES = ["model", "image", "texture", "other"];
export const BYTES_IN_MB = 1024 * 1024;
export const MAX_MB = 4096;
export const API_BASE_URL =
  (import.meta.env.DEV
    ? import.meta.env.VITE_API_URL_DEV
    : import.meta.env.VITE_API_URL_PROD) || "";
