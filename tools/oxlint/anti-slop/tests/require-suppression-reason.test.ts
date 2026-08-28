import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('require-suppression-reason', () => {
	const results = lintCases(
		'require-suppression-reason',
		[
			{ name: 'ordinary comment', code: '// The pool size follows the host, not the input.\nexport const size = 1;\n' },
			{
				name: 'named rule with a reason',
				code: '// oxlint-disable-next-line anti-slop/no-unknown-parameters -- this is the boundary parser, and it must admit arbitrary payloads to reject them.\nexport function parse(payload: unknown): string {\n\treturn String(payload);\n}\n',
			},
			{ name: 'blanket disable', code: '// oxlint-disable\nexport const size = 1;\n' },
			{ name: 'named rule without a reason', code: '// oxlint-disable-next-line anti-slop/no-unknown-parameters\nexport function parse(payload: unknown): string {\n\treturn String(payload);\n}\n' },
			{
				name: 'placeholder reason',
				code: '// oxlint-disable-next-line anti-slop/no-unknown-parameters -- needed\nexport function parse(payload: unknown): string {\n\treturn String(payload);\n}\n',
			},
			{ name: 'directive for another linter', code: '// eslint-disable-next-line no-console -- a real reason, but for a linter this repo does not run.\nexport const size = 1;\n' },
		],
	);

	assert.deepEqual(results.get('ordinary comment'), []);
	assert.deepEqual(results.get('named rule with a reason'), []);

	for (const name of ['blanket disable', 'named rule without a reason', 'placeholder reason', 'directive for another linter']) {
		const reported = results.get(name);

		assert.equal(reported?.length, 1, name);
		assert.equal(reported?.[0]?.code, 'anti-slop(require-suppression-reason)', name);
		assert.equal(reported?.[0]?.labels[0]?.span.line, 1, name);
	}
});
