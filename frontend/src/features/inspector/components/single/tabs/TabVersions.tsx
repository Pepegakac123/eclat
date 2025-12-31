import { useQueryClient } from "@tanstack/react-query";
import { ArrowRight, Clock, FileWarning, GitCommitVertical, Layers, Trash2 } from "lucide-react";
import React, { useMemo, useState } from "react";
import { DeleteAssetModal } from "@/features/gallery/components/DeleteAssetModal";
import { useGalleryStore } from "@/features/gallery/stores/useGalleryStore";
import { Button } from "@heroui/button";
import { Chip } from "@heroui/chip";
import { Listbox, ListboxItem } from "@heroui/listbox";
import { ScrollShadow } from "@heroui/scroll-shadow";
import { Spinner } from "@heroui/spinner";
import { addToast } from "@heroui/toast";
import { Tooltip } from "@heroui/tooltip";
import { app } from "@wailsjs/go/models";
import { DeleteAssetsPermanently } from "../../../../../../wailsjs/go/app/AssetService";
import { useAssetVersions } from "../../../hooks/useAssetVersions";

interface TabVersionsProps {
	asset: app.AssetDetails;
}

const formatFileSize = (bytes: number) => {
	if (bytes === 0) return "0 B";
	const k = 1024;
	const sizes = ["B", "KB", "MB", "GB"];
	const i = Math.floor(Math.log(bytes) / Math.log(k));
	return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
};

export const TabVersions = ({ asset }: TabVersionsProps) => {
	const queryClient = useQueryClient();
	const { data: siblings, isLoading } = useAssetVersions(asset.id);
	const [isDeletingDupId, setIsDeletingDupId] = useState<number | null>(null);
	const [deleteModalOpen, setDeleteModalOpen] = useState(false);
	const [selectedDup, setSelectedDup] = useState<{
		id: number;
		name: string;
	} | null>(null);

	const selectAsset = useGalleryStore((state) => state.selectAsset);

	const { duplicates, timeline } = useMemo(() => {
		if (!siblings)
			return {
				duplicates: [] as app.AssetDetails[],
				timeline: [] as app.AssetDetails[],
			};

		// Filter out the current asset itself for the related lists
		const others = siblings.filter((s) => s.id !== asset.id);

		const dups = others.filter((s) => s.fileHash === asset.fileHash);
		const time = others.filter((s) => s.fileHash !== asset.fileHash);

		return { duplicates: dups, timeline: time };
	}, [siblings, asset.id, asset.fileHash]);

	const handleDeleteClick = (id: number, name: string) => {
		setSelectedDup({ id, name });
		setDeleteModalOpen(true);
	};

	const handleDeleteConfirm = async () => {
		if (!selectedDup) return;

		setIsDeletingDupId(selectedDup.id);
		try {
			await DeleteAssetsPermanently([selectedDup.id]);
			addToast({
				title: "Deleted",
				description: "Duplicate removed permanently",
				color: "success",
			});
			queryClient.invalidateQueries({ queryKey: ["asset-versions", asset.id] });
			queryClient.invalidateQueries({ queryKey: ["assets"] });
			queryClient.invalidateQueries({ queryKey: ["sidebar-stats"] });
			queryClient.invalidateQueries({ queryKey: ["library-stats"] });
			setDeleteModalOpen(false);
		} catch (err) {
			console.error(err);
			addToast({
				title: "Error",
				description: "Failed to delete duplicate",
				color: "danger",
			});
		} finally {
			setIsDeletingDupId(null);
			setSelectedDup(null);
		}
	};

	if (isLoading) {
		return (
			<div className="flex justify-center py-8">
				<Spinner size="sm" color="primary" />
			</div>
		);
	}

	return (
		<>
			<ScrollShadow className="flex-1 -m-1 p-1 max-h-[400px] custom-scrollbar">
				<div className="flex flex-col gap-6">
					{/* SECTION: EXACT DUPLICATES */}
					{duplicates.length > 0 && (
						<div className="flex flex-col gap-2 animate-appearance-in">
							<div className="flex items-center gap-2 px-1 text-warning font-semibold text-[10px] uppercase tracking-wider">
								<FileWarning size={14} />
								Exact Duplicates ({duplicates.length})
							</div>
							<div className="rounded-xl border border-warning/20 bg-warning/5 overflow-hidden">
								<Listbox
									aria-label="Duplicate files list"
									className="p-0"
									itemClasses={{
										base: "px-3 py-2 border-b border-warning/10 last:border-0 hover:bg-warning/10 transition-colors",
									}}
								>
									{duplicates.map((dup: app.AssetDetails) => (
										<ListboxItem
											key={dup.id}
											textValue={dup.fileName}
											onClick={() => selectAsset(dup.id, false)}
											endContent={
												<Tooltip content="Delete Duplicate" color="danger">
													<Button
														isIconOnly
														size="sm"
														variant="light"
														color="danger"
														className="h-7 w-7 min-w-0"
														isLoading={isDeletingDupId === dup.id}
														onClick={(e) => {
															e.stopPropagation();
															handleDeleteClick(dup.id, dup.fileName);
														}}
													>
														<Trash2 size={14} />
													</Button>
												</Tooltip>
											}
										>
											<div className="flex flex-col gap-0.5 overflow-hidden">
												<Tooltip
													content={dup.filePath}
													placement="top-start"
													delay={500}
													closeDelay={0}
												>
													<span className="text-tiny font-medium text-warning-700 truncate">
														{dup.fileName}
													</span>
												</Tooltip>
												<div className="flex items-center gap-2">
													<span className="text-[9px] text-warning-600/70 font-mono uppercase">
														{dup.fileType}
													</span>
													<div className="w-1 h-1 rounded-full bg-warning-300" />
													<span className="text-[9px] text-warning-600/70">
														{formatFileSize(dup.fileSize)}
													</span>
												</div>
											</div>
										</ListboxItem>
									))}
								</Listbox>
							</div>
						</div>
					)}

					{/* SECTION: TIMELINE / HISTORY */}
					<div className="flex flex-col gap-2">
						<div className="flex items-center gap-2 px-1 text-default-500 font-semibold text-[10px] uppercase tracking-wider">
							<Clock size={14} />
							Asset Timeline
						</div>
						<div className="relative ml-3.5 pl-6 border-l-2 border-default-100 flex flex-col gap-4">
							{/* CURRENT ASSET MARKER (Always Shown) */}
							<div className="relative">
								<div className="absolute -left-[31px] top-1 w-4 h-4 rounded-full bg-primary flex items-center justify-center ring-4 ring-content1">
									<GitCommitVertical size={10} className="text-white" />
								</div>
								<div className="flex flex-col gap-1 p-2 rounded-lg bg-primary/5 border border-primary/10">
									<div className="flex items-center justify-between">
										<span className="text-tiny font-bold text-primary italic">
											Active Version
										</span>
										<Chip
											size="sm"
											variant="flat"
											color="primary"
											className="h-4 text-[8px] uppercase"
										>
											Current
										</Chip>
									</div>
									<span className="text-tiny text-default-900 truncate font-medium">
										{asset.fileName}
									</span>
									<span className="text-[9px] text-default-400">
										Modified {new Date(asset.lastModified).toLocaleString()}
									</span>
								</div>
							</div>

							{/* OTHER VERSIONS */}
							{timeline.map((version: app.AssetDetails) => (
								<div
									key={version.id}
									className="relative group animate-appearance-in"
									onClick={() => selectAsset(version.id, false)}
								>
									<div className="absolute -left-[31px] top-1 w-4 h-4 rounded-full bg-default-200 flex items-center justify-center ring-4 ring-content1 group-hover:bg-default-300 transition-colors">
										<GitCommitVertical size={10} className="text-default-500" />
									</div>
									<div className="flex flex-col gap-1 p-2 rounded-lg bg-default-50 border border-default-100 group-hover:border-default-200 transition-all cursor-pointer">
										<div className="flex items-center justify-between">
											<span className="text-tiny font-medium text-default-700 truncate max-w-[70%]">
												{version.fileName}
											</span>
											<span className="text-[9px] text-default-400 font-mono">
												{formatFileSize(version.fileSize)}
											</span>
										</div>
										<div className="flex items-center justify-between">
											<span className="text-[9px] text-default-400">
												{new Date(version.lastModified).toLocaleDateString()}
											</span>
											<Tooltip content="Switch to this version">
												<ArrowRight
													size={10}
													className="text-default-300 opacity-0 group-hover:opacity-100 transition-opacity"
												/>
											</Tooltip>
										</div>
									</div>
								</div>
							))}

							{timeline.length === 0 && !siblings && (
								<div className="py-2 text-[10px] text-default-400 italic">
									No other historical versions found.
								</div>
							)}
						</div>
					</div>
				</div>
			</ScrollShadow>

			<DeleteAssetModal
				isOpen={deleteModalOpen}
				onClose={() => setDeleteModalOpen(false)}
				onConfirm={handleDeleteConfirm}
				fileName={selectedDup?.name || ""}
				isLoading={isDeletingDupId !== null}
			/>
		</>
	);
};