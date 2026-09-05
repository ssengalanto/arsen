import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";
import tseslint from "typescript-eslint";
import prettier from "eslint-config-prettier";

// Shared so the import-boundary test lints against the exact same options.
export const restrictedImportsOptions = {
  paths: [
    {
      name: "swr",
      importNames: ["mutate"],
      message:
        "Use the bound `mutate` from `useSWRConfig()` or the hook return, not the global `mutate`.",
    },
  ],
  patterns: [
    {
      group: ["@/features/*/*", "@/features/*/**"],
      message:
        "Import a feature's public entry (`@/features/<name>`) only. Use relative paths within a feature; never reach into another feature's internals.",
    },
  ],
};

const NEXT_RESERVED_DEFAULT_EXPORT = [
  "src/app/**/page.tsx",
  "src/app/**/layout.tsx",
  "src/app/**/loading.tsx",
  "src/app/**/error.tsx",
  "src/app/**/global-error.tsx",
  "src/app/**/not-found.tsx",
  "src/app/**/template.tsx",
  "src/app/**/default.tsx",
  "src/app/**/opengraph-image.tsx",
  "**/*.config.{ts,mts,js,mjs}",
  // Storybook requires default exports: the story `meta` and the .storybook config.
  "**/*.stories.tsx",
  ".storybook/**/*.ts",
];

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  ...tseslint.configs.recommendedTypeChecked,
  {
    languageOptions: {
      parserOptions: {
        projectService: {
          // eslint.config.mjs is pulled into the TS project graph via the
          // import-boundary test, so it must NOT also be listed here.
          allowDefaultProject: ["postcss.config.mjs"],
        },
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      "@typescript-eslint/no-explicit-any": "error",
      "no-restricted-imports": ["error", restrictedImportsOptions],
      "no-restricted-syntax": [
        "error",
        {
          selector: "ExportDefaultDeclaration",
          message:
            "Use named exports. Default exports are reserved for Next.js special files (page/layout/error/etc.) and config files.",
        },
      ],
    },
  },
  {
    files: NEXT_RESERVED_DEFAULT_EXPORT,
    rules: {
      "no-restricted-syntax": "off",
    },
  },
  {
    files: ["**/*.{test,spec}.{ts,tsx}", "src/test/**", "e2e/**"],
    rules: {
      "@typescript-eslint/no-explicit-any": "off",
      "@typescript-eslint/no-unsafe-assignment": "off",
      "@typescript-eslint/no-unsafe-member-access": "off",
      "@typescript-eslint/no-unsafe-call": "off",
      "@typescript-eslint/no-unsafe-return": "off",
      "@typescript-eslint/no-unsafe-argument": "off",
    },
  },
  prettier,
  globalIgnores([
    ".next/**",
    "out/**",
    "build/**",
    "coverage/**",
    "playwright-report/**",
    "test-results/**",
    "storybook-static/**",
    "next-env.d.ts",
  ]),
]);

export default eslintConfig;
