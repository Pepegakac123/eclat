import { Accordion, AccordionItem } from "@heroui/accordion";
import { Button } from "@heroui/button";
import { Input, Textarea } from "@heroui/input";
import {
	Modal,
	ModalBody,
	ModalContent,
	ModalFooter,
	ModalHeader,
} from "@heroui/modal";
import { zodResolver } from "@hookform/resolvers/zod";
import type { app } from "@wailsjs/go/models"; // Typ z Wails
import { Image as ImageIcon, Upload } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";
import { PRESET_COLORS } from "@/config/constants";

declare global {
	interface Window {
		runtime: {
			OpenFileDialog: (options?: any) => Promise<string>;
		};
	}
}

const formSchema = z.object({
	name: z.string().min(2, "Name is required").max(50, "Name is too long"),
	description: z.string().optional(),
	customCoverUrl: z
		.string()
		.url("Must be a valid URL")
		.optional()
		.or(z.literal("")),
	customColor: z.string().optional(),
	coverFilePath: z.string().optional(),
});
export type MaterialSetForm = z.infer<typeof formSchema>;

interface MaterialSetFormModalProps {
	mode: "create" | "edit";
	initialData?: app.MaterialSet;
	isOpen: boolean;
	onOpenChange: (isOpen: boolean) => void;
	onSubmit: (data: MaterialSetForm, onClose: () => void) => void;
	isLoading: boolean;
}

export const MaterialSetFormModal = ({
	mode,
	initialData,
	isOpen,
	onOpenChange,
	onSubmit,
	isLoading,
}: MaterialSetFormModalProps) => {
	// Używamy useMemo i useEffect do prawidłowego zarządzania stanem RHF przy edycji
	const defaultValues = useMemo(() => {
		return initialData
			? {
					name: initialData.name,
					description: initialData.description || "",
					customCoverUrl: initialData.customCoverUrl || "",
					customColor: initialData.customColor || PRESET_COLORS[0],
					coverFilePath: "",
				}
			: {
					name: "",
					description: "",
					customCoverUrl: "",
					customColor: PRESET_COLORS[0],
					coverFilePath: "",
				};
	}, [initialData]); // Re-kalkulacja gdy zmienia się set (dla edycji)

	const {
		control,
		handleSubmit,
		reset,
		setValue,
		watch,
		formState: { errors },
	} = useForm<MaterialSetForm>({
		resolver: zodResolver(formSchema),
		defaultValues,
	});

	const coverFilePath = watch("coverFilePath");

	useEffect(() => {
		if (isOpen) {
			reset(defaultValues);
		}
	}, [defaultValues, reset, isOpen]);

	const handleSelectFile = async () => {
		try {
			const result = await window.runtime.OpenFileDialog({
				Title: "Select Cover Image",
				Filters: [
					{ DisplayName: "Images", Pattern: "*.png;*.jpg;*.jpeg;*.webp" },
				],
			});
			if (result) {
				setValue("coverFilePath", result);
				setValue("customCoverUrl", ""); // Clear URL if file is selected
			}
		} catch (err) {
			console.error("Failed to open file dialog", err);
		}
	};

	const title =
		mode === "create"
			? "Create Collection"
			: `Edit: ${initialData?.name || "Collection"}`;
	const submitText = mode === "create" ? "Create" : "Save Changes";

	return (
		<Modal
			isOpen={isOpen}
			onOpenChange={onOpenChange}
			placement="center"
			backdrop="blur"
		>
			<ModalContent>
				{(onClose) => (
					<form onSubmit={handleSubmit((data) => onSubmit(data, onClose))}>
						<ModalHeader>{title}</ModalHeader>

						<ModalBody className="flex flex-col gap-4">
							{/* NAME INPUT */}
							<Controller
								name="name"
								control={control}
								render={({ field }) => (
									<Input
										{...field}
										autoFocus
										label="Name"
										placeholder="e.g. Sci-Fi Weapons"
										variant="bordered"
										isInvalid={!!errors.name}
										errorMessage={errors.name?.message}
									/>
								)}
							/>

							{/* DESCRIPTION INPUT */}
							<Controller
								name="description"
								control={control}
								render={({ field }) => (
									<Textarea
										{...field}
										label="Description"
										placeholder="Optional description..."
										variant="bordered"
										minRows={2}
									/>
								)}
							/>

							{/* COLOR PICKER SECTION */}
							<div className="flex flex-col gap-3">
								<span className="text-xs text-default-500">
									Collection Color
								</span>
								<Controller
									name="customColor"
									control={control}
									render={({ field }) => (
										<div className="flex flex-col gap-3">
											<div className="flex flex-wrap gap-2">
												{PRESET_COLORS.map((color) => (
													<button
														key={color}
														type="button"
														onClick={() => field.onChange(color)}
														className={`w-6 h-6 rounded-full border transition-all ${
															field.value === color
																? "ring-2 ring-primary ring-offset-2 ring-offset-content1 border-transparent scale-110"
																: "border-default-200 hover:scale-105"
														}`}
														style={{ backgroundColor: color }}
														title={color}
														aria-label={`Select color ${color}`}
													/>
												))}
											</div>
											<div className="flex items-center gap-2 mt-1">
												<div className="relative">
													<input
														type="color"
														{...field}
														className="opacity-0 absolute inset-0 w-full h-full cursor-pointer z-10"
													/>
													<Button
														size="sm"
														variant="flat"
														className="min-w-0 px-3 bg-default-100 pointer-events-none"
														startContent={
															<div
																className="w-4 h-4 rounded-full border border-default-300"
																style={{ backgroundColor: field.value }}
															/>
														}
													>
														Custom
													</Button>
												</div>

												<span className="text-xs text-default-400 font-mono uppercase">
													{field.value}
												</span>
											</div>
										</div>
									)}
								/>
							</div>

							{/* ADVANCED OPTIONS*/}
							<Accordion isCompact showDivider={false} className="px-0 -mx-2">
								<AccordionItem
									key="advanced"
									aria-label="Advanced Options"
									title="Advanced Options"
									classNames={{
										trigger:
											"py-2 data-[hover=true]:bg-default-100 rounded-lg px-2",
										title: "text-small font-medium text-default-500",
										indicator: "text-default-400",
										content: "pt-2 pb-1 px-2",
									}}
								>
									<div className="flex flex-col gap-3">
										<span className="text-xs text-default-500">
											Collection Cover
										</span>
										<div className="flex gap-2">
											<Button
												size="sm"
												variant="flat"
												startContent={<Upload size={16} />}
												onPress={handleSelectFile}
											>
												Upload Image
											</Button>
											{coverFilePath && (
												<span className="text-xs text-default-500 self-center truncate max-w-[200px]">
													{coverFilePath}
												</span>
											)}
										</div>

										{!coverFilePath && (
											<Controller
												name="customCoverUrl"
												control={control}
												render={({ field }) => (
													<Input
														{...field}
														label="Or Custom Cover URL"
														placeholder="https://example.com/image.jpg"
														variant="bordered"
														size="sm"
														startContent={
															<ImageIcon
																size={16}
																className="text-default-400"
															/>
														}
														isInvalid={!!errors.customCoverUrl}
														errorMessage={errors.customCoverUrl?.message}
													/>
												)}
											/>
										)}
									</div>
								</AccordionItem>
							</Accordion>
						</ModalBody>

						<ModalFooter>
							<Button color="danger" variant="light" onPress={onClose}>
								Cancel
							</Button>
							<Button color="primary" type="submit" isLoading={isLoading}>
								{submitText}
							</Button>
						</ModalFooter>
					</form>
				)}
			</ModalContent>
		</Modal>
	);
};
