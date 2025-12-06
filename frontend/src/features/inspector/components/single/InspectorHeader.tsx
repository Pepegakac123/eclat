import { useState, useEffect, KeyboardEvent } from "react";
import { Input } from "@heroui/input";
import { Button } from "@heroui/button";
import { Tooltip } from "@heroui/tooltip";
import {
  Copy,
  Check,
  X,
  FileText,
  RefreshCw,
  Layers,
  Edit3,
} from "lucide-react";
import { Asset } from "@/types/api";
import { addToast } from "@heroui/toast";
import { useAssetMutation } from "../../hooks/useAsset";

interface InspectorHeaderProps {
  asset: Asset;
}

export const InspectorHeader = ({ asset }: InspectorHeaderProps) => {
  const [localFileName, setLocalFileName] = useState(asset.fileName);
  const { patch } = useAssetMutation(asset.id);

  useEffect(() => {
    setLocalFileName(asset.fileName);
  }, [asset.id, asset.fileName]);

  // --- HANDLERS ---
  const handleSave = () => {
    const trimmed = localFileName.trim();
    if (!trimmed || trimmed === asset.fileName) {
      setLocalFileName(asset.fileName);
      return;
    }
    patch({ fileName: trimmed });
  };

  const handleCancel = () => {
    setLocalFileName(asset.fileName);
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") e.currentTarget.blur();
    if (e.key === "Escape") {
      handleCancel();
      e.currentTarget.blur();
    }
  };

  const handleCopyPath = () => {
    navigator.clipboard.writeText(asset.filePath);
    addToast({
      title: "Path Copied",
      description: "Copied to clipboard",
      color: "success",
      variant: "flat",
      timeout: 2000,
    });
  };

  // --- LOGIKA ZMIANY TYPU ---
  const isImage = asset.fileType?.toLowerCase() === "image";
  const isTexture = asset.fileType?.toLowerCase() === "texture";
  const canConvert = isImage || isTexture;

  const handleConvertType = () => {
    const newType = isImage ? "texture" : "image";
    patch({ fileType: newType });
    addToast({
      title: "Type Updated",
      description: `Converted to ${newType}`,
      color: "primary",
      variant: "solid",
    });
  };

  return (
    <div className="flex-none p-4 flex flex-col gap-3 border-b border-default-100 bg-content1">
      {/* 1. SEKCJA TYTUŁU */}
      <div className="flex flex-col gap-1 group/title">
        <span className="text-[10px] uppercase font-bold text-default-400 tracking-wider flex items-center gap-1 justify-between">
          <div className="flex items-center gap-1">
            <FileText size={10} /> Asset Name
          </div>
          <Edit3
            size={10}
            className="opacity-0 group-hover/title:opacity-100 transition-opacity text-default-400"
          />
        </span>
        <Input
          variant="underlined"
          value={localFileName}
          onValueChange={setLocalFileName}
          onBlur={handleSave}
          onKeyDown={handleKeyDown}
          size="lg"
          placeholder="Enter asset name"
          classNames={{
            input:
              "font-bold text-medium text-default-900 group-hover/title:text-primary transition-colors",
            inputWrapper:
              "border-b-default-200 group-hover/title:border-b-primary/50 px-0 h-4 transition-colors",
          }}
          endContent={
            localFileName !== asset.fileName && (
              <div className="flex gap-1 animate-appearance-in">
                <button
                  type="button"
                  onClick={handleSave}
                  className="text-success hover:text-success-600 transition-colors"
                >
                  <Check size={16} />
                </button>
                <button
                  type="button"
                  onClick={handleCancel}
                  className="text-danger hover:text-danger-600 transition-colors"
                >
                  <X size={16} />
                </button>
              </div>
            )
          }
        />
      </div>

      {/* 2. SEKCJA ŚCIEŻKI */}
      <div className="flex flex-col gap-1">
        <span className="text-[10px] uppercase font-bold text-default-400 tracking-wider">
          File Path
        </span>
        <div
          className="flex items-center gap-2 group p-2 rounded-medium bg-default-100 border border-default-200 cursor-pointer hover:bg-default-200 transition-colors"
          onClick={handleCopyPath}
          title={asset.filePath}
        >
          <div className="flex-1 truncate font-mono text-tiny text-default-600">
            {asset.filePath}
          </div>
          <Tooltip content="Copy Path">
            <Button
              size="sm"
              isIconOnly
              variant="light"
              className="h-6 w-6 min-w-0 text-default-400 group-hover:text-default-700"
            >
              <Copy size={12} />
            </Button>
          </Tooltip>
        </div>
      </div>

      {/* 3. SEKCJA TYPU PLIKU */}
      <div className="flex flex-col gap-1">
        <span className="text-[10px] uppercase font-bold text-default-400 tracking-wider flex items-center gap-1">
          <Layers size={10} /> File Type
        </span>
        <div className="flex items-center justify-between p-2 rounded-medium bg-default-50 border border-default-100">
          <div className="flex items-center gap-2">
            <span className="uppercase text-tiny font-bold text-default-700 bg-default-200 px-2 py-0.5 rounded-full">
              {asset.fileType}
            </span>{" "}
          </div>

          {canConvert && (
            <Tooltip content={`Convert to ${isImage ? "Texture" : "Image"}`}>
              <Button
                size="sm"
                variant="flat"
                color="primary"
                className="h-6 text-[12px] font-medium px-2 flex flex-row items-center"
                startContent={<RefreshCw size={10} />}
                onPress={handleConvertType}
              >
                To {isImage ? "Texture" : "Image"}
              </Button>
            </Tooltip>
          )}
        </div>
      </div>
    </div>
  );
};
