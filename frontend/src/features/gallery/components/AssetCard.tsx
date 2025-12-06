import { Card, CardFooter, CardHeader } from "@heroui/card";
import { Image } from "@heroui/image";
import { Checkbox } from "@heroui/checkbox";
import { Button } from "@heroui/button";
import {
  Dropdown,
  DropdownTrigger,
  DropdownMenu,
  DropdownItem,
} from "@heroui/dropdown";
import {
  Heart,
  MoreVertical,
  Edit,
  FolderPlus,
  Trash,
  Box,
  Image as ImageIcon,
  FileBox,
  FolderOpen,
  Maximize2,
  HardDrive,
} from "lucide-react";
import { useState } from "react";
import { Asset } from "@/types/api";
import { AxiosResponse } from "axios";
import { UseMutateFunction } from "@tanstack/react-query";
import { useAssetActions } from "@/features/inspector/hooks/useAssetActions";
import { API_BASE_URL } from "@/config/constants";

interface AssetCardProps {
  asset: Asset;
  isSelected: boolean;
  isBulkMode: boolean;
  onClick: (e: React.MouseEvent) => void;
  onDoubleClick: () => void;
  explorerfn: UseMutateFunction<
    AxiosResponse<any, any, {}>,
    any,
    string,
    unknown
  >;
  style?: React.CSSProperties;
}

export const AssetCard = ({
  asset,
  isSelected,
  isBulkMode,
  onClick,
  onDoubleClick,
  explorerfn,
  style,
}: AssetCardProps) => {
  const {
    id,
    fileName,
    fileType,
    filePath,
    thumbnailPath,
    imageWidth,
    imageHeight,
    fileExtension,
  } = asset;

  const [isHovered, setIsHovered] = useState(false);
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const { toggleFavorite } = useAssetActions(asset.id);
  const getThumbnailUrl = (path: string | null) => {
    if (!path) return "thumbnails/placeholdery/generic_placeholder.webp";
    if (path.startsWith("http")) return path;

    const cleanPath = path.startsWith("/") ? path.slice(1) : path;

    return `${API_BASE_URL}/${cleanPath}`;
  };

  const showControls = isHovered || isSelected || isMenuOpen;
  const showCheckbox = isSelected && isBulkMode;

  // Helper icons
  const getFileIcon = () => {
    switch (fileType?.toLowerCase()) {
      case "model":
        return <Box size={14} className="text-white/80" />;
      case "image":
      case "texture":
        return <ImageIcon size={14} className="text-white/80" />;
      default:
        return <FileBox size={14} className="text-white/80" />;
    }
  };

  const formatBytes = (bytes: number, decimals = 0) => {
    if (!+bytes) return "0 B";
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
  };

  // Zatrzymuje propagację, żeby kliknięcie w przycisk nie zaznaczało karty
  const stopProp = (e: React.SyntheticEvent) => {
    e.stopPropagation();
  };

  return (
    <Card
      shadow="sm"
      radius="lg"
      // Usuwamy onClick stąd, bo Karta bywa kapryśna. Przenosimy go do "Click Zone".
      className={`group relative h-full w-full border-none bg-black/20 transition-transform hover:scale-[1.02] ${
        isSelected ? "ring-2 ring-primary" : ""
      }`}
      style={style}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      // Ważne: isPressable={false} domyślnie, traktujemy to jako kontener
    >
      {/* WARSTWA 1 (GÓRA - Z-30): AKCJE
          Te elementy muszą być klikalne i leżeć NAJWYŻEJ.
      */}
      <CardHeader className="absolute top-0 z-30 flex w-full justify-between p-2 pointer-events-none">
        {/* Checkbox Wrapper - pointer-events-auto przywraca klikalność wewnątrz nagłówka */}
        <div
          className={`flex gap-2 transition-opacity duration-200 pointer-events-auto ${
            showCheckbox ? "opacity-100" : "opacity-0"
          }`}
          // Kliknięcie w checkbox-wrapper ma działać jak klik w kartę (zaznaczenie)
          // Ale uwaga: Checkbox w środku ma pointer-events-none, więc ten div łapie klik.
          onClick={onClick}
        >
          <Checkbox
            isSelected={isSelected}
            classNames={{
              wrapper:
                "pointer-events-none bg-black/40 border-white/50 group-data-[selected=true]:bg-primary",
              base: "pointer-events-none",
            }}
          />
        </div>

        {/* Przyciski - pointer-events-auto */}
        <div
          className={`flex gap-1 transition-opacity duration-200 pointer-events-auto ${
            showControls ? "opacity-100" : "opacity-0"
          }`}
          onClick={stopProp}
          onKeyDown={stopProp}
          onDoubleClick={stopProp} // Zapobiega otwarciu explorera przy szybkim klikaniu w menu
        >
          <Button
            isIconOnly
            size="sm"
            radius="full"
            variant="light"
            className="bg-black/40 text-white backdrop-blur-md hover:bg-primary/80"
            onPress={() => explorerfn(filePath)}
          >
            <FolderOpen size={16} />
          </Button>

          <Button
            isIconOnly
            size="sm"
            radius="full"
            variant="light"
            className="bg-black/40 text-white backdrop-blur-md hover:bg-black/60"
            onPress={() => toggleFavorite()}
          >
            <Heart
              size={16}
              className={
                asset.isFavorite ? "fill-danger text-danger" : "text-white"
              }
            />
          </Button>

          <Dropdown placement="bottom-end" onOpenChange={setIsMenuOpen}>
            <DropdownTrigger>
              <Button
                isIconOnly
                size="sm"
                radius="full"
                variant="light"
                className="bg-black/40 text-white backdrop-blur-md hover:bg-black/60"
              >
                <MoreVertical size={16} />
              </Button>
            </DropdownTrigger>
            <DropdownMenu aria-label="Asset Actions">
              <DropdownItem key="rename" startContent={<Edit size={16} />}>
                Rename
              </DropdownItem>
              <DropdownItem
                key="add-set"
                startContent={<FolderPlus size={16} />}
              >
                Add to Collection
              </DropdownItem>
              <DropdownItem
                key="delete"
                className="text-danger"
                color="danger"
                startContent={<Trash size={16} />}
              >
                Delete
              </DropdownItem>
            </DropdownMenu>
          </Dropdown>
        </div>
      </CardHeader>

      {/* WARSTWA 2 (ŚRODEK - Z-20): CLICK ZONE (The Magic Fix 🪄)
          To jest niewidzialna tafla szkła, która łapie wszystkie kliknięcia w "ciało" karty.
      */}
      <div
        className="absolute inset-0 z-20 w-full h-full cursor-pointer"
        onClick={onClick}
        onDoubleClick={onDoubleClick}
      />

      {/* WARSTWA 3 (DÓŁ - Z-0/10): CONTENT
          Elementy wizualne, nie interaktywne.
      */}

      {/* Obrazek - Z-0 */}
      <Image
        removeWrapper
        alt={fileName}
        className="z-0 h-full w-full object-cover pointer-events-none"
        src={getThumbnailUrl(thumbnailPath ?? "")}
        fallbackSrc="/thumbnails/placeholdery/generic_placeholder.webp"
      />

      {/* Footer*/}
      <CardFooter className="absolute bottom-0 z-40 w-full justify-between border-t border-white/10 bg-black/60 p-2 backdrop-blur-md pointer-events-none">
        <div className="flex w-full flex-col items-start gap-1">
          <p className="w-full truncate text-tiny font-bold text-white/90 text-left">
            {fileName}
          </p>
          <div className="flex w-full items-center justify-between mt-1">
            <span className="flex items-center gap-1 text-[10px] text-white/60">
              {getFileIcon()}
              <span className="uppercase tracking-wider font-medium">
                {fileExtension?.replace(".", "")}
              </span>
            </span>
            <div className="flex items-center gap-2 text-[9px] text-white/50">
              {(imageWidth ?? 0) > 0 && (imageHeight ?? 0) > 0 && (
                <span className="flex items-center gap-1">
                  <Maximize2 size={10} className="text-white/40" />
                  {imageWidth}×{imageHeight}
                </span>
              )}
              <span className="flex items-center gap-1">
                <HardDrive size={10} className="text-white/40" />
                {formatBytes(asset.fileSize)}
              </span>
            </div>
          </div>
        </div>
      </CardFooter>
    </Card>
  );
};
