import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTypeScript from "eslint-config-next/typescript";

export default defineConfig([
  ...nextVitals,
  ...nextTypeScript,
  { rules: { curly: ["error", "multi-line"] } },
  globalIgnores(["playwright-report/**", "test-results/**", "src/shared/api/*.generated.ts"]),
]);
