import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('no-reflect-apply', () => {
	const results = lintCases(
		'no-reflect-apply',
		[
			{ name: 'direct call', code: 'export const value = ((a: number): number => a)(1);\n' },
			{ name: 'unrelated member call', code: 'export const value = Math.max(1, 2);\n' },
			{ name: 'local Reflect binding', code: 'const Reflect = { apply: (): number => 1 };\n\nexport const value = Reflect.apply();\n' },
			{ name: 'Reflect.apply', code: 'export function call(fn: (a: number) => number, a: number): number {\n\treturn Reflect.apply(fn, undefined, [a]);\n}\n' },
			{ name: 'computed Reflect access', code: 'export function call(fn: (a: number) => number, a: number): number {\n\treturn Reflect["apply"](fn, undefined, [a]);\n}\n' },
		],
	);

	assert.deepEqual(results.get('direct call'), []);
	assert.deepEqual(results.get('unrelated member call'), []);
	assert.deepEqual(results.get('local Reflect binding'), []);

	const reported = results.get('Reflect.apply');

	assert.equal(reported?.length, 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-reflect-apply)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 2);

	assert.equal(results.get('computed Reflect access')?.length, 1);
});
