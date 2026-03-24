function violationCount(report) {
  return report.results.reduce((total, result) => total + result.violations.length, 0);
}

function toJSONReport(report) {
  return {
    result: report.result,
    files: report.files,
    changed: report.changed,
    results: report.results.length > 0 ? report.results : undefined,
    errors: report.errors.length > 0 ? report.errors : undefined,
  };
}

function toAgentReport(report) {
  const changed = report.results
    .filter((result) => result.changed)
    .map((result) => ({
      file: result.file,
      steps: result.applied,
    }));
  const violations = report.results.flatMap((result) =>
    result.violations.map((violation) => ({
      file: result.file,
      rule: violation.rule,
      line: violation.line || undefined,
      message: violation.message,
    })),
  );

  return {
    result: report.result,
    summary: {
      files: report.files,
      changed: report.changed,
      violations: violationCount(report),
    },
    changed: changed.length > 0 ? changed : undefined,
    violations: violations.length > 0 ? violations : undefined,
    errors: report.errors.length > 0 ? report.errors : undefined,
  };
}

function renderText(report, mode) {
  if (report.files === 0) {
    return "No supported files found.\n";
  }

  const lines = [];
  const action = mode === "format" ? "Formatted" : "Checked";
  lines.push(`${action} ${report.files} file(s).`);

  for (const result of report.results) {
    if (result.changed && result.violations.length === 0) {
      lines.push(`~ ${result.file}`);
    }

    for (const violation of result.violations) {
      if (violation.line) {
        lines.push(`~ ${result.file}:${violation.line} [${violation.rule}] ${violation.message}`);
      } else {
        lines.push(`~ ${result.file} [${violation.rule}] ${violation.message}`);
      }
    }

    if (result.changed) {
      const verb = mode === "format" ? "applied" : "would apply";
      lines.push(`  ${verb} ${result.applied.join(", ")}`);
    }
  }

  for (const error of report.errors) {
    lines.push(`! ${error.file} ${error.message}`);
  }

  lines.push(
    `Result: ${report.result}. ${report.changed} changed, ${violationCount(report)} violation(s), ${report.errors.length} error(s).`,
  );

  return `${lines.join("\n")}\n`;
}

export function renderReport(report, outputFormat, mode) {
  switch (outputFormat) {
    case "text":
      return renderText(report, mode);
    case "json":
      return `${JSON.stringify(toJSONReport(report))}\n`;
    case "agent":
      return `${JSON.stringify(toAgentReport(report), null, 2)}\n`;
    default:
      throw new Error(`unsupported output format: ${outputFormat}`);
  }
}
