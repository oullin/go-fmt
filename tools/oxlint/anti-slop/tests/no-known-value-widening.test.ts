import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('no-known-value-widening', () => {
	const results = lintCases(
		'no-known-value-widening',
		[
			{ name: 'inference keeps the evidence', code: 'export const owner = { id: "a", total: 1 };\n' },
			{ name: 'named contract', code: 'type Owner = { readonly id: string };\n\nexport const owner: Owner = { id: "a" };\n' },
			{ name: 'widened to an open record', code: 'export const owner: Record<string, unknown> = { id: "a", total: 1 };\n' },
		],
	);

	assert.deepEqual(results.get('inference keeps the evidence'), []);
	assert.deepEqual(results.get('named contract'), []);

	const reported = results.get('widened to an open record');

	assert.ok((reported?.length ?? 0) >= 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-known-value-widening)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 1);
});
