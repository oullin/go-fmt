import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('no-ambient-nondeterminism', () => {
	const results = lintCases(
		'no-ambient-nondeterminism',
		[
			{ name: 'instant supplied by the caller', code: 'export function stamp(at: number): string {\n\treturn String(at);\n}\n' },
			{ name: 'date built from an argument', code: 'export function stamp(at: number): Date {\n\treturn new Date(at);\n}\n' },
			{ name: 'shadowed Math', code: 'const Math = { random: (): number => 0 };\n\nexport const pick = Math.random();\n' },
			{ name: 'Date.now', code: 'export function stamp(): string {\n\treturn String(Date.now());\n}\n' },
			{ name: 'bare new Date', code: 'export function stamp(): Date {\n\treturn new Date();\n}\n' },
			{ name: 'Math.random', code: 'export function pick(): number {\n\treturn Math.random();\n}\n' },
			{ name: 'computed Math access', code: 'export function pick(): number {\n\treturn Math["random"]();\n}\n' },
			{ name: 'performance.now', code: 'export function elapsed(): number {\n\treturn performance.now();\n}\n' },
			{ name: 'crypto.randomUUID', code: 'export function id(): string {\n\treturn crypto.randomUUID();\n}\n' },
		],
	);

	assert.deepEqual(results.get('instant supplied by the caller'), []);
	assert.deepEqual(results.get('date built from an argument'), []);
	assert.deepEqual(results.get('shadowed Math'), []);

	const reported = results.get('Date.now');

	assert.equal(reported?.length, 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-ambient-nondeterminism)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 2);

	for (const name of ['bare new Date', 'Math.random', 'computed Math access', 'performance.now', 'crypto.randomUUID']) {
		assert.equal(results.get(name)?.length, 1, name);
	}
});
