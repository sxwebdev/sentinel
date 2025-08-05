import MultipleSelector, {
  type Option,
} from "@shared/components/ui/multiselect";
import { cn } from "../lib/utils";

interface MultiSelectProps {
  options: Option[];
  value: Option[];
  onChange: (value: Option[]) => void;
  placeholder?: string;
  className?: string;
}

export default function MultiSelect({
  options,
  value,
  onChange,
  placeholder = "Select",
  className,
}: MultiSelectProps) {
  return (
    <div className={cn("*:not-first:mt-2", className)}>
      <MultipleSelector
        commandProps={{
          label: placeholder,
        }}
        value={value}
        options={options}
        placeholder={placeholder}
        onChange={onChange}
        hideClearAllButton
        hidePlaceholderWhenSelected
        emptyIndicator={
          <p className="text-center text-sm pt-3">No results found</p>
        }
      />
    </div>
  );
}
