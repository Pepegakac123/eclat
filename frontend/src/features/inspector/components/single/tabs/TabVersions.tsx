import { Layers } from "lucide-react";

export const TabVersions = () => {
  return (
    <div className="flex flex-col items-center justify-center py-8 gap-3 text-default-400">
      <div className="w-10 h-10 rounded-full bg-default-100 flex items-center justify-center">
        <Layers size={20} className="opacity-40" />
      </div>
      <div className="text-center">
        <p className="text-small font-medium text-default-600">
          No versions linked
        </p>
        <p className="text-[10px] opacity-60 max-w-[150px]">
          Versioning allows you to group related files (e.g. source file +
          render).
        </p>
      </div>
    </div>
  );
};
