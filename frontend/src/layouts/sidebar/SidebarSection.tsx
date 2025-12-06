import { Accordion, AccordionItem } from "@heroui/accordion";
import { ReactNode } from "react";

interface SidebarSectionProps {
	title: string;
	children: ReactNode;
	defaultOpen?: boolean;
}

export const SidebarSection = ({
	title,
	children,
	defaultOpen = true,
}: SidebarSectionProps) => {
	return (
		<Accordion
			isCompact
			showDivider={false}
			defaultExpandedKeys={defaultOpen ? ["1"] : []}
			itemClasses={{
				base: "px-0 w-full",
				trigger: "px-2 py-1 data-[hover=true]:bg-default-100 rounded-md",
				title: "text-xs font-bold uppercase tracking-wider text-default-400",
				indicator: "text-default-300",
				content: "pt-1 pb-2 pl-2",
			}}
		>
			<AccordionItem key="1" aria-label={title} title={title}>
				<div className="flex flex-col gap-0.5">{children}</div>
			</AccordionItem>
		</Accordion>
	);
};
