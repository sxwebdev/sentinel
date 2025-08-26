import { defineConfig } from "vite";
import path from "path";
import react from "@vitejs/plugin-react-swc";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import tailwindcss from "@tailwindcss/vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    tanstackRouter({
      target: "react",
      autoCodeSplitting: true,
    }),
    react(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "@app": path.resolve(__dirname, "./src/app"),
      "@processes": path.resolve(__dirname, "./src/processes"),
      "@pages": path.resolve(__dirname, "./src/pages"),
      "@widgets": path.resolve(__dirname, "./src/widgets"),
      "@shared": path.resolve(__dirname, "./src/shared"),
      "@entities": path.resolve(__dirname, "./src/entities"),
      "@features": path.resolve(__dirname, "./src/features"),
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // React and core libraries
          react: ["react", "react-dom", "@tanstack/react-router"],

          // UI libraries
          ui: [
            "radix-ui",
            "class-variance-authority",
            "clsx",
            "tailwind-merge",
            "sonner",
            "cmdk",
          ],

          // Data and forms
          forms: ["formik", "yup"],

          // Charts and visualization
          charts: ["recharts"],

          // State management
          state: ["zustand"],

          // Data tables
          tables: ["@tanstack/react-table"],

          // Icons
          icons: ["lucide-react"],

          // CSS utilities
          css: ["tw-animate-css"],
        },
      },
    },
    chunkSizeWarningLimit: 500, // Increase warning threshold to 1MB
  },
});
