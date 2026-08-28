import { execFileSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

/**
 * These rules only exist inside oxlint. `@oxlint/plugins` ships no rule tester,
 * and every rule here uses `createOnce`, so `context.sourceCode` is reachable
 * only from a visitor callback after oxlint swaps the context prototype. The
 * real binary is therefore the only honest oracle, and each assertion runs it.
 */
const require = createRequire(import.meta.url);

/** The plugin under test, as oxlint's `jsPlugins` specifier wants it. */
const PLUGIN_ENTRY = fileURLToPath(new URL('../index.ts', import.meta.url));

/** Resolved through the package's own dependency, never the repo-root `.bin` shim. */
const OXLINT_BIN = join(
	dirname(
		require.resolve('oxlint/package.json'),
	),
	'bin',
	'oxlint',
);

/** One source file handed to the linter under a name the assertions can refer back to. */
export type RuleCase = {
	readonly name: string;
	readonly code: string;
};

/** A single oxlint diagnostic, narrowed to the fields assertions read. */
export type Diagnostic = {
	readonly message: string;
	readonly code?: string;
	readonly filename: string;
	readonly labels: readonly { readonly span: { readonly line: number; readonly column: number; readonly offset: number } }[];
};

type OxlintOutput = {
	readonly diagnostics: readonly Diagnostic[];
};

/**
 * Lint every case against exactly one rule.
 *
 * One rule per run is not a convenience: these detectors overlap heavily, and
 * a case written for `no-widen-then-assert` also trips `no-known-value-widening`
 * and `require-safety-comment-for-type-assertion`. Isolation keeps each
 * assertion about the rule it names.
 *
 * @param rule - The rule id without the `anti-slop/` prefix.
 * @param cases - Sources to lint, each under a name the result is keyed by.
 * @param options - Rule options, when the rule takes any.
 * @returns Diagnostics per case name, in source order.
 */
export function lintCases(rule: string, cases: readonly RuleCase[], options?: unknown): ReadonlyMap<string, readonly Diagnostic[]> {
	const directory = mkdtempSync(
		join(
			tmpdir(),
			'anti-slop-',
		),
	);

	try {
		mkdirSync(
			join(directory, 'cases'),
		);

		for (const [index, ruleCase] of cases.entries()) {
			writeFileSync(
				join(directory, 'cases', `${index}.ts`),
				ruleCase.code,
			);
		}

		writeFileSync(
			join(directory, '.oxlintrc.json'),
			JSON.stringify(configFor(rule, options)),
		);

		return groupByCase(
			cases,
			runOxlint(directory),
		);
	} finally {
		rmSync(
			directory,
			{ recursive: true, force: true },
		);
	}
}

/**
 * Build a config that enables one rule and silences everything else.
 *
 * `categories: {}` is not enough — built-in correctness rules still report as
 * warnings and pollute the diagnostics a case is asserting on.
 *
 * @param rule - The rule id without the `anti-slop/` prefix.
 * @param options - Rule options, when the rule takes any.
 * @returns The oxlint configuration to write beside the cases.
 */
function configFor(rule: string, options: unknown): Record<string, unknown> {
	return {
		plugins: [],
		categories: { correctness: 'allow' },
		jsPlugins: [{ name: 'anti-slop', specifier: PLUGIN_ENTRY }],
		rules: { [`anti-slop/${rule}`]: options === undefined ? 'error' : ['error', options] },
	};
}

/**
 * Run the linter over a prepared case directory.
 *
 * oxlint exits non-zero whenever it reports anything, which is the normal path
 * here, so the diagnostics are read off the failure as readily as the success.
 *
 * @param directory - The temporary root holding `.oxlintrc.json` and `cases/`.
 * @returns Every diagnostic oxlint produced.
 */
function runOxlint(directory: string): readonly Diagnostic[] {
	const argv = ['-c', join(directory, '.oxlintrc.json'), '--disable-nested-config', '--no-ignore', '-f', 'json', join(directory, 'cases')];

	let stdout: string;

	try {
		stdout = execFileSync(
			OXLINT_BIN,
			argv,
			{ encoding: 'utf8' },
		);
	} catch (cause) {
		const reported = (cause as { stdout?: string }).stdout;

		if (reported === undefined) {
			throw cause;
		}

		stdout = reported;
	}

	const parsed = JSON.parse(stdout) as OxlintOutput;

	return parsed.diagnostics.filter((diagnostic) => {
		return diagnostic.code?.startsWith('anti-slop(') === true;
	});
}

/**
 * Attribute diagnostics back to the case that produced them.
 *
 * oxlint lints files across a thread pool, so neither file order nor the order
 * within a file survives the run. Sorting by offset makes line assertions
 * deterministic.
 *
 * @param cases - The cases in the order they were written out.
 * @param diagnostics - Every anti-slop diagnostic from the run.
 * @returns Diagnostics per case name, in source order.
 */
function groupByCase(cases: readonly RuleCase[], diagnostics: readonly Diagnostic[]): ReadonlyMap<string, readonly Diagnostic[]> {
	const collected = new Map<string, Diagnostic[]>(
		cases.map((ruleCase) => {
			return [ruleCase.name, []];
		}),
	);

	for (const diagnostic of diagnostics) {
		const index = Number(
			diagnostic.filename.split('/')
				.at(-1)
				?.replace('.ts', ''),
		);

		const ruleCase = cases[index];

		if (ruleCase === undefined) {
			throw new Error(`diagnostic in an unknown file: ${diagnostic.filename}`);
		}

		collected.get(ruleCase.name)?.push(diagnostic);
	}

	for (const bucket of collected.values()) {
		bucket.sort((left, right) => {
			return (left.labels[0]?.span.offset ?? 0) - (right.labels[0]?.span.offset ?? 0);
		});
	}

	return collected;
}
