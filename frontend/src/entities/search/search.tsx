import {Button} from "@/shared/components/ui/button";
import {Input} from "@/shared/components/ui/input";
import {Search as SearchIcon, XIcon} from "lucide-react";

interface SearchProps {
  placeholder?: string;
  value?: string;
  onChange?: (value: string) => void;
  clear?: boolean;
}

export const Search = ({
  placeholder = "Search",
  value = "",
  onChange,
  clear = false,
}: SearchProps) => {
  return (
    <div className="flex items-center gap-2 relative">
      <SearchIcon className="absolute left-2 size-4 text-muted-foreground" />
      <Input
        type="text"
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange?.(e.target.value)}
        className="pl-8"
      />
      {clear && value && (
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onChange?.("")}
          className="absolute right-2 p-0 "
        >
          <XIcon className="size-4 text-muted-foreground" />
        </Button>
      )}
    </div>
  );
};
