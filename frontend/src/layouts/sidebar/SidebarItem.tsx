import { useLocation, Link } from "react-router-dom";
import { CSSProperties } from "react";

interface SidebarItemProps {
  icon: any;
  label: string;
  to?: string;
  count?: number;
  onClick?: (e: React.MouseEvent) => void;
  className?: string;
  style?: CSSProperties;
}

export const SidebarItem = ({
  icon: Icon,
  label,
  to,
  count,
  onClick,
  className: additionalClassName,
  style,
}: SidebarItemProps) => {
  const location = useLocation();
  const isActive = to
    ? to === "/"
      ? location.pathname === "/"
      : location.pathname.startsWith(to)
    : false;

  const content = (
    <>
      <Icon
        size={18}
        strokeWidth={isActive ? 2.5 : 2}
        className={isActive ? "text-primary" : "text-default-500"}
      />
      <span
        className={`flex-1 truncate text-sm ${isActive ? "text-foreground font-medium" : "text-default-600"}`}
      >
        {label}
      </span>
      {count !== undefined && (
        <span className="text-[10px] font-bold text-default-400 bg-default-100 px-2 py-0.5 rounded-full">
          {count}
        </span>
      )}
      {/* Akcje i ich wrappery ZNIKAJĄ */}
    </>
  );

  const baseClass = `w-full flex items-center gap-3 px-2 h-9 rounded-md transition-colors cursor-pointer select-none ${
    isActive ? "bg-primary/10" : "hover:bg-default-100"
  }`;

  // Zapewniamy, że dodatkowe klasy i style zostaną zaaplikowane na nadrzędny element
  const finalClassName = `${baseClass} ${additionalClassName || ""}`;

  if (to) {
    return (
      <Link to={to} className={finalClassName} style={style}>
        {content}
      </Link>
    );
  }

  // ... div render ...
  return (
    <div onClick={onClick} className={finalClassName} style={style}>
      {content}
    </div>
  );
};
