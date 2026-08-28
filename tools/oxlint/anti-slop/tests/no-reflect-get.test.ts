import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('no-reflect-get', () => {
	const results = lintCases(
		'no-reflect-get',
		[
			{ name: 'typed property access', code: 'export function read(owner: { name: string }): string {\n\treturn owner.name;\n}\n' },
			{ name: 'unrelated get method', code: 'export function read(owner: Map<string, string>): string | undefined {\n\treturn owner.get("name");\n}\n' },
			{ name: 'Reflect.get', code: 'export function read(owner: { name: string }): unknown {\n\treturn Reflect.get(owner, "name");\n}\n' },
			{ name: 'computed Reflect access', code: 'export function read(owner: { name: string }): unknown {\n\treturn Reflect["get"](owner, "name");\n}\n' },
		],
	);

	assert.deepEqual(results.get('typed property access'), []);
	assert.deepEqual(results.get('unrelated get method'), []);

	const reported = results.get('Reflect.get');

	assert.equal(reported?.length, 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-reflect-get)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 2);

	assert.equal(results.get('computed Reflect access')?.length, 1);
});
