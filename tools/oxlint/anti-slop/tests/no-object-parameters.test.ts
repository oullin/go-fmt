import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('no-object-parameters', () => {
	const results = lintCases(
		'no-object-parameters',
		[
			{ name: 'named contract', code: 'export function read(owner: { readonly id: string }): string {\n\treturn owner.id;\n}\n' },
			{ name: 'object parameter', code: 'export function read(owner: object): string {\n\treturn String(owner);\n}\n' },
			{ name: 'alias of object', code: 'type Bag = object;\n\nexport function read(owner: Bag): string {\n\treturn String(owner);\n}\n' },
		],
	);

	assert.deepEqual(results.get('named contract'), []);

	const reported = results.get('object parameter');

	assert.equal(reported?.length, 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-object-parameters)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 1);

	assert.ok((results.get('alias of object')?.length ?? 0) >= 1);
});
