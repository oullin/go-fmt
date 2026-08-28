import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('no-widen-then-assert', () => {
	const results = lintCases(
		'no-widen-then-assert',
		[
			{ name: 'precise type survives', code: 'export function read(): string {\n\tconst owner = { id: "a" };\n\n\treturn owner.id;\n}\n' },
			{ name: 'named contract, no assertion', code: 'type Owner = { readonly id: string };\n\nexport function read(): Owner {\n\tconst owner: Owner = { id: "a" };\n\n\treturn owner;\n}\n' },
			{
				name: 'widened to unknown then asserted back',
				code: 'type Owner = { readonly id: string };\n\nexport function read(): Owner {\n\tconst owner: unknown = { id: "a" };\n\n\treturn owner as Owner;\n}\n',
			},
			{ name: 'widened at module scope', code: 'type Owner = { readonly id: string };\n\nconst owner: unknown = { id: "a" };\n\nexport const found = owner as Owner;\n' },
		],
	);

	assert.deepEqual(results.get('precise type survives'), []);
	assert.deepEqual(results.get('named contract, no assertion'), []);

	const reported = results.get('widened to unknown then asserted back');

	assert.equal(reported?.length, 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-widen-then-assert)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 6);

	assert.equal(results.get('widened at module scope')?.length, 1);
});

/**
 * The rule tracks a binding widened to `unknown`. `Record<string, unknown>` and
 * `object` are the neighbouring widenings it does not follow — those belong to
 * `no-known-value-widening` and `no-unsafe-dictionary-type`. Pinning the split
 * keeps a future change from silently moving the boundary.
 */
test('no-widen-then-assert leaves open-record and object widening to its neighbours', () => {
	const results = lintCases(
		'no-widen-then-assert',
		[
			{
				name: 'open record',
				code: 'type Owner = { readonly id: string };\n\nexport function read(): Owner {\n\tconst owner: Record<string, unknown> = { id: "a" };\n\n\treturn owner as Owner;\n}\n',
			},
			{ name: 'object', code: 'type Owner = { readonly id: string };\n\nexport function read(): Owner {\n\tconst owner: object = { id: "a" };\n\n\treturn owner as Owner;\n}\n' },
		],
	);

	assert.deepEqual(results.get('open record'), []);
	assert.deepEqual(results.get('object'), []);
});
