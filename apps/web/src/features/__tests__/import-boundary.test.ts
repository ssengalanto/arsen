import { describe, it, expect } from "vitest";
import { ESLint } from "eslint";
import { restrictedImportsOptions } from "../../../eslint.config.mjs";

// Lint an inline snippet against the exact `no-restricted-imports` options the
// project config uses, so the boundary rule is verified at its real source.
async function boundaryMessages(code: string) {
  const eslint = new ESLint({
    overrideConfigFile: true,
    overrideConfig: {
      rules: { "no-restricted-imports": ["error", restrictedImportsOptions] },
    },
  });
  const results = await eslint.lintText(code, {
    filePath: "src/features/resource/hooks/use-thing.js",
  });
  const messages = results[0]?.messages ?? [];
  return messages.filter((m) => m.ruleId === "no-restricted-imports");
}

describe("feature import boundary", () => {
  it("flags a cross-feature reach into another slice's internals", async () => {
    const messages = await boundaryMessages(
      `import { useAuthStore } from "@/features/auth/store";\n`,
    );
    expect(messages.length).toBeGreaterThan(0);
    expect(messages[0]?.message).toMatch(/public entry/i);
  });

  it("allows importing another feature's public entry", async () => {
    const messages = await boundaryMessages(
      `import { useSession } from "@/features/auth";\n`,
    );
    expect(messages).toHaveLength(0);
  });

  it("blocks the global swr `mutate`", async () => {
    const messages = await boundaryMessages(`import { mutate } from "swr";\n`);
    expect(messages.length).toBeGreaterThan(0);
  });
});
