import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('no-unknown-type-aliases', () => {
	const results = lintCases(
		'no-unknown-type-aliases',
		[
			{ name: 'domain alias', code: 'export type Payload = { readonly id: string };\n' },
			{ name: 'alias of unknown', code: 'export type Payload = unknown;\n' },
			{ name: 'alias of an unknown alias', code: 'type Raw = unknown;\n\nexport type Payload = Raw;\n' },
		],
	);

	assert.deepEqual(results.get('domain alias'), []);

	const reported = results.get('alias of unknown');

	assert.equal(reported?.length, 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-unknown-type-aliases)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 1);

	assert.ok((results.get('alias of an unknown alias')?.length ?? 0) >= 1);
});
