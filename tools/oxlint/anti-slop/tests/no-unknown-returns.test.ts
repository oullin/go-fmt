import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('no-unknown-returns', () => {
	const results = lintCases(
		'no-unknown-returns',
		[
			{ name: 'domain return', code: 'export function read(): string {\n\treturn "value";\n}\n' },
			{ name: 'promise of a domain type', code: 'export async function read(): Promise<string> {\n\treturn "value";\n}\n' },
			{ name: 'unknown return', code: 'export function read(): unknown {\n\treturn "value";\n}\n' },
			{ name: 'promise of unknown', code: 'export async function read(): Promise<unknown> {\n\treturn "value";\n}\n' },
		],
	);

	assert.deepEqual(results.get('domain return'), []);
	assert.deepEqual(results.get('promise of a domain type'), []);

	const reported = results.get('unknown return');

	assert.equal(reported?.length, 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-unknown-returns)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 1);

	assert.equal(results.get('promise of unknown')?.length, 1);
});
