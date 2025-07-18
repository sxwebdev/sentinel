import * as React from "react";
import {XIcon} from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./select";
import {cn} from "@shared/lib/utils";

interface SelectWithClearProps {
  value?: string;
  onValueChange: (value: string) => void;
  onClear?: () => void;
  placeholder?: string;
  children: React.ReactNode;
  className?: string;
}

export function SelectWithClear({
  value,
  onValueChange,
  onClear,
  placeholder,
  children,
  className,
}: SelectWithClearProps) {
  const handleClear = (e: React.MouseEvent) => {
    e.stopPropagation();
    e.preventDefault();
    // Сначала вызываем onClear для обновления внешнего состояния
    onClear?.();
    // Затем обновляем внутреннее состояние Select
    onValueChange("");
  };

  return (
    <div className="relative">
      <Select value={value} onValueChange={onValueChange}>
        <SelectTrigger className={cn("pr-8", className)}>
          <SelectValue placeholder={placeholder} />
        </SelectTrigger>
        <SelectContent>{children}</SelectContent>
      </Select>
      {value && value !== "" && onClear && (
        <button
          type="button"
          onClick={handleClear}
          className="absolute right-2 top-1/2 -translate-y-1/2 p-1 hover:bg-gray-100 rounded-sm transition-colors z-10"
        >
          <XIcon className="size-3 text-gray-500" />
        </button>
      )}
    </div>
  );
}
