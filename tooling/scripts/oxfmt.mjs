import path from "node:path";
import { fileURLToPath } from "node:url";

import { runOxcOnly } from "./oxfmt-lib.mjs";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");

const mode = process.argv[2];

if (mode !== "check" && mode !== "format") {
  console.error("usage: node ./scripts/oxfmt.mjs <check|format>");
  process.exit(1);
}

const inspection = await runOxcOnly(mode, repoRoot);

if (mode === "check") {
  if (inspection.errors.length > 0) {
    process.exit(1);
  }

  process.exit(inspection.dirtyFiles.length > 0 ? 1 : 0);
}

process.exit(inspection.errors.length > 0 ? 1 : 0);
