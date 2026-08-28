import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

const NARROWING = 'export function read(value: string | number): string {\n\tif (typeof value === "string") {\n\t\treturn value;\n\t}\n\n\treturn String(value);\n}\n';

const TYPE_GUARD = 'export function isText(value: string | number): value is string {\n\treturn typeof value === "string";\n}\n';

test('no-runtime-typeof', () => {
	const results = lintCases(
		'no-runtime-typeof',
		[
			{ name: 'domain branch', code: 'export function read(value: { kind: "text"; body: string }): string {\n\treturn value.body;\n}\n' },
			{ name: 'runtime narrowing', code: NARROWING },
		],
	);

	assert.deepEqual(results.get('domain branch'), []);

	const reported = results.get('runtime narrowing');

	assert.equal(reported?.length, 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-runtime-typeof)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 2);
});

test('no-runtime-typeof allowInTypeGuards changes the verdict', () => {
	const guarded = [{ name: 'type guard', code: TYPE_GUARD }];

	assert.equal(lintCases(
		'no-runtime-typeof',
		guarded,
		{ allowInTypeGuards: false },
	).get('type guard')?.length, 1);
	assert.deepEqual(lintCases(
		'no-runtime-typeof',
		guarded,
		{ allowInTypeGuards: true },
	).get('type guard'), []);
});
