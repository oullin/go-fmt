import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('no-conditional-empty-object-spread', () => {
	const results = lintCases(
		'no-conditional-empty-object-spread',
		[
			{ name: 'unconditional spread', code: 'declare const base: { id: string };\n\nexport const record = { ...base, active: true };\n' },
			{ name: 'conditional empty spread', code: 'declare const flag: boolean;\ndeclare const extra: { id: string };\n\nexport const record = { ...(flag ? extra : {}) };\n' },
			{ name: 'conditional empty spread on the left', code: 'declare const flag: boolean;\ndeclare const extra: { id: string };\n\nexport const record = { ...(flag ? {} : extra) };\n' },
		],
	);

	assert.deepEqual(results.get('unconditional spread'), []);

	const reported = results.get('conditional empty spread');

	assert.equal(reported?.length, 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-conditional-empty-object-spread)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 4);

	assert.equal(results.get('conditional empty spread on the left')?.length, 1);
});
