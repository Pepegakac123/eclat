import { Button } from "@heroui/button";
import {
	Modal,
	ModalBody,
	ModalContent,
	ModalFooter,
	ModalHeader,
	useDisclosure,
} from "@heroui/modal";
import { ScrollShadow } from "@heroui/scroll-shadow";
import { Skeleton } from "@heroui/skeleton";
import { app } from "@wailsjs/go/models"; // Typy Wails
import { useState } from "react";
import { useMaterialSets } from "@/layouts/sidebar/hooks/useMaterialSets";
import {
	type MaterialSetForm,
	MaterialSetFormModal,
} from "./MaterialSetFormModal"; // Import ujednoliconego Modala
import { MaterialSetSidebarItem } from "./MaterialSetSidebarItem";
import { SidebarSection } from "./SidebarSection";

export const SidebarCollections = () => {
	const {
		materialSets,
		isLoading,
		createMaterialSet,
		isCreating,
		updateMaterialSet,
		isUpdating,
		deleteMaterialSet,
		isDeleting,
		setCoverFromFile,
		isSettingCover,
	} = useMaterialSets();

	// --- STANY MODALI ---
	const {
		isOpen: isCreateOpen,
		onOpen: onCreateOpen,
		onOpenChange: onCreateOpenChange,
	} = useDisclosure();
	// Edycja
	const {
		isOpen: isEditOpen,
		onOpen: onEditOpen,
		onOpenChange: onEditOpenChange,
	} = useDisclosure();
	// Usuwanie
	const {
		isOpen: isDeleteOpen,
		onOpen: onDeleteOpen,
		onOpenChange: onDeleteOpenChange,
	} = useDisclosure();

	const [selectedSet, setSelectedSet] = useState<app.MaterialSet | undefined>(
		undefined,
	);

	// --- HANDLERY MUTACJI ---

	// Create Handler
	const handleCreateSubmit = async (
		data: MaterialSetForm,
		onClose: () => void,
	) => {
		try {
			const payload = new app.CreateMaterialSetRequest({
				name: data.name,
				description: data.description || undefined,
				customCoverUrl: data.customCoverUrl || undefined,
				customColor: data.customColor || undefined,
			});

			const newSet = await createMaterialSet(payload);

			if (data.coverFilePath && newSet?.id) {
				await setCoverFromFile({ id: newSet.id, filePath: data.coverFilePath });
			}

			onClose();
		} catch (error) {
			console.error("Failed to create collection:", error);
		}
	};

	// Edit Handler
	const handleEditSubmit = async (
		data: MaterialSetForm,
		onClose: () => void,
	) => {
		if (!selectedSet) return;

		try {
			const payload = new app.CreateMaterialSetRequest({
				name: data.name,
				description: data.description || undefined,
				customCoverUrl: data.customCoverUrl || undefined,
				customColor: data.customColor || undefined,
			});

			const updatePayload = {
				id: selectedSet.id,
				data: payload,
			};

			await updateMaterialSet(updatePayload);

			if (data.coverFilePath) {
				await setCoverFromFile({
					id: selectedSet.id,
					filePath: data.coverFilePath,
				});
			}

			onClose();
		} catch (error) {
			console.error("Failed to update collection:", error);
		}
	};

	// Delete Handler
	const handleDeleteConfirm = async (onClose: () => void) => {
		if (!selectedSet) return;
		try {
			await deleteMaterialSet(selectedSet.id);
			setSelectedSet(undefined);
			onClose();
		} catch (error) {
			console.error("Failed to delete collection:", error);
		}
	};

	// --- HANDLERY UI ---

	const handleEditOpen = (set: app.MaterialSet) => {
		setSelectedSet(set);
		onEditOpen();
	};

	const handleDeleteOpen = (set: app.MaterialSet) => {
		setSelectedSet(set);
		onDeleteOpen();
	};

	const handleCreateOpen = () => {
		onCreateOpen();
	};

	const isBusy = isCreating || isUpdating || isSettingCover;

	return (
		<SidebarSection title="Collections">
			{/* LISTA KOLEKCJI */}
			<ScrollShadow className="max-h-48 custom-scrollbar" hideScrollBar={false}>
				{isLoading && (
					<div className="flex flex-col gap-2">
						<Skeleton className="h-8 w-full rounded-md" />
						<Skeleton className="h-8 w-full rounded-md" />
						<Skeleton className="h-8 w-full rounded-md" />
					</div>
				)}

				{!isLoading &&
					Array.isArray(materialSets) &&
					materialSets.map((set) => (
						<MaterialSetSidebarItem
							key={set.id}
							set={set}
							handleEditOpen={handleEditOpen}
							handleDeleteOpen={handleDeleteOpen}
						/>
					))}

				{!isLoading && materialSets.length === 0 && (
					<p className="text-xs text-default-400 px-2 py-2">
						No collections yet.
					</p>
				)}
			</ScrollShadow>

			{/* TRIGGER BUTTON */}
			<Button
				size="sm"
				variant="light"
				className="w-full justify-start h-8 text-xs text-default-400 data-[hover=true]:text-primary mt-1 pl-2"
				startContent={<span className="text-lg font-light mr-1">+</span>}
				onPress={handleCreateOpen}
			>
				New Collection
			</Button>

			{/* --- MODALE --- */}

			{/* 1. Modal TWORZENIA (używa ujednoliconego komponentu) */}
			<MaterialSetFormModal
				mode="create"
				isOpen={isCreateOpen}
				onOpenChange={onCreateOpenChange}
				onSubmit={handleCreateSubmit}
				isLoading={isBusy}
			/>

			{/* 2. Modal EDYCJI (używa ujednoliconego komponentu) */}
			<MaterialSetFormModal
				mode="edit"
				initialData={selectedSet}
				isOpen={isEditOpen}
				onOpenChange={onEditOpenChange}
				onSubmit={handleEditSubmit}
				isLoading={isBusy}
				// Wymuszamy re-render, żeby useForm złapał nowe defaultValues przy zmianie selectedSet
				key={selectedSet?.id || "edit-modal-closed"}
			/>

			{/* 3. Modal USUWANIA (potwierdzenie) */}
			<Modal
				isOpen={isDeleteOpen}
				onOpenChange={onDeleteOpenChange}
				placement="center"
				backdrop="blur"
			>
				<ModalContent>
					{(onClose) => (
						<>
							<ModalHeader>Confirm Deletion</ModalHeader>
							<ModalBody>
								Are you sure you want to delete the collection?
								<span className="font-bold text-danger ml-1">
									{selectedSet?.name || "..."}
								</span>
								This operation cannot be undone.
							</ModalBody>
							<ModalFooter>
								<Button color="default" variant="light" onPress={onClose}>
									Cancel
								</Button>
								<Button
									color="danger"
									onPress={() => handleDeleteConfirm(onClose)}
									isLoading={isDeleting}
								>
									Delete
								</Button>
							</ModalFooter>
						</>
					)}
				</ModalContent>
			</Modal>
		</SidebarSection>
	);
};
