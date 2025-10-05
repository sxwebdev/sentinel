import { MoonStarIcon, SunIcon } from "lucide-react";
import { useTheme } from "@/shared/components/theme-provider";
import { Button } from "@/shared/components/ui/button";

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();

  return (
    <Button
      variant="ghost"
      size="icon"
      aria-label="Toggle theme"
      onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
    >
      <SunIcon className="size-5 -rotate-90 transition-all dark:rotate-0 dark:opacity-0" />
      <MoonStarIcon className="absolute size-5 -rotate-90 opacity-0 transition-all dark:rotate-0 dark:opacity-100" />
    </Button>
  );
}
