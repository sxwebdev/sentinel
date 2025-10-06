import { MoonStarIcon, SunIcon } from "lucide-react";
import { Button } from "@/shared/components/ui/button";
import { useUserPreferenceStore } from "@/app/stores/userPreferience";
import { useShallow } from "zustand/react/shallow";

export function ThemeToggle() {
  const { toggleTheme } = useUserPreferenceStore(
    useShallow((s) => ({
      toggleTheme: s.toggleTheme,
    })),
  );

  return (
    <Button
      variant="ghost"
      size="icon"
      aria-label="Toggle theme"
      onClick={() => toggleTheme()}
    >
      <SunIcon className="size-5 -rotate-90 transition-all dark:rotate-0 dark:opacity-0" />
      <MoonStarIcon className="absolute size-5 -rotate-90 opacity-0 transition-all dark:rotate-0 dark:opacity-100" />
    </Button>
  );
}
