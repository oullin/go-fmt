import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";

import { inspectOxcFiles } from "../scripts/oxfmt-lib.mjs";

test("ignores unsupported files and go files", async () => {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "go-fmt-oxc-"));
  await fs.writeFile(path.join(root, "README.md"), "# title\n");
  await fs.writeFile(path.join(root, "main.go"), "package main\n");
  await fs.writeFile(path.join(root, "script.sh"), "echo hi\n");

  const files = await inspectOxcFiles(root);

  assert.equal(files.files.includes("main.go"), false);
  assert.equal(files.files.includes("script.sh"), false);
  assert.equal(files.files.includes("README.md"), true);
});
