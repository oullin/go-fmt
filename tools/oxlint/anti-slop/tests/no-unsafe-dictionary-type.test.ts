import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('no-unsafe-dictionary-type', () => {
	const results = lintCases(
		'no-unsafe-dictionary-type',
		[
			{ name: 'concrete value type', code: 'export type Totals = Record<string, number>;\n' },
			{ name: 'unknown value type', code: 'export type Totals = Record<string, unknown>;\n' },
			{ name: 'index signature of unknown', code: 'export type Totals = { [key: string]: unknown };\n' },
			{ name: 'union containing unknown', code: 'export type Totals = Record<string, number | unknown>;\n' },
		],
	);

	assert.deepEqual(results.get('concrete value type'), []);

	const reported = results.get('unknown value type');

	assert.equal(reported?.length, 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-unsafe-dictionary-type)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 1);

	assert.ok((results.get('index signature of unknown')?.length ?? 0) >= 1);
	assert.ok((results.get('union containing unknown')?.length ?? 0) >= 1);
});
