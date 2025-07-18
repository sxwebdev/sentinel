import MultipleSelector, {type Option} from "@shared/components/ui/multiselect";

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
    <div className="*:not-first:mt-2">
      <MultipleSelector
        commandProps={{
          label: placeholder,
        }}
        value={value}
        options={options}
        placeholder={placeholder}
        className={className}
        onChange={onChange}
        hideClearAllButton
        hidePlaceholderWhenSelected
        emptyIndicator={<p className="text-center text-sm">No results found</p>}
      />
    </div>
  );
}
