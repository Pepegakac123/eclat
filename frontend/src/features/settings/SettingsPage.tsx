import React, { useState, useEffect, useCallback } from "react";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Button } from "@heroui/button";
import { Input } from "@heroui/input";
import { Switch } from "@heroui/switch";
import { cn } from "@heroui/theme";
import { Divider } from "@heroui/divider";
import { Chip } from "@heroui/chip";
import { Snippet } from "@heroui/snippet";
import { CircularProgress } from "@heroui/progress";
import { Spinner } from "@heroui/spinner";
import {
  FolderPlus,
  Trash2,
  Folder,
  Play,
  StopCircle,
  AlertCircle,
  CheckCircle2,
  FolderOpen,
  FolderSearch,
} from "lucide-react";
import { useScanFolders } from "./hooks/useScanFolders";
import { useScanProgress } from "./hooks/useScanProgress";
import { addToast } from "@heroui/toast";
import { EventsOn, OnFileDrop } from "@wailsjs/runtime/runtime";

export default function SettingsPage() {
  const {
    folders,
    isLoading,
    addFolder,
    deleteFolder,
    validatePath,
    openFolderPicker,
    isValidating,
    updateFolderStatus,
    startScan,
    isStartingScan,
    extensions,
    addExtension,
    removeExtension,
    openInExplorer,
  } = useScanFolders();

  const [pathInput, setPathInput] = useState("");
  const [extInput, setExtInput] = useState("");
  const [validationState, setValidationState] = useState<
    "valid" | "invalid" | "idle"
  >("idle");
  const [backendError, setBackendError] = useState<string>("");
  const { isScanning, progress } = useScanProgress();

  // --- GLOBAL TOAST LISTENER ---
  useEffect(() => {
    const stopToast = EventsOn("toast", (data: any) => {
      addToast({
        title: data.title,
        description: data.message,
        color: data.type === "error" ? "danger" : data.type,
        timeout: 4000,
      });
    });
    return () => stopToast();
  }, []);

  const handleValidate = useCallback(
    async (pathToCheck?: string) => {
      const path = pathToCheck || pathInput;
      if (!path.trim()) {
        setValidationState("idle");
        return;
      }

      try {
        const result = await validatePath(path);
        setValidationState(result.isValid ? "valid" : "invalid");
      } catch (error) {
        setValidationState("invalid");
      }
    },
    [pathInput, validatePath],
  );
  useEffect(() => {
    const handleFileDrop = (x: number, y: number, paths: string[]) => {
      if (paths.length > 0) {
        const droppedPath = paths[0];
        setPathInput(droppedPath);
        handleValidate(droppedPath);
      }
    };
    OnFileDrop(handleFileDrop, true);
    return () => {
      OnFileDrop((x, y, paths) => {}, true);
    };
  }, [handleValidate]);
  // Handler kliknięcia
  const handleBrowse = async () => {
    const selectedPath = await openFolderPicker();
    if (selectedPath) {
      setPathInput(selectedPath);

      const validation = await validatePath(selectedPath);
      setValidationState(validation.isValid ? "valid" : "invalid");
    }
  };
  const handleAddFolder = async () => {
    if (validationState === "valid") {
      setBackendError("");
      try {
        await addFolder(pathInput);
        setPathInput("");
        setValidationState("idle");
      } catch (error: any) {
        console.error("Błąd dodawania:", error);
        const serverMessage = error || "Failed to add folder. Check logs.";

        setBackendError(serverMessage);
        setValidationState("invalid");
      }
    }
  };

  const getInputColor = () => {
    if (validationState === "valid") return "success";
    if (validationState === "invalid") return "danger";
    return "default";
  };

  const getEndContent = () => {
    if (isValidating) return <Spinner size="sm" />;
    if (validationState === "valid")
      return <CheckCircle2 className="text-success" />;
    if (validationState === "invalid")
      return <AlertCircle className="text-danger" />;
    return null;
  };

  const handleAddExtension = async (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && extInput.trim()) {
      let newExt = extInput.trim().toLowerCase();
      if (!newExt.startsWith(".")) newExt = "." + newExt;

      if (!extensions.includes(newExt)) {
        await addExtension(newExt);
        setExtInput("");
      } else {
        addToast({
          title: "Duplicate Extension",
          description: `This extension (${newExt}) is already in the list.`,
          color: "warning",
        });
        setExtInput("");
      }
    }
  };

  const handleRemoveExtension = async (extToRemove: string) => {
    await removeExtension(extToRemove);
  };

  return (
    <div className="w-full mx-auto p-6 space-y-8">
      {/* SEKCJA 1: HEADER & SCANNER CONTROL */}
      <div className="flex flex-col md:flex-row justify-between items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            Library Settings
          </h1>
          <p className="text-default-500">
            Manage your asset folders and scanner status.
          </p>
        </div>

        <Card className="w-full md:w-auto border-none bg-content2">
          <CardBody className="flex flex-row items-center gap-4 p-3">
            <div className="flex flex-col">
              <span className="text-xs font-semibold uppercase text-default-500">
                Scanner Status
              </span>
              <div className="flex items-center gap-2">
                <span
                  className={`text-sm font-bold ${
                    isScanning ? "text-success" : "text-default-400"
                  }`}
                >
                  {isScanning ? "RUNNING" : "IDLE"}
                </span>

                {isScanning && (
                  <CircularProgress
                    size="sm"
                    value={progress}
                    color="success"
                    showValueLabel={true}
                    strokeWidth={4}
                    classNames={{
                      svg: "w-8 h-8",
                      value: "text-[10px]",
                    }}
                    aria-label="Scanning progress"
                  />
                )}
              </div>
            </div>

            <Divider orientation="vertical" className="h-8" />

            <Button
              color={isScanning ? "danger" : "primary"}
              variant="shadow"
              isLoading={isStartingScan}
              isDisabled={isScanning}
              startContent={
                !isStartingScan &&
                (isScanning ? <StopCircle size={18} /> : <Play size={18} />)
              }
              onPress={() => startScan()}
            >
              {isScanning ? "Scanning..." : "Scan Now"}
            </Button>
          </CardBody>
        </Card>
      </div>

      {/* SEKCJA 2: ADD NEW FOLDER */}
      <Card className="w-full overflow-visible" shadow="sm">
        <CardHeader className="flex flex-col items-start px-6 pt-6 pb-0">
          <h4 className="text-large font-bold">Add Source Folder</h4>
          <p className="text-small text-default-500">
            Path must be accessible by the server.
          </p>
        </CardHeader>
        <CardBody className="px-6 py-6">
          <div className="flex gap-2 items-stretch">
            <Input
              label="Path"
              value={pathInput}
              onChange={(e) => setPathInput(e.target.value)}
              placeholder="Paste path or browse... (e.g. D:\Assets\SciFi)"
              startContent={
                <FolderPlus className="text-default-400" size={20} />
              }
              errorMessage={
                validationState === "invalid"
                  ? backendError || "Path validation failed locally."
                  : ""
              }
              isInvalid={validationState === "invalid"}
              color={getInputColor()}
              description="Backend will validate if path exists on blur."
              className="flex-1"
              size="lg"
              variant="bordered"
              onBlur={() => handleValidate()}
              onKeyDown={(e) => {
                if (e.key === "Enter") handleValidate();
              }}
              endContent={
                <div className="flex items-center gap-2">
                  {getEndContent()}
                  <div className="h-6 w-px bg-default-300 mx-1" />{" "}
                  <Button
                    isIconOnly
                    size="sm"
                    variant="light"
                    onPress={handleBrowse}
                    title="Browse Folders"
                  >
                    <FolderSearch size={18} className="text-default-500" />
                  </Button>
                </div>
              }
            />
            <Button
              size="lg"
              color="primary"
              isDisabled={validationState !== "valid" || isValidating}
              variant="solid"
              onPress={handleAddFolder}
            >
              Add Library
            </Button>
          </div>
        </CardBody>
      </Card>

      {/* SEKCJA 3: FOLDER LIST */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Spinner size="lg" color="primary" label="Loading folders..." />
        </div>
      ) : (
        <div>
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-xl font-semibold flex items-center gap-2">
              <Folder size={20} /> Linked Folders
              <Chip size="sm" variant="flat">
                {folders?.length}
              </Chip>
            </h3>
          </div>

          <div className="grid grid-cols-1 gap-4">
            {folders?.map((folder) => (
              <Card
                key={folder.id}
                className={cn(
                  "flex items-center justify-between p-4 rounded-lg transition-all",
                  !folder.isActive &&
                    "opacity-60 bg-default-100 grayscale-[0.5]",
                )}
              >
                <CardBody className="flex flex-row items-center justify-between p-4 gap-4">
                  <div className="flex items-center gap-4 overflow-hidden flex-1">
                    <div
                      className={`p-3 rounded-xl ${
                        folder.isActive
                          ? "bg-primary/10 text-primary"
                          : "bg-default-100 text-default-400"
                      }`}
                    >
                      <Folder size={24} />
                    </div>
                    <div className="flex flex-col overflow-hidden">
                      <Snippet
                        symbol=""
                        className="bg-transparent p-0 text-medium font-medium truncate w-full"
                        codeString={folder.path}
                      >
                        {folder.path}
                      </Snippet>
                      <span className="text-tiny text-default-400">
                        {folder.isActive
                          ? "Monitoring active"
                          : "Monitoring paused"}
                      </span>
                    </div>
                  </div>

                  {/* AKCJE */}
                  <div className="flex items-center gap-2">
                    <Button
                      isIconOnly
                      variant="light"
                      onPress={() => openInExplorer(folder.path)}
                      title="Open in Explorer"
                    >
                      <FolderOpen size={20} className="text-default-500" />
                    </Button>
                    <Switch
                      size="sm"
                      color="primary"
                      isSelected={folder.isActive}
                      onValueChange={(isSelected) =>
                        updateFolderStatus({
                          id: folder.id,
                          isActive: isSelected,
                        })
                      }
                      aria-label="Toggle folder activity"
                    />

                    <Divider orientation="vertical" className="h-6" />

                    <Button
                      isIconOnly
                      color="danger"
                      variant="light"
                      onPress={() => deleteFolder(folder.id)}
                    >
                      <Trash2 size={20} />
                    </Button>
                  </div>
                </CardBody>
              </Card>
            ))}
          </div>
        </div>
      )}

      {/* SEKCJA 4: FILE EXTENSIONS */}
      <div className="space-y-4">
        <h2 className="text-2xl font-bold">File Types</h2>
        <Card className="bg-content1">
          <CardBody className="p-6 space-y-4">
            <div className="flex gap-4 items-end">
              <Input
                label="Add Extension"
                placeholder=".blend, .obj, .png"
                value={extInput}
                onValueChange={setExtInput}
                onKeyDown={handleAddExtension}
                description="Press Enter to add."
                className="max-w-xs"
              />
            </div>

            <Divider />

            <div className="flex flex-wrap gap-2">
              {extensions.map((ext) => (
                <Chip
                  key={ext}
                  onClose={() => handleRemoveExtension(ext)}
                  variant="flat"
                  color="primary"
                >
                  {ext}
                </Chip>
              ))}
              {extensions.length === 0 && (
                <p className="text-default-400 text-sm italic">
                  No extensions defined. Scanner will allow everything
                  (dangerous!) or nothing depending on logic.
                </p>
              )}
            </div>
          </CardBody>
        </Card>
      </div>
    </div>
  );
}
