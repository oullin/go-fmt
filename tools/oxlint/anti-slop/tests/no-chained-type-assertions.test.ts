import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('no-chained-type-assertions', () => {
	const results = lintCases(
		'no-chained-type-assertions',
		[
			{ name: 'no assertion', code: 'export const total: number = 1;\n' },
			{ name: 'const assertion chain', code: 'export const modes = ["read", "write"] as const;\n' },
			{ name: 'double assertion through unknown', code: 'declare const raw: string;\n\n// SAFETY: fixture.\nexport const total = raw as unknown as number;\n' },
			{ name: 'parenthesized chain', code: 'declare const raw: string;\n\n// SAFETY: fixture.\nexport const total = (raw as unknown) as number;\n' },
		],
	);

	assert.deepEqual(results.get('no assertion'), []);
	assert.deepEqual(results.get('const assertion chain'), []);

	const reported = results.get('double assertion through unknown');

	assert.ok((reported?.length ?? 0) >= 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-chained-type-assertions)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 4);

	assert.ok((results.get('parenthesized chain')?.length ?? 0) >= 1);
});
