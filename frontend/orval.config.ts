// orval.config.ts
export default {
  api: {
    input: "./src/shared/types/doc.json", // путь к Swagger JSON
    output: {
      mode: "tags-split", // или 'split' / 'single'
      target: "./src/shared/api/generated.ts", // куда будет сгенерировано
      schemas: "./src/shared/types/model", // типы
      client: "axios", // или 'fetch' / 'axios' / 'swr'
      override: {
        mutator: {
          path: "./src/shared/api/baseApi.ts",
          name: "customFetcher",
        },
      },
    },
  },
};
