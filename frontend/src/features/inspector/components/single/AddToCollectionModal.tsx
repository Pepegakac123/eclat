import { Button } from "@heroui/button";
import { Input } from "@heroui/input";
import {
	Modal,
	ModalBody,
	ModalContent,
	ModalFooter,
	ModalHeader,
} from "@heroui/modal";
import { ScrollShadow } from "@heroui/scroll-shadow";
import { app } from "@wailsjs/go/models";
import { Check, FolderPlus, Plus, Search, Shapes, X } from "lucide-react";
import { useMemo, useState } from "react";
import { useMaterialSets } from "@/layouts/sidebar/hooks/useMaterialSets";
import {
	type MaterialSetForm,
	MaterialSetFormModal,
} from "@/layouts/sidebar/MaterialSetFormModal";

interface AddToCollectionModalProps {
	isOpen: boolean;
	onOpenChange: (open: boolean) => void;
	asset: app.AssetDetails;
}

export const AddToCollectionModal = ({
	isOpen,
	onOpenChange,
	asset,
}: AddToCollectionModalProps) => {
	const {
		materialSets,
		addAssetToSet,
		removeAssetFromSet,
		createMaterialSet,
		setCoverFromFile,
	} = useMaterialSets();

	const [searchQuery, setSearchQuery] = useState("");
	const [loadingSetId, setLoadingSetId] = useState<number | null>(null);

	const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
	const [isCreatingSet, setIsCreatingSet] = useState(false);
	const [hoveredSetId, setHoveredSetId] = useState<number | null>(null);

	const filteredSets = useMemo(() => {
		return materialSets.filter((set) =>
			set.name.toLowerCase().includes(searchQuery.toLowerCase()),
		);
	}, [materialSets, searchQuery]);

	const assetSetIds = useMemo(() => {
		return (asset.materialSets || []).map((s) => s.id);
	}, [asset.materialSets]);

	const handleToggle = async (setId: number, isAdded: boolean) => {
		setLoadingSetId(setId);
		try {
			if (isAdded) {
				await removeAssetFromSet({ setId, assetId: asset.id });
			} else {
				await addAssetToSet({ setId, assetId: asset.id });
			}
		} finally {
			setLoadingSetId(null);
		}
	};

	const handleCreateSet = async (
		data: MaterialSetForm,
		onCloseForm: () => void,
	) => {
		setIsCreatingSet(true);
		try {
			let coverAssetId: number | undefined;

			// If no custom cover provided (file or url), use current asset
			if (!data.coverFilePath && !data.customCoverUrl) {
				coverAssetId = asset.id;
			}

			const payload = new app.CreateMaterialSetRequest({
				name: data.name,
				description: data.description || undefined,
				customCoverUrl: data.customCoverUrl || undefined,
				customColor: data.customColor || undefined,
				coverAssetId: coverAssetId,
			});

			const newSet = await createMaterialSet(payload);

			if (newSet?.id) {
				// If a cover file was selected, set it now
				if (data.coverFilePath) {
					await setCoverFromFile({
						id: newSet.id,
						filePath: data.coverFilePath,
					});
				}

				// Add the asset to the newly created set
				await addAssetToSet({ setId: newSet.id, assetId: asset.id });
			}
			onCloseForm();
			// setSearchQuery("") - handled in onOpenChange
		} catch (error) {
			console.error("Failed to create set", error);
		} finally {
			setIsCreatingSet(false);
		}
	};

	// Logika zamykania modala tworzenia
	const handleCreateModalOpenChange = (open: boolean) => {
		setIsCreateModalOpen(open);

		// Jeśli zamykamy modal (czy to przez Cancel, X, czy po sukcesie)
		// resetujemy wyszukiwarkę w głównym oknie
		if (!open) {
			setSearchQuery("");
		}
	};

	return (
		<>
			<Modal
				isOpen={isOpen}
				onOpenChange={onOpenChange}
				scrollBehavior="inside"
				backdrop="blur"
				size="md"
			>
				<ModalContent>
					{(onClose) => (
						<>
							<ModalHeader className="flex flex-col gap-1">
								Add to Collection
								<span className="text-tiny font-normal text-default-400">
									Select collections for{" "}
									<span className="font-mono text-foreground">
										{asset.fileName}
									</span>
								</span>
							</ModalHeader>

							<ModalBody className="pt-0">
								<Input
									placeholder="Search collections..."
									startContent={
										<Search size={16} className="text-default-400" />
									}
									value={searchQuery}
									onValueChange={setSearchQuery}
									variant="faded"
									size="sm"
									classNames={{ inputWrapper: "bg-default-100" }}
								/>

								<ScrollShadow className="h-[300px] mt-2">
									<div className="flex flex-col gap-1">
										{filteredSets.length > 0 ? (
											filteredSets.map((set) => {
												const isAlreadyAdded = assetSetIds.includes(set.id);
												const isLoading = loadingSetId === set.id;
												const isHovered = hoveredSetId === set.id;

												return (
													<div
														key={set.id}
														className="flex items-center justify-between p-2 rounded-lg hover:bg-default-100 transition-colors border border-transparent hover:border-default-200"
														onMouseEnter={() => setHoveredSetId(set.id)}
														onMouseLeave={() => setHoveredSetId(null)}
													>
														<div className="flex items-center gap-3 overflow-hidden">
															<div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary">
																<Shapes
																	size={16}
																	style={{
																		color: set.customColor || undefined,
																	}}
																	className={
																		!set.customColor ? "text-primary" : ""
																	}
																/>
															</div>
															<span className="text-small text-default-700 truncate">
																{set.name}
															</span>
														</div>

														<Button
															size="sm"
															isIconOnly
															variant={isAlreadyAdded ? "flat" : "light"}
															color={
																isAlreadyAdded
																	? isHovered
																		? "danger"
																		: "primary"
																	: "default"
															}
															className={`rounded-full transition-colors ${
																!isAlreadyAdded
																	? "text-default-400 hover:text-primary hover:bg-primary/10"
																	: isHovered
																		? "bg-danger/10 text-danger"
																		: "bg-primary/10 text-primary"
															}`}
															onPress={() => handleToggle(set.id, isAlreadyAdded)}
															isLoading={isLoading}
														>
															{!isLoading &&
																(isAlreadyAdded ? (
																	isHovered ? (
																		<X size={16} />
																	) : (
																		<Check size={16} />
																	)
																) : (
																	<Plus size={18} />
																))}
														</Button>
													</div>
												);
											})
										) : (
											<div className="flex flex-col items-center justify-center py-8 text-default-400 gap-2">
												<p>No collections found.</p>
												{/* Przycisk tworzenia z nazwą z wyszukiwarki */}
												<Button
													size="sm"
													variant="flat"
													onPress={() => setIsCreateModalOpen(true)}
												>
													Create "{searchQuery}"
												</Button>
											</div>
										)}
									</div>
								</ScrollShadow>
							</ModalBody>

							<ModalFooter className="flex justify-between items-center">
								<Button
									variant="light"
									color="primary"
									startContent={<FolderPlus size={18} />}
									onPress={() => setIsCreateModalOpen(true)}
								>
									New Collection
								</Button>

								<Button variant="light" onPress={onClose}>
									Done
								</Button>
							</ModalFooter>
						</>
					)}
				</ModalContent>
			</Modal>

			<MaterialSetFormModal
				isOpen={isCreateModalOpen}
				onOpenChange={handleCreateModalOpenChange}
				mode="create"
				isLoading={isCreatingSet}
				onSubmit={handleCreateSet}
				// Przekazujemy searchQuery jako initialData (tylko nazwę).
				initialData={
					searchQuery
						? app.MaterialSet.createFrom({ name: searchQuery })
						: undefined
				}
			/>
		</>
	);
};
