import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('no-shape-in-symbol-names', () => {
	const results = lintCases(
		'no-shape-in-symbol-names',
		[
			{ name: 'domain name', code: 'export const invoiceTotal = 1;\n' },
			{ name: 'shape in a binding', code: 'export const invoiceShape = 1;\n' },
			{ name: 'shape is matched case-insensitively', code: 'export const SHAPE_VERSION = 1;\n' },
			{ name: 'shape in a function name', code: 'export function readShape(): number {\n\treturn 1;\n}\n' },
		],
	);

	assert.deepEqual(results.get('domain name'), []);

	const reported = results.get('shape in a binding');

	assert.equal(reported?.length, 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-shape-in-symbol-names)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 1);

	assert.equal(results.get('shape is matched case-insensitively')?.length, 1);
	assert.equal(results.get('shape in a function name')?.length, 1);
});
