import { Accordion, AccordionItem } from "@heroui/accordion";
import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Chip } from "@heroui/chip";
import { Divider } from "@heroui/divider";
import { Input } from "@heroui/input";
import {
	Modal,
	ModalBody,
	ModalContent,
	ModalFooter,
	ModalHeader,
	useDisclosure,
} from "@heroui/modal";
import { CircularProgress } from "@heroui/progress";
import { Snippet } from "@heroui/snippet";
import { Spinner } from "@heroui/spinner";
import { Switch } from "@heroui/switch";
import { cn } from "@heroui/theme";
import { addToast } from "@heroui/toast";
import { GetAppVersion } from "@wailsjs/go/app/App";
import { OnFileDrop } from "@wailsjs/runtime/runtime";
import {
	AlertCircle,
	CheckCircle2,
	Folder,
	FolderOpen,
	FolderPlus,
	FolderSearch,
	Play,
	StopCircle,
	Trash2,
} from "lucide-react";
import type React from "react";
import { useCallback, useEffect, useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { useToastListener } from "@/hooks/useToastListener";
import { useScanFolders } from "./hooks/useScanFolders";
import { useScanProgress } from "./hooks/useScanProgress";
import { useUpdateChecker } from "./hooks/useUpdateChecker";

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
		appConfig,
		setDebugMode,
	} = useScanFolders();

	const {
		releaseInfo,
		isLoading: isCheckingUpdate,
		isUpdating,
		checkUpdates,
		handleUpdate,
	} = useUpdateChecker();
	const { isOpen, onOpen, onOpenChange } = useDisclosure();

	const [pathInput, setPathInput] = useState("");
	const [extInput, setExtInput] = useState("");
	const [version, setVersion] = useState<string>("");
	const [validationState, setValidationState] = useState<
		"valid" | "invalid" | "idle"
	>("idle");

	useEffect(() => {
		GetAppVersion().then(setVersion);
	}, []);

	useEffect(() => {
		if (releaseInfo?.isUpdateAvailable) {
			onOpen();
		}
	}, [releaseInfo, onOpen]);

	const [backendError, setBackendError] = useState<string>("");
	const { isScanning, progress, message } = useScanProgress();

	// --- GLOBAL TOAST LISTENER ---
	useToastListener();

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
			} catch (_error) {
				setValidationState("invalid");
			}
		},
		[pathInput, validatePath],
	);
	useEffect(() => {
		const handleFileDrop = (_x: number, _y: number, paths: string[]) => {
			if (paths.length > 0) {
				const droppedPath = paths[0];
				setPathInput(droppedPath);
				handleValidate(droppedPath);
			}
		};
		OnFileDrop(handleFileDrop, true);
		return () => {
			OnFileDrop((_x, _y, _paths) => {}, true);
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
			if (!newExt.startsWith(".")) newExt = `.${newExt}`;

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
			{/* SEKCJA 1: HEADER */}
			<div className="flex flex-col gap-1">
				<h1 className="text-3xl font-bold tracking-tight">Library Settings</h1>
				<p className="text-default-500">
					Manage your asset folders and scanner status.
				</p>
			</div>

			{/* SEKCJA 2: ADD NEW FOLDER */}
			<Card className="w-full overflow-visible" shadow="sm">
				<CardHeader className="flex flex-col items-start px-6 pt-6 pb-0">
					<h4 className="text-large font-bold">Add Source Folder</h4>
					<p className="text-small text-default-500">
						Select a folder containing your digital assets.
					</p>
				</CardHeader>
				<CardBody className="px-6 py-6">
					<div className="flex flex-col gap-2">
						<div className="flex flex-row gap-2 items-end">
							<Input
								value={pathInput}
								onChange={(e) => setPathInput(e.target.value)}
								placeholder="Paste path or browse... (e.g. D:\Assets\SciFi)"
								startContent={
									<FolderPlus className="text-default-400" size={20} />
								}
								isInvalid={validationState === "invalid"}
								color={getInputColor()}
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
						<div className="flex flex-col gap-1 px-1">
							{validationState === "invalid" && (
								<p className="text-tiny text-danger">
									{backendError || "Please enter a valid folder path."}
								</p>
							)}
							<p className="text-tiny text-default-400">
								We will automatically check if this folder exists.
							</p>
						</div>
					</div>
				</CardBody>
			</Card>

			{/* SEKCJA 2.5: QUICK SCAN CONTROL */}
			<Card className="w-full border-none bg-content2/50" shadow="none">
				<CardBody className="p-4 sm:p-6">
					<div className="flex flex-col sm:flex-row items-center justify-between gap-4">
						{/* LEWA STRONA: OPIS */}
						<div className="flex flex-col gap-1 text-center sm:text-left">
							<h4 className="text-medium font-bold">Refresh Library</h4>
							<p className="text-small text-default-500 max-w-md">
								Scan your folders to find new assets or update changes you've
								made.
							</p>
						</div>

						{/* PRAWA STRONA: STATUS + BUTTON */}
						<div className="flex flex-col sm:flex-row items-center gap-4 sm:gap-6 w-full sm:w-auto">
							{/* STATUS INDICATOR */}
							<div className="flex flex-col items-center sm:items-end min-w-[120px]">
								<span className="text-[10px] font-bold uppercase text-default-400 tracking-wider mb-1">
									Scanner Status
								</span>
								<div className="flex items-center gap-2 h-6">
									{isScanning && (
										<CircularProgress
											size="sm"
											value={progress}
											color="success"
											strokeWidth={3}
											classNames={{
												svg: "w-4 h-4",
											}}
											aria-label="Scanning progress"
										/>
									)}
									<span
										className={`text-small font-semibold transition-colors ${
											isScanning ? "text-success" : "text-default-500"
										}`}
									>
										{message}
									</span>
								</div>
							</div>

							{/* ACTION BUTTON */}
							<Button
								color={isScanning ? "danger" : "primary"}
								variant="shadow"
								size="lg"
								isLoading={isStartingScan}
								isDisabled={isScanning}
								startContent={
									!isStartingScan &&
									(isScanning ? <StopCircle size={20} /> : <Play size={20} />)
								}
								onPress={() => startScan()}
								className="font-bold min-w-[140px] w-full sm:w-auto"
							>
								{isScanning ? "Stop Scan" : "Scan Now"}
							</Button>
						</div>
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
												onCopy={() => {
													addToast({
														title: "Copied to Clipboard",
														description: "Folder path copied successfully.",
														color: "success",
													});
												}}
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
						<div className="flex flex-col gap-1">
							<div className="flex gap-4 items-end">
								<Input
									placeholder=".blend, .obj, .png"
									value={extInput}
									onValueChange={setExtInput}
									onKeyDown={handleAddExtension}
									className="max-w-xs"
								/>
							</div>
							<p className="text-tiny text-default-400 px-1">
								Press Enter to add.
							</p>
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
									No file types specified. Please add at least one (e.g.,
									.blend, .png) to start scanning.
								</p>
							)}
						</div>
					</CardBody>
				</Card>
			</div>

			{/* SEKCJA 5: ADVANCED SETTINGS */}
			<div className="space-y-4">
				<h2 className="text-2xl font-bold text-danger">Advanced</h2>
				<Accordion
					variant="splitted"
					className="px-0"
					itemClasses={{
						base: "bg-content1 border border-default-100",
						title: "text-medium font-semibold",
						trigger: "px-4",
					}}
				>
					<AccordionItem
						key="advanced"
						aria-label="Advanced System Settings"
						title="Developer & Debug Settings"
						subtitle="Only for troubleshooting"
					>
						<div className="p-2 space-y-6">
							<div className="flex items-center justify-between">
								<div className="flex flex-col gap-1">
									<span className="text-medium font-medium">Debug Mode</span>
									<p className="text-tiny text-default-400">
										Enable detailed logging for troubleshooting scanner and
										matcher issues.
									</p>
								</div>
								<Switch
									isSelected={appConfig?.debugMode}
									onValueChange={setDebugMode}
									color="danger"
								/>
							</div>
						</div>
					</AccordionItem>
				</Accordion>
			</div>

			{/* SEKCJA 6: ABOUT */}
			<div className="pt-8 border-t border-default-100 flex flex-col gap-4">
				<div className="flex justify-between items-end">
					<div className="space-y-1">
						<h2 className="text-xl font-bold">Application</h2>
						<p className="text-tiny text-default-400">
							© 2025 Eclat Asset Manager
						</p>
					</div>
					<div className="flex items-center gap-4">
						<div className="flex flex-col items-end">
							<span className="text-tiny text-default-400 uppercase font-bold">
								Current Version
							</span>
							<span className="font-mono font-bold text-primary">
								{version}
							</span>
						</div>
						<Button
							size="sm"
							variant="flat"
							color="primary"
							isLoading={isCheckingUpdate}
							onPress={() =>
								checkUpdates().then((info) => {
									if (info && !info.isUpdateAvailable) {
										addToast({
											title: "Up to date",
											description: "You are using the latest version of Eclat.",
											color: "success",
										});
									}
								})
							}
						>
							Check for updates
						</Button>
					</div>
				</div>
			</div>

			<Modal
				isOpen={isOpen}
				onOpenChange={onOpenChange}
				scrollBehavior="inside"
				size="2xl"
				backdrop="blur"
				classNames={{
					base: "border border-default-100 bg-content1",
					header: "border-b border-default-100",
					footer: "border-t border-default-100",
				}}
			>
				<ModalContent>
					{(onClose) => (
						<>
							<ModalHeader className="flex flex-col gap-1">
								<div className="flex items-center gap-2">
									<span className="text-primary text-xl font-bold">
										New Version Available
									</span>
									<Chip color="primary" variant="flat" size="sm">
										{releaseInfo?.tagName}
									</Chip>
								</div>
							</ModalHeader>
							<ModalBody className="py-6">
								<div className="space-y-6">
									{releaseInfo?.history && releaseInfo.history.length > 0 ? (
										releaseInfo.history.map((rel, index) => (
											<div key={rel.tagName} className="space-y-2">
												<div className="flex items-center gap-2">
													<div className="h-px flex-1 bg-default-100" />
													<Chip
														size="sm"
														variant="dot"
														color={index === 0 ? "primary" : "default"}
														className="border-none bg-transparent font-bold"
													>
														{rel.tagName} {index === 0 && "(Latest)"}
													</Chip>
													<div className="h-px flex-1 bg-default-100" />
												</div>
												<div className="p-4 rounded-xl bg-default-50 border border-default-100">
													<div className="text-small text-default-700 leading-relaxed font-sans">
														<ReactMarkdown
															remarkPlugins={[remarkGfm]}
															components={{
																h1: ({ children }) => (
																	<h1 className="text-lg font-bold mb-2">
																		{children}
																	</h1>
																),
																h2: ({ children }) => (
																	<h2 className="text-md font-bold mb-2">
																		{children}
																	</h2>
																),
																h3: ({ children }) => (
																	<h3 className="text-sm font-bold mb-1">
																		{children}
																	</h3>
																),
																ul: ({ children }) => (
																	<ul className="list-disc ml-4 mb-2">
																		{children}
																	</ul>
																),
																ol: ({ children }) => (
																	<ol className="list-decimal ml-4 mb-2">
																		{children}
																	</ol>
																),
																li: ({ children }) => (
																	<li className="mb-1">{children}</li>
																),
																code: ({ children }) => (
																	<code className="bg-default-200 px-1 rounded font-mono text-xs">
																		{children}
																	</code>
																),
																p: ({ children }) => (
																	<p className="mb-2 last:mb-0">{children}</p>
																),
															}}
														>
															{rel.body || "No release notes provided."}
														</ReactMarkdown>
													</div>
												</div>
											</div>
										))
									) : (
										<div className="p-4 rounded-xl bg-default-50 border border-default-100">
											<h5 className="text-sm font-bold text-default-500 uppercase tracking-wider mb-2">
												Release Notes
											</h5>
											<div className="text-small text-default-700 leading-relaxed font-sans">
												<ReactMarkdown
													remarkPlugins={[remarkGfm]}
													components={{
														h1: ({ children }) => (
															<h1 className="text-lg font-bold mb-2">
																{children}
															</h1>
														),
														h2: ({ children }) => (
															<h2 className="text-md font-bold mb-2">
																{children}
															</h2>
														),
														h3: ({ children }) => (
															<h3 className="text-sm font-bold mb-1">
																{children}
															</h3>
														),
														ul: ({ children }) => (
															<ul className="list-disc ml-4 mb-2">
																{children}
															</ul>
														),
														ol: ({ children }) => (
															<ol className="list-decimal ml-4 mb-2">
																{children}
															</ol>
														),
														li: ({ children }) => (
															<li className="mb-1">{children}</li>
														),
														code: ({ children }) => (
															<code className="bg-default-200 px-1 rounded font-mono text-xs">
																{children}
															</code>
														),
														p: ({ children }) => (
															<p className="mb-2 last:mb-0">{children}</p>
														),
													}}
												>
													{releaseInfo?.body || "No release notes provided."}
												</ReactMarkdown>
											</div>
										</div>
									)}
									<p className="text-tiny text-default-400 italic text-center pt-2">
										Note: Updates on Windows will restart the application
										automatically. On other platforms, this will open the
										download page in your browser.
									</p>
								</div>
							</ModalBody>
							<ModalFooter>
								<Button variant="light" onPress={onClose}>
									Later
								</Button>
								<Button
									color="primary"
									isLoading={isUpdating}
									onPress={() => {
										if (releaseInfo?.downloadUrl) {
											handleUpdate(releaseInfo.downloadUrl);
										}
									}}
								>
									{navigator.platform.includes("Win")
										? "Update & Restart"
										: "Download Update"}
								</Button>
							</ModalFooter>
						</>
					)}
				</ModalContent>
			</Modal>
		</div>
	);
}
