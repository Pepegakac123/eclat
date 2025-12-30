import { Checkbox, CheckboxGroup } from "@heroui/checkbox";
import { Divider } from "@heroui/divider";
import {
  FileType2,
  Ruler,
  Settings2,
  Star,
  X,
  Palette,
  HardDrive,
  Layers,
} from "lucide-react";
import { useGalleryStore } from "@/features/gallery/stores/useGalleryStore";
import { useShallow } from "zustand/react/shallow";
import { ALLOWED_FILE_TYPES, UI_CONFIG } from "@/config/constants";
import { Slider } from "@heroui/slider";
import { PRESET_COLORS } from "@/config/constants";
import { useEffect, useState } from "react";
import { Input } from "@heroui/input";
import { Accordion, AccordionItem } from "@heroui/accordion";
import { useColors } from "../hooks/useAssets";
import { DateRangePicker } from "@heroui/date-picker";
import { parseDate, getLocalTimeZone, today } from "@internationalized/date";
import { CalendarDays } from "lucide-react";
import { I18nProvider } from "@react-aria/i18n";
import { Switch } from "@heroui/switch";

export const SidebarFilters = () => {
  const { filters, setFilters } = useGalleryStore(
    useShallow((state) => ({
      filters: state.filters,
      setFilters: state.setFilters,
    })),
  );
  const avaliableColors = useColors();
  const displayColors =
    avaliableColors &&
    Array.isArray(avaliableColors) &&
    avaliableColors.length > 0
      ? avaliableColors
      : PRESET_COLORS;
  const toggleColor = (color: string) => {
    const currentColors = filters.colors || [];
    const newColors = currentColors.includes(color)
      ? currentColors.filter((c) => c !== color)
      : [...currentColors, color];
    setFilters({ colors: newColors });
  };
  const getDateValue = () => {
    if (filters.dateRange.from && filters.dateRange.to) {
      try {
        return {
          start: parseDate(filters.dateRange.from),
          end: parseDate(filters.dateRange.to),
        };
      } catch (e) {
        console.error("Invalid date format in store", e);
        return null;
      }
    }
    return null;
  };

  // Helper: Zamień obiekt HeroUI na stringi dla Store
  const handleDateChange = (value: any) => {
    if (value && value.start && value.end) {
      setFilters({
        dateRange: {
          from: value.start.toString(), // "2023-11-29"
          to: value.end.toString(),
        },
      });
    } else {
      setFilters({
        dateRange: { from: null, to: null },
      });
    }
  };
  const hasDateFilter = !!(filters.dateRange.from || filters.dateRange.to);

  const resetDate = (e: React.MouseEvent) => {
    e.stopPropagation();
    setFilters({ dateRange: { from: null, to: null } });
  };
  const LazyNumberInput = ({
    value,
    onChange,
    label,
    placeholder,
    maxValue,
    minValue,
  }: {
    value: number;
    onChange: (val: number) => void;
    label: string;
    placeholder?: string;
    maxValue?: number;
    minValue?: number;
  }) => {
    // Lokalny stan do edycji
    const [localValue, setLocalValue] = useState<string>(value.toString());

    useEffect(() => {
      setLocalValue(value.toString());
    }, [value]);

    const commitValue = () => {
      let parsed = parseInt(localValue, 10);

      if (isNaN(parsed) || parsed < 0) parsed = 0;
      onChange(parsed);
      setLocalValue(parsed.toString());
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
      if (e.key === "Enter") {
        commitValue();
        (e.target as HTMLInputElement).blur(); // Usuń focus po Enter
      }
    };
    return (
      <Input
        type="number"
        label={label}
        labelPlacement="inside"
        size="sm"
        placeholder={placeholder}
        value={localValue}
        onValueChange={setLocalValue}
        onBlur={commitValue}
        onKeyDown={handleKeyDown}
        min={minValue ?? 0}
        max={maxValue ?? Infinity}
        classNames={{
          input: "text-tiny",
          label: "text-[10px] text-default-500",
        }}
      />
    );
  };

  return (
    <div className="flex flex-col gap-4 w-full">
      {/* 1. FILE TYPES */}
      <div className="flex flex-col gap-2">
        <div className="flex items-center gap-2 text-default-500 mb-1 px-1">
          <FileType2 size={14} />
          <span className="text-xs font-semibold uppercase tracking-wider">
            File Types
          </span>
        </div>

        <CheckboxGroup
          value={filters.fileTypes}
          onValueChange={(val) => setFilters({ fileTypes: val })}
          size="sm"
          classNames={{
            wrapper: "grid grid-cols-2 gap-2",
          }}
        >
          {ALLOWED_FILE_TYPES.map((type) => (
            <Checkbox
              key={type}
              value={type}
              classNames={{
                label:
                  "text-tiny text-default-600 capitalize truncate select-none",
                wrapper: "mr-1",
              }}
            >
              {type}
            </Checkbox>
          ))}
        </CheckboxGroup>

        <Divider className="my-1 opacity-50" />
      </div>
      {/* 2. RATING */}
      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-2 text-default-500 px-1">
          <Star size={14} />
          <span className="text-xs font-semibold uppercase tracking-wider">
            Rating
          </span>
        </div>

        <Slider
          size="sm"
          step={1}
          minValue={0}
          maxValue={5}
          value={filters.ratingRange}
          onChange={(value) => {
            if (Array.isArray(value)) {
              setFilters({ ratingRange: [value[0], value[1]] });
            }
          }}
          className="max-w-full"
          aria-label="Rating Range"
          showSteps={true}
          showTooltip={true}
          classNames={{
            thumb: "w-4 h-4 after:w-3 after:h-3",
            step: "data-[in-range=true]:bg-black/50 dark:data-[in-range=true]:bg-white/50",
          }}
        />
      </div>

      <Divider className="my-1 opacity-50" />

      {/* 3. COLORS */}
      <div className="flex flex-col gap-2">
        <div className="flex items-center gap-2 text-default-500 px-1">
          <Palette size={14} />
          <span className="text-xs font-semibold uppercase tracking-wider">
            Colors
          </span>
        </div>

        <div className="flex flex-wrap gap-2 px-1">
          {displayColors.map((color) => {
            const isSelected = filters.colors?.includes(color);

            return (
              <button
                key={color}
                type="button"
                onClick={() => toggleColor(color)}
                className={`
                        w-5 h-5 rounded-full shadow-sm transition-all
                        ${
                          isSelected
                            ? "ring-2 ring-primary ring-offset-2 ring-offset-content1 scale-110"
                            : "hover:scale-110 hover:ring-2 hover:ring-default-300 ring-1 ring-transparent"
                        }
                        ${color === "#FFFFFF" ? "border border-default-200" : ""}
                      `}
                style={{ backgroundColor: color }}
                title={color}
                aria-label={`Filter by color ${color}`}
              />
            );
          })}
          {displayColors.length === 0 && (
            <span className="text-tiny text-default-400 italic">
              No colors found
            </span>
          )}
        </div>
      </div>

      <Divider className="my-1 opacity-50" />
      {/* 4. ADVANCED - ACCORDION */}
      <Accordion
        isCompact
        showDivider={false}
        className="px-0 -mx-2"
        itemClasses={{
          title: "text-small text-default-500 font-semibold",
          trigger:
            "py-2 data-[hover=true]:bg-default-100 rounded-lg transition-colors px-2",
          content: "pb-4 px-2 flex flex-col gap-5",
        }}
      >
        <AccordionItem
          key="advanced"
          aria-label="Advanced Filters"
          title={
            <div className="flex items-center gap-2">
              <Settings2 size={14} />
              <span>Advanced Properties</span>
            </div>
          }
        >
          {/* --- FILE SIZE (MB) --- */}
          <div className="flex flex-col gap-3">
            <div className="flex items-center gap-2 text-default-400 pb-1">
              <HardDrive size={12} />
              <span className="text-[10px] font-bold tracking-wide uppercase">
                File Size (MB)
              </span>
            </div>

            <Slider
              size="sm"
              step={1}
              minValue={0}
              maxValue={2048} // 2 GB
              value={filters.fileSizeRange}
              onChange={(value) => {
                if (Array.isArray(value)) {
                  setFilters({ fileSizeRange: [value[0], value[1]] });
                }
              }}
              className="max-w-full"
              aria-label="File Size Range"
              // Formatowanie tooltipa żeby dodawał "MB"
              getValue={(v) =>
                Array.isArray(v) ? `${v[0]} MB - ${v[1]} MB` : `${v} MB`
              }
            />

            <div className="flex items-center gap-2">
              <LazyNumberInput
                label="Min MB"
                value={filters.fileSizeRange[0]}
                onChange={(v) =>
                  setFilters({ fileSizeRange: [v, filters.fileSizeRange[1]] })
                }
                maxValue={2048}
              />
              <LazyNumberInput
                label="Max MB"
                value={filters.fileSizeRange[1]}
                onChange={(v) =>
                  setFilters({ fileSizeRange: [filters.fileSizeRange[0], v] })
                }
                maxValue={2048}
              />
            </div>
          </div>
          <Divider className="my-1 opacity-50" />
          {/* --- DIMENSIONS --- */}
          <div className="flex flex-col gap-3">
            <div className="flex items-center gap-2 text-default-400 pb-1 ">
              <Ruler size={12} />
              <span className="text-[10px] font-bold tracking-wide uppercase">
                Dimensions (px)
              </span>
            </div>

            {/* Width */}
            <div className="flex items-center gap-2">
              <LazyNumberInput
                label="Min Width"
                value={filters.widthRange[0]}
                onChange={(v) =>
                  setFilters({ widthRange: [v, filters.widthRange[1]] })
                }
              />
              <LazyNumberInput
                label="Max Width"
                value={filters.widthRange[1]}
                onChange={(v) =>
                  setFilters({ widthRange: [filters.widthRange[0], v] })
                }
                maxValue={UI_CONFIG.GALLERY.FilterOptions.MAX_DIMENSION}
              />
            </div>

            {/* Height */}
            <div className="flex items-center gap-2">
              <LazyNumberInput
                label="Min Height"
                value={filters.heightRange[0]}
                onChange={(v) =>
                  setFilters({ heightRange: [v, filters.heightRange[1]] })
                }
              />
              <LazyNumberInput
                label="Max Height"
                value={filters.heightRange[1]}
                onChange={(v) =>
                  setFilters({ heightRange: [filters.heightRange[0], v] })
                }
                maxValue={UI_CONFIG.GALLERY.FilterOptions.MAX_DIMENSION}
              />
            </div>
            <Divider className="my-1 opacity-50" />
            {/* --- DATE ADDED --- */}
            <div className="flex flex-col gap-2">
              <div className="flex justify-between items-center text-default-400">
                <div className="flex items-center gap-2">
                  <CalendarDays size={12} />
                  <span className="text-[10px] font-bold tracking-wide uppercase">
                    Date Added
                  </span>
                </div>
                {hasDateFilter && (
                  <button
                    onClick={resetDate}
                    className="text-default-400 hover:text-danger hover:cursor-pointer transition-colors p-1 rounded-full hover:bg-default-100"
                    title="Clear date filter"
                    type="button"
                  >
                    <X size={12} />
                  </button>
                )}
              </div>

              <I18nProvider locale="pl-PL">
                <DateRangePicker
                  aria-label="Filter by Date"
                  size="sm"
                  variant="bordered"
                  labelPlacement="outside"
                  value={getDateValue()}
                  onChange={handleDateChange}
                  visibleMonths={1}
                  pageBehavior="single"
                  showMonthAndYearPickers
                  classNames={{
                    inputWrapper: "bg-default-50 border-default-200 pr-1",
                    label: "hidden",
                    input: "text-[10px]",
                  }}
                />
              </I18nProvider>
            </div>

            <Divider className="my-3 opacity-50" />

            {/* --- ALPHA CHANNEL--- */}
            <div className="flex items-center justify-between px-1">
              <div className="flex items-center gap-2 text-default-500">
                <Layers size={14} />
                <span className="text-tiny font-medium text-default-600">
                  Alpha Channel
                </span>
              </div>

              <Switch
                size="sm"
                color="primary"
                isSelected={filters.hasAlpha === true}
                onValueChange={(val) =>
                  setFilters({ hasAlpha: val ? true : null })
                }
                classNames={{
                  wrapper: "group-data-[selected=true]:bg-primary",
                }}
              />
            </div>
          </div>
        </AccordionItem>
      </Accordion>
    </div>
  );
};
