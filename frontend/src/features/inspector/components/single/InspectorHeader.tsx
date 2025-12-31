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
  Maximize2,
} from "lucide-react";
import { app } from "@wailsjs/go/models";
import { addToast } from "@heroui/toast";
import { useAssetActions } from "../../hooks/useAssetActions";
import { ClipboardSetText } from "../../../../../wailsjs/runtime/runtime";
import { GetThumbnailData } from "../../../../../wailsjs/go/app/AssetService";

interface InspectorHeaderProps {
  asset: app.AssetDetails;
}

export const InspectorHeader = ({ asset }: InspectorHeaderProps) => {
  const [localFileName, setLocalFileName] = useState(asset.fileName);
  const [renameError, setRenameError] = useState("");
  const [displayThumb, setDisplayThumb] = useState<string>("");
  const { renameAsset, isRenaming, updateAssetType, isUpdatingType, openInProgram } = useAssetActions(asset.id);

  useEffect(() => {
    setLocalFileName(asset.fileName);
    setRenameError("");
  }, [asset.fileName]);

  useEffect(() => {
    const loadThumb = async () => {
      if (!asset.thumbnailPath) {
        setDisplayThumb("/placeholders/generic_placeholder.webp");
        return;
      }

      if (asset.thumbnailPath.startsWith("/placeholders/")) {
        setDisplayThumb(asset.thumbnailPath);
        return;
      }

      try {
        const data = await GetThumbnailData(asset.id);
        setDisplayThumb(data);
      } catch (err) {
        console.error("Failed to load thumbnail via Go", err);
        setDisplayThumb("/placeholders/generic_placeholder.webp");
      }
    };

    loadThumb();
  }, [asset.thumbnailPath, asset.id]);

  // --- HANDLERS ---
  const handleSave = () => {
    const trimmed = localFileName.trim();
    
    // Validation
    if (!trimmed) {
      setRenameError("Filename cannot be empty");
      return;
    }
    if (/[\\/:*?"<>|]/.test(trimmed)) {
      setRenameError('Invalid characters: \ / : * ? " < > |');
      return;
    }
    if (trimmed === asset.fileName) {
      setLocalFileName(asset.fileName);
      setRenameError("");
      return;
    }

    renameAsset(trimmed, {
      onSuccess: () => {
        setRenameError("");
      },
      onError: (err: any) => {
        setRenameError(err || "Failed to rename");
        setLocalFileName(asset.fileName);
      }
    });
  };

  const handleCancel = () => {
    setLocalFileName(asset.fileName);
    setRenameError("");
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") e.currentTarget.blur();
    if (e.key === "Escape") {
      handleCancel();
      e.currentTarget.blur();
    }
  };

  const handleCopyPath = async (e?: any) => {
    // Prevent bubbling if triggered from the button to avoid double toast
    if (e && typeof e.stopPropagation === "function") {
      e.stopPropagation();
    }

    await ClipboardSetText(asset.filePath);
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
    updateAssetType(newType);
  };

  return (
    <div className="flex-none p-4 flex flex-col gap-3 border-b border-default-100 bg-content1">
      {/* 0. THUMBNAIL PREVIEW */}
      <div className="flex justify-center mb-1">
        <div 
          className={`relative group overflow-hidden rounded-lg bg-default-100 border border-default-200 w-full aspect-video max-h-40 flex items-center justify-center ${isImage ? 'cursor-pointer' : ''}`}
          onClick={() => isImage && openInProgram(asset.filePath)}
        >
          <img
            src={displayThumb}
            alt={asset.fileName}
            className={`w-full h-full object-contain transition-transform duration-500 ${isImage ? 'group-hover:scale-110' : ''}`}
          />
          {isImage && (
            <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center pointer-events-none">
              <Maximize2 size={24} className="text-white drop-shadow-md" />
            </div>
          )}
        </div>
      </div>

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
          onValueChange={(val) => {
            setLocalFileName(val);
            if (renameError) setRenameError("");
          }}
          onBlur={handleSave}
          onKeyDown={handleKeyDown}
          size="lg"
          placeholder="Enter asset name"
          isInvalid={!!renameError}
          errorMessage={renameError}
          isDisabled={isRenaming}
          classNames={{
            input:
              "font-bold text-medium text-default-900 group-hover/title:text-primary transition-colors",
            inputWrapper:
              "border-b-default-200 group-hover/title:border-b-primary/50 px-0 h-4 transition-colors",
          }}
          endContent={
            localFileName !== asset.fileName && !isRenaming && (
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
          onClick={() => handleCopyPath()}
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
              onClick={handleCopyPath}
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
                isLoading={isUpdatingType}
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