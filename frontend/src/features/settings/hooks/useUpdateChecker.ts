import { addToast } from "@heroui/toast";
import type { update } from "@wailsjs/go/models";
import {
	CheckForUpdates,
	DownloadAndInstall,
} from "@wailsjs/go/update/UpdateService";
import { useCallback, useEffect, useState } from "react";

export const useUpdateChecker = () => {
	const [releaseInfo, setReleaseInfo] = useState<update.ReleaseInfo | null>(
		null,
	);
	const [isLoading, setIsLoading] = useState(false);
	const [isUpdating, setIsUpdating] = useState(false);

	const checkUpdates = useCallback(async () => {
		setIsLoading(true);
		try {
			const info = await CheckForUpdates();
			setReleaseInfo(info);
			return info;
		} catch (error) {
			console.error("Update check failed:", error);
			addToast({
				title: "Update Check Failed",
				description: "Could not connect to GitHub to check for updates.",
				color: "danger",
			});
		} finally {
			setIsLoading(false);
		}
	}, []);

	const handleUpdate = useCallback(async (url: string) => {
		setIsUpdating(true);
		try {
			const message = await DownloadAndInstall(url);
			addToast({
				title: "Update Started",
				description: message,
				color: "success",
			});
		} catch (error) {
			console.error("Update installation failed:", error);
			addToast({
				title: "Update Failed",
				description: "Failed to download or start the installer.",
				color: "danger",
			});
			setIsUpdating(false);
		}
	}, []);

	useEffect(() => {
		checkUpdates();
	}, [checkUpdates]);

	return {
		releaseInfo,
		isLoading,
		isUpdating,
		checkUpdates,
		handleUpdate,
	};
};
