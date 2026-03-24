import test from "node:test";
import assert from "node:assert/strict";

import { renderReport } from "../scripts/report.mjs";

test("renders merged json output with oxfmt entries", () => {
  const report = {
    result: "fail",
    files: 2,
    changed: 1,
    results: [
      {
        file: "README.md",
        applied: ["oxfmt"],
        violations: [
          {
            rule: "oxfmt",
            line: 0,
            message: "file is not formatted",
          },
        ],
        changed: true,
      },
    ],
    errors: [],
  };

  const output = JSON.parse(renderReport(report, "json", "check"));
  assert.equal(output.files, 2);
  assert.equal(output.changed, 1);
  assert.equal(output.results[0].file, "README.md");
  assert.equal(output.results[0].applied[0], "oxfmt");
});

test("renders merged agent summary counts", () => {
  const report = {
    result: "fixed",
    files: 1,
    changed: 1,
    results: [
      {
        file: "README.md",
        applied: ["oxfmt"],
        violations: [],
        changed: true,
      },
    ],
    errors: [],
  };

  const output = JSON.parse(renderReport(report, "agent", "format"));
  assert.equal(output.summary.files, 1);
  assert.equal(output.summary.changed, 1);
  assert.equal(output.changed[0].steps[0], "oxfmt");
});

test("renders text output when no supported files exist", () => {
  const report = {
    result: "pass",
    files: 0,
    changed: 0,
    results: [],
    errors: [],
  };

  assert.equal(renderReport(report, "text", "check"), "No supported files found.\n");
});
