import path from "node:path";
import { spawn } from "node:child_process";

import { inspectOxcFiles, writeOxcFiles } from "./oxfmt-lib.mjs";
import { renderReport } from "./report.mjs";

function parseArgs(argv) {
  const [, , mode, ...rest] = argv;

  if (mode !== "check" && mode !== "format") {
    throw new Error(
      "usage: node ./tooling/scripts/orchestrate.mjs <check|format> [--output text|json|agent] [--config path] [paths...]",
    );
  }

  let outputFormat = "text";
  let configPath = "";
  const paths = [];

  for (let index = 0; index < rest.length; index += 1) {
    const token = rest[index];

    if (token === "--output") {
      outputFormat = rest[index + 1] ?? "text";
      index += 1;
      continue;
    }

    if (token === "--config") {
      configPath = rest[index + 1] ?? "";
      index += 1;
      continue;
    }

    paths.push(token);
  }

  return { mode, outputFormat, configPath, paths };
}

function runCommand(command, args, options = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd: options.cwd,
      env: options.env ?? process.env,
      stdio: ["ignore", "pipe", "pipe"],
    });

    let stdout = "";
    let stderr = "";

    child.stdout.on("data", (chunk) => {
      stdout += chunk;
    });

    child.stderr.on("data", (chunk) => {
      stderr += chunk;
    });

    child.on("error", reject);
    child.on("close", (code) => {
      resolve({ code: code ?? 1, stdout, stderr });
    });
  });
}

async function runSemantic(mode, root, configPath, paths) {
  const semanticRoot = path.join(root, "semantic");
  const args = ["-C", semanticRoot, "run", "./cmd/fmt", mode, "--format", "json", "--cwd", root];

  if (configPath) {
    args.push("--config", path.resolve(root, configPath));
  }

  for (const targetPath of paths) {
    args.push(path.resolve(root, targetPath));
  }

  const result = await runCommand("go", args, { cwd: root });

  if (result.stderr.trim() !== "") {
    throw new Error(result.stderr.trim());
  }

  if (result.stdout.trim() === "") {
    throw new Error("semantic formatter produced no output");
  }

  return JSON.parse(result.stdout);
}

function toInternalResult(result) {
  return {
    file: result.file,
    applied: result.applied ?? [],
    violations: (result.violations ?? []).map((violation) => ({
      rule: violation.rule,
      line: violation.line ?? 0,
      message: violation.message,
    })),
    changed: Boolean(result.changed),
  };
}

function buildMergedReport(mode, semanticReport, oxcInspection) {
  const report = {
    result: "pass",
    files: semanticReport.files + oxcInspection.files.length,
    changed: semanticReport.changed,
    results: (semanticReport.results ?? []).map(toInternalResult),
    errors: [...(semanticReport.errors ?? [])],
  };

  for (const file of oxcInspection.dirtyFiles) {
    const item = {
      file,
      applied: ["oxfmt"],
      violations: [],
      changed: true,
    };

    if (mode === "check") {
      item.violations.push({
        rule: "oxfmt",
        line: 0,
        message: "file is not formatted",
      });
    }

    report.results.push(item);
    report.changed += 1;
  }

  for (const error of oxcInspection.errors) {
    report.errors.push(error);
  }

  if (report.errors.length > 0) {
    report.result = "fail";
  } else if (mode === "format" && report.changed > 0) {
    report.result = "fixed";
  } else if (mode === "check" && report.changed > 0) {
    report.result = "fail";
  }

  report.results.sort((left, right) => left.file.localeCompare(right.file));
  report.errors.sort((left, right) => left.file.localeCompare(right.file));

  return report;
}

function exitCodeFor(mode, report) {
  if (mode === "check") {
    return report.result === "pass" ? 0 : 1;
  }

  return report.errors.length > 0 ? 1 : 0;
}

try {
  const { mode, outputFormat, configPath, paths } = parseArgs(process.argv);
  const root = process.cwd();
  const semanticReport = await runSemantic(mode, root, configPath, paths);
  const oxcInspection = await inspectOxcFiles(root);

  if (mode === "format") {
    await writeOxcFiles(root, oxcInspection.dirtyFiles);
  }

  const mergedReport = buildMergedReport(mode, semanticReport, oxcInspection);
  process.stdout.write(renderReport(mergedReport, outputFormat, mode));
  process.exit(exitCodeFor(mode, mergedReport));
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`${message}\n`);
  process.exit(1);
}
