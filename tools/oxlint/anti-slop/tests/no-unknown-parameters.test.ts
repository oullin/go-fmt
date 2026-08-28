import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('no-unknown-parameters', () => {
	const results = lintCases(
		'no-unknown-parameters',
		[
			{ name: 'domain parameter', code: 'export function widen(value: string): number {\n\treturn value.length;\n}\n' },
			{ name: 'cause is the sanctioned exception', code: 'export function wrap(cause: unknown): Error {\n\treturn new Error("failed", { cause });\n}\n' },
			{ name: 'unknown parameter', code: 'export function decode(payload: unknown): string {\n\treturn String(payload);\n}\n' },
			{ name: 'unknown arrow parameter', code: 'export const decode = (payload: unknown): string => String(payload);\n' },
		],
	);

	assert.deepEqual(results.get('domain parameter'), []);
	assert.deepEqual(results.get('cause is the sanctioned exception'), []);

	const reported = results.get('unknown parameter');

	assert.equal(reported?.length, 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-unknown-parameters)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 1);

	assert.equal(results.get('unknown arrow parameter')?.length, 1);
});
