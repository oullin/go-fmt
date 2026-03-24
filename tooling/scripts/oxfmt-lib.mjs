import fs from "node:fs/promises";
import path from "node:path";
import { spawn } from "node:child_process";
import { fileURLToPath } from "node:url";

const toolingRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const oxfmtBinary = path.join(toolingRoot, "node_modules", ".bin", "oxfmt");

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

function normalizePath(root, file) {
  return path.relative(root, path.resolve(root, file)).split(path.sep).join("/");
}

function unsupportedByOxc(output) {
  const text = output.toLowerCase();

  return text.includes("no files found matching the given patterns") || text.includes("on 0 files");
}

async function listFilesFromGit(root) {
  const result = await runCommand(
    "git",
    ["ls-files", "--cached", "--others", "--exclude-standard"],
    { cwd: root },
  );

  if (result.code !== 0) {
    return null;
  }

  return result.stdout
    .split(/\r?\n/)
    .map((file) => file.trim())
    .filter(Boolean);
}

async function walkFiles(root, current = root, files = []) {
  const entries = await fs.readdir(current, { withFileTypes: true });

  for (const entry of entries) {
    if (entry.name === ".git" || entry.name === "node_modules") {
      continue;
    }

    const fullPath = path.join(current, entry.name);

    if (entry.isDirectory()) {
      await walkFiles(root, fullPath, files);
      continue;
    }

    files.push(path.relative(root, fullPath).split(path.sep).join("/"));
  }

  return files;
}

async function listRepoFiles(root) {
  const gitFiles = await listFilesFromGit(root);

  if (gitFiles !== null) {
    return gitFiles;
  }

  return walkFiles(root);
}

async function probeFile(root, file) {
  const result = await runCommand(
    oxfmtBinary,
    ["--check", "--no-error-on-unmatched-pattern", file],
    {
      cwd: root,
    },
  );

  const combined = `${result.stdout}\n${result.stderr}`;

  if (unsupportedByOxc(combined)) {
    return { status: "unsupported" };
  }

  if (result.code === 0) {
    return { status: "clean" };
  }

  if (result.code === 1) {
    return { status: "dirty" };
  }

  return {
    status: "error",
    message: result.stderr.trim() || result.stdout.trim() || "oxfmt failed",
  };
}

export async function inspectOxcFiles(root) {
  const repoFiles = await listRepoFiles(root);
  const candidates = repoFiles.filter((file) => path.extname(file) !== ".go");
  const supportedFiles = [];
  const dirtyFiles = [];
  const errors = [];

  for (const file of candidates) {
    const outcome = await probeFile(root, file);

    if (outcome.status === "unsupported") {
      continue;
    }

    supportedFiles.push(normalizePath(root, file));

    if (outcome.status === "dirty") {
      dirtyFiles.push(normalizePath(root, file));
      continue;
    }

    if (outcome.status === "error") {
      errors.push({
        file: normalizePath(root, file),
        message: outcome.message,
      });
    }
  }

  return {
    files: supportedFiles.sort(),
    dirtyFiles: dirtyFiles.sort(),
    errors,
  };
}

export async function writeOxcFiles(root, files) {
  if (files.length === 0) {
    return;
  }

  const result = await runCommand(
    oxfmtBinary,
    ["--write", "--no-error-on-unmatched-pattern", ...files],
    { cwd: root },
  );

  if (result.code !== 0) {
    throw new Error(result.stderr.trim() || result.stdout.trim() || "oxfmt failed");
  }
}

export async function runOxcOnly(mode, root = process.cwd()) {
  const inspection = await inspectOxcFiles(root);

  if (mode === "format") {
    await writeOxcFiles(root, inspection.dirtyFiles);
  }

  return inspection;
}
