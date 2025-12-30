import { app } from "@wailsjs/go/models";
import { Link, useLocation } from "react-router-dom";
import { Tooltip } from "@heroui/tooltip";
import { Button } from "@heroui/button";
import { Shapes, Pencil, Trash2 } from "lucide-react";

interface MaterialSetSidebarItemProps {
  set: app.MaterialSet;
  handleEditOpen: (set: app.MaterialSet) => void;
  handleDeleteOpen: (set: app.MaterialSet) => void;
}

export const MaterialSetSidebarItem = ({
  set,
  handleEditOpen,
  handleDeleteOpen,
}: MaterialSetSidebarItemProps) => {
  const location = useLocation();

  // 1. Logika Active State
  const to = `/collections/${set.id}`;
  const isActive = location.pathname.startsWith(to);

  // 2. Logika Kolorów
  const iconColorClass = isActive
    ? "text-primary"
    : set.customColor
      ? undefined
      : "text-default-500";
  const iconStyle = set.customColor ? { color: set.customColor } : undefined;

  const wrapperClasses = `relative w-full h-9 flex items-center rounded-md transition-colors group px-2 gap-3 cursor-pointer select-none ${
    isActive ? "bg-primary/10" : "hover:bg-default-100"
  }`;

  return (
    <Tooltip
      content={set.description}
      isDisabled={!set.description}
      delay={600}
      placement="right"
      showArrow={true}
      classNames={{
        content: "max-w-[200px] text-xs text-default-500",
      }}
    >
      <div
        className={wrapperClasses}
        style={
          set.customColor
            ? { borderLeft: `3px solid ${set.customColor}` }
            : undefined
        }
      >
        <Link
          to={to}
          className="absolute inset-0 flex items-center px-2 gap-3 w-full h-full z-0 overflow-hidden"
        >
          <Shapes
            size={18}
            strokeWidth={isActive ? 2.5 : 2}
            className={`flex-shrink-0 ${iconColorClass}`}
            style={iconStyle}
          />
          <span
            className={`flex-1 truncate text-sm transition-colors pr-8 ${
              isActive ? "text-foreground font-medium" : "text-default-600"
            }`}
          >
            {set.name}
          </span>
        </Link>

        {/* WARSTWA 2: PRZYCISKI AKCJI*/}
        <div className="absolute right-1 z-10 hidden group-hover:flex items-center gap-0.5 bg-default-100/80 backdrop-blur-sm rounded-md pl-1">
          <Button
            isIconOnly
            size="sm"
            variant="light"
            className="text-default-400 hover:text-primary min-w-6 w-6 h-6 data-[hover=true]:bg-transparent"
            onPress={() => handleEditOpen(set)}
            title={`Edit ${set.name}`}
          >
            <Pencil size={14} />
          </Button>

          <Button
            isIconOnly
            size="sm"
            variant="light"
            className="text-default-400 hover:text-danger min-w-6 w-6 h-6 data-[hover=true]:bg-transparent"
            onPress={() => handleDeleteOpen(set)}
            title={`Delete ${set.name}`}
          >
            <Trash2 size={14} />
          </Button>
        </div>
      </div>
    </Tooltip>
  );
};
