import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('no-module-mocking', () => {
	const results = lintCases(
		'no-module-mocking',
		[
			{ name: 'injected fake', code: 'export function build(clock: { now: () => number }): number {\n\treturn clock.now();\n}\n' },
			{ name: 'vitest module mock', code: 'import { vi } from "vitest";\n\nvi.mock("#app/clock");\n' },
			{ name: 'jest module mock', code: 'import { jest } from "@jest/globals";\n\njest.mock("#app/clock");\n' },
		],
	);

	assert.deepEqual(results.get('injected fake'), []);

	const reported = results.get('vitest module mock');

	assert.equal(reported?.length, 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(no-module-mocking)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 3);

	assert.equal(results.get('jest module mock')?.length, 1);
});
