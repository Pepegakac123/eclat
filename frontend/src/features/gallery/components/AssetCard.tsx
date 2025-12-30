import { Button } from "@heroui/button";
import { Card, CardFooter, CardHeader } from "@heroui/card";
import { Checkbox } from "@heroui/checkbox";
import {
	Dropdown,
	DropdownItem,
	DropdownMenu,
	DropdownTrigger,
} from "@heroui/dropdown";
import { Image } from "@heroui/image";
import {
	Modal,
	ModalBody,
	ModalContent,
	ModalFooter,
	ModalHeader,
} from "@heroui/modal";
import { GetThumbnailData } from "@wailsjs/go/app/AssetService";
import type { app } from "@wailsjs/go/models";
import {
	Box,
	Edit,
	FileBox,
	FolderOpen,
	FolderPlus,
	HardDrive,
	Heart,
	Image as ImageIcon,
	Maximize2,
	MoreVertical,
	Trash,
} from "lucide-react";
import { useEffect, useState } from "react";
import { API_BASE_URL } from "@/config/constants";
import { useAssetActions } from "@/features/inspector/hooks/useAssetActions";

interface AssetCardProps {
	asset: app.AssetDetails;
	isSelected: boolean;
	isBulkMode: boolean;
	onClick: (e: React.MouseEvent) => void;
	onDoubleClick: () => void;
	explorerfn: (path: string) => void;
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
	const [isDeleteOpen, setIsDeleteOpen] = useState(false);
	const [displayThumb, setDisplayThumb] = useState<string>("");
	const { toggleFavorite, deleteAsset, isDeleting } = useAssetActions(asset.id);

	useEffect(() => {
		const loadThumb = async () => {
			if (!thumbnailPath) {
				setDisplayThumb("/placeholders/generic_placeholder.webp");
				return;
			}

			if (thumbnailPath.startsWith("/placeholders/")) {
				setDisplayThumb(thumbnailPath);
				return;
			}

			// For generated thumbnails, try to use the Go method to bypass Vite dev server issues
			try {
				const data = await GetThumbnailData(id);
				setDisplayThumb(data);
			} catch (err) {
				console.error("Failed to load thumbnail via Go", err);
				setDisplayThumb("/placeholders/generic_placeholder.webp");
			}
		};

		loadThumb();
	}, [thumbnailPath, id]);

	const showControls = isHovered || isSelected || isMenuOpen;
	const showCheckbox = isSelected && isBulkMode;

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
		return `${parseFloat((bytes / k ** i).toFixed(dm))} ${sizes[i]}`;
	};

	// Zatrzymuje propagację, żeby kliknięcie w przycisk nie zaznaczało karty
	const stopProp = (e: React.SyntheticEvent) => {
		e.stopPropagation();
	};

	const handleMenuAction = (key: React.Key) => {
		if (key === "delete") {
			setIsDeleteOpen(true);
		}
	};

	const handleDeleteConfirm = () => {
		deleteAsset(undefined, {
			onSuccess: () => setIsDeleteOpen(false),
		});
	};

	return (
		<>
			<Card
				shadow="sm"
				radius="lg"
				className={`group relative h-full w-full border-none bg-black/20 transition-transform hover:scale-[1.02] ${
					isSelected ? "ring-2 ring-primary" : ""
				}`}
				style={style}
				onMouseEnter={() => setIsHovered(true)}
				onMouseLeave={() => setIsHovered(false)}
			>
				<CardHeader className="absolute top-0 z-30 flex w-full justify-between p-2 pointer-events-none">
					<div
						className={`flex gap-2 transition-opacity duration-200 pointer-events-auto ${
							showCheckbox ? "opacity-100" : "opacity-0"
						}`}
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

					<div
						className={`flex gap-1 transition-opacity duration-200 pointer-events-auto ${
							showControls ? "opacity-100" : "opacity-0"
						}`}
						onClick={stopProp}
						onKeyDown={stopProp}
						onDoubleClick={stopProp}
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
							<DropdownMenu
								aria-label="Asset Actions"
								onAction={handleMenuAction}
							>
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

				<div
					className="absolute inset-0 z-20 w-full h-full cursor-pointer"
					onClick={onClick}
					onDoubleClick={onDoubleClick}
				/>

				{/* Obrazek - Z-0 */}
				<Image
					removeWrapper
					alt={fileName}
					className="z-0 h-full w-full object-cover pointer-events-none"
					src={displayThumb}
					fallbackSrc="/placeholders/generic_placeholder.webp"
				/>

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

			<Modal isOpen={isDeleteOpen} onClose={() => setIsDeleteOpen(false)}>
				<ModalContent>
					{(onClose) => (
						<>
							<ModalHeader className="flex flex-col gap-1">
								Delete Asset
							</ModalHeader>
							<ModalBody>
								<p>
									Are you sure you want to permanently delete <b>{fileName}</b>?
									This action cannot be undone and the file will be removed from
									your disk.
								</p>
							</ModalBody>
							<ModalFooter>
								<Button color="default" variant="light" onPress={onClose}>
									Cancel
								</Button>
								<Button
									color="danger"
									onPress={handleDeleteConfirm}
									isLoading={isDeleting}
								>
									Delete
								</Button>
							</ModalFooter>
						</>
					)}
				</ModalContent>
			</Modal>
		</>
	);
};
