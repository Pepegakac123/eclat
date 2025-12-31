import { app } from "@wailsjs/go/models";
import { Copy, FileText, HardDrive, Calendar, Maximize2, Layers, Hash } from "lucide-react";
import { Tooltip } from "@heroui/tooltip";
import { Button } from "@heroui/button";
import { addToast } from "@heroui/toast";
import { ClipboardSetText } from "../../../../../../wailsjs/runtime/runtime";

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

const copyToClipboard = async (text: string, label: string) => {
  await ClipboardSetText(text);
  addToast({
    title: "Copied!",
    description: `${label} copied to clipboard`,
    color: "success",
    variant: "flat",
    timeout: 1500,
  });
};

const DetailTile = ({
  label,
  value,
  icon: Icon,
  onClick,
}: {
  label: string;
  value: string | React.ReactNode;
  icon?: any;
  onClick?: () => void;
}) => (
  <div 
    className={`flex flex-col gap-1 p-2 rounded-lg bg-default-50 border border-default-100 transition-colors ${onClick ? 'cursor-pointer hover:bg-default-100 group' : ''}`}
    onClick={onClick}
  >
    <div className="flex items-center justify-between">
      <span className="text-[10px] uppercase font-bold text-default-400 tracking-wider flex items-center gap-1">
        {Icon && <Icon size={10} />} {label}
      </span>
      {onClick && <Copy size={10} className="opacity-0 group-hover:opacity-100 transition-opacity text-default-400" />}
    </div>
    <span className="text-small font-medium text-default-900 truncate">
      {value}
    </span>
  </div>
);

export const TabDetails = ({ asset }: { asset: app.AssetDetails }) => {
  const isImageOrTexture = asset.fileType?.toLowerCase() === "image" || asset.fileType?.toLowerCase() === "texture";

  return (
    <div className="flex flex-col gap-4 p-1">
      {/* 1. TECHNICAL SPECS */}
      <div className="flex flex-col gap-2">
        <div className="grid grid-cols-2 gap-2">
          <DetailTile
            label="Format"
            icon={FileText}
            value={asset.fileExtension?.toUpperCase().replace(".", "") || asset.fileType}
          />
          <DetailTile 
            label="Size" 
            icon={HardDrive}
            value={formatFileSize(asset.fileSize)} 
          />
          <DetailTile
            label="Modified"
            icon={Calendar}
            value={new Date(asset.lastModified).toLocaleDateString()}
          />
          {isImageOrTexture && (asset.bitDepth ?? 0) > 0 && (
            <DetailTile 
              label="Bit Depth" 
              icon={Layers}
              value={`${asset.bitDepth}-bit`} 
            />
          )}
        </div>
      </div>

      {/* 2. IMAGE SPECS (Conditional) */}
      {isImageOrTexture && (
        <div className="flex flex-col gap-2">
          <div className="grid grid-cols-1 gap-2">
            {asset.imageWidth > 0 && asset.imageHeight > 0 && (
              <DetailTile
                label="Dimensions"
                icon={Maximize2}
                value={`${asset.imageWidth} × ${asset.imageHeight} px`}
              />
            )}

            {/* DOMINANT COLOR */}
            {asset.dominantColor && (
              <DetailTile
                label="Primary Color"
                value={
                  <div className="flex items-center gap-2">
                    <div
                      className="w-3 h-3 rounded-full ring-1 ring-default-300/20 shadow-sm"
                      style={{ backgroundColor: asset.dominantColor }}
                    />
                    <span className="font-mono uppercase">{asset.dominantColor}</span>
                  </div>
                }
                onClick={() => copyToClipboard(asset.dominantColor!, "Color HEX")}
              />
            )}
          </div>
        </div>
      )}

      {/* 3. FILE HASH */}
      {asset.fileHash && (
        <div className="flex flex-col gap-1 mt-1">
          <div className="flex items-center justify-between px-1">
            <span className="text-[10px] uppercase font-bold text-default-400 tracking-wider flex items-center gap-1">
              <Hash size={10} /> File Hash (SHA256)
            </span>
            <Tooltip content="Copy Hash">
              <Button
                size="sm"
                isIconOnly
                variant="light"
                className="h-4 w-4 min-w-0 text-default-400 hover:text-default-700"
                onClick={() => copyToClipboard(asset.fileHash!, "Hash")}
              >
                <Copy size={10} />
              </Button>
            </Tooltip>
          </div>
          <code 
            className="block w-full bg-default-100 border border-default-200 p-2 rounded-md text-[10px] leading-tight text-default-600 break-all font-mono select-all hover:bg-default-200 transition-colors cursor-pointer"
            onClick={() => copyToClipboard(asset.fileHash!, "Hash")}
          >
            {asset.fileHash}
          </code>
        </div>
      )}
    </div>
  );
};

