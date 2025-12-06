import { Asset } from "@/types/api";
import { Copy } from "lucide-react";
import { Tooltip } from "@heroui/tooltip";
import { Button } from "@heroui/button";
import { addToast } from "@heroui/toast";

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

const copyToClipboard = (text: string, label: string) => {
  navigator.clipboard.writeText(text);
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
}: {
  label: string;
  value: string | React.ReactNode;
}) => (
  <div className="flex flex-col gap-1 p-2 rounded-lg bg-default-50 border border-default-100">
    <span className="text-[10px] uppercase font-bold text-default-400 tracking-wider">
      {label}
    </span>
    <span className="text-small font-medium text-default-900 truncate">
      {value}
    </span>
  </div>
);

export const TabDetails = ({ asset }: { asset: Asset }) => {
  return (
    <div className="flex flex-col gap-3 p-1">
      <div className="grid grid-cols-2 gap-2">
        <DetailTile
          label="Format"
          value={asset.fileExtension?.toUpperCase() || asset.fileType}
        />
        <DetailTile label="File Size" value={formatFileSize(asset.fileSize)} />

        {asset.imageWidth && asset.imageHeight && (
          <DetailTile
            label="Dimensions"
            value={`${asset.imageWidth} × ${asset.imageHeight}`}
          />
        )}

        <DetailTile
          label="Created"
          value={new Date(asset.dateAdded).toLocaleDateString()}
        />

        {asset.bitDepth && (
          <DetailTile label="Bit Depth" value={`${asset.bitDepth}-bit`} />
        )}

        {/* DOMINANT COLOR */}
        {asset.dominantColor && (
          <div
            className="flex flex-col gap-1 p-2 rounded-lg bg-default-50 border border-default-100 cursor-pointer hover:bg-default-100 transition-colors group"
            onClick={() => copyToClipboard(asset.dominantColor!, "Color HEX")}
          >
            <span className="text-[10px] uppercase font-bold text-default-400 tracking-wider flex justify-between">
              Color
              <Copy
                size={10}
                className="opacity-0 group-hover:opacity-100 transition-opacity"
              />
            </span>
            <div className="flex items-center gap-2">
              <div
                className="w-4 h-4 rounded-full ring-1 ring-default-300/20 shadow-sm"
                style={{ backgroundColor: asset.dominantColor }}
              />
              <span className="text-small font-mono font-medium text-default-900 uppercase">
                {asset.dominantColor}
              </span>
            </div>
          </div>
        )}
      </div>

      {/* 2. FILE HASH */}
      {asset.fileHash && (
        <div className="mt-2">
          <div className="flex items-center justify-between mb-1">
            <span className="text-[10px] uppercase font-bold text-default-400 tracking-wider">
              File Hash (SHA256)
            </span>
            <Tooltip content="Copy Hash">
              <Button
                size="sm"
                isIconOnly
                variant="light"
                className="h-4 w-4 min-w-0 text-default-400 hover:text-default-700"
                onPress={() => copyToClipboard(asset.fileHash!, "Hash")}
              >
                <Copy size={10} />
              </Button>
            </Tooltip>
          </div>
          <code className="block w-full bg-default-100 border border-default-200 p-2 rounded-md text-[10px] leading-tight text-default-600 break-all font-mono select-all hover:bg-default-200 transition-colors cursor-text">
            {asset.fileHash}
          </code>
        </div>
      )}
    </div>
  );
};
