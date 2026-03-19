import { create } from "zustand";
import { persist } from "zustand/middleware";

type Theme = "none" | "dark" | "light";

type UserPreferenceStore = {
  theme: Theme;
  toggleTheme: () => void;
  initTheme: () => void;
};

export const useUserPreferenceStore = create<UserPreferenceStore>()(
  persist(
    (set, get) => ({
      theme: "none",
      toggleTheme: () => {
        const currentTheme = get().theme;
        const newTheme = currentTheme === "dark" ? "light" : "dark";
        set({ theme: newTheme });
        get().initTheme();
      },
      initTheme: () => {
        const systemTheme = window.matchMedia("(prefers-color-scheme: dark)")
          .matches
          ? "dark"
          : "light";

        if (get().theme === "none") {
          set({ theme: systemTheme });
        }

        const root = window.document.documentElement;

        root.classList.remove("light", "dark");
        root.classList.add(get().theme);
      },
    }),
    {
      name: "userPreferenceStore",
      partialize: (state) => ({
        theme: state.theme,
      }),
    },
  ),
);
