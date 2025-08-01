import { defineConfig } from "orval";

export default defineConfig({
  api: {
    input: "../docs/docsv1/swagger.json", // путь к Swagger JSON
    output: {
      mode: "tags-split", // или 'split' / 'single'
      target: "./src/shared/api/generated.ts", // куда будет сгенерировано
      schemas: "./src/shared/types/model", // типы
      client: "axios",
      override: {
        mutator: {
          path: "./src/shared/api/baseApi.ts",
          name: "customFetcher",
        },
      },
    },
    hooks: {
      afterAllFilesWrite: "pnpm format",
    },
  },
});
