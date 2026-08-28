import assert from 'node:assert/strict';
import { test } from 'node:test';

import { lintCases } from '#anti-slop/tests/rule-tester';

test('require-safety-comment-for-type-assertion', () => {
	const results = lintCases(
		'require-safety-comment-for-type-assertion',
		[
			{ name: 'no assertion', code: 'export const total: number = 1;\n' },
			{ name: 'const assertion needs no justification', code: 'export const modes = ["read", "write"] as const;\n' },
			{ name: 'justified declaration', code: 'declare const raw: string | number;\n\n// SAFETY: the caller has already parsed this as text.\nconst text = raw as string;\n\nexport { text };\n' },
			{
				name: 'justified return',
				code: 'declare const raw: string | number;\n\nexport function read(): string {\n\t// SAFETY: the caller has already parsed this as text.\n\treturn raw as string;\n}\n',
			},
			{ name: 'unjustified assertion', code: 'declare const raw: string | number;\n\nexport const text = raw as string;\n' },
		],
	);

	assert.deepEqual(results.get('no assertion'), []);
	assert.deepEqual(results.get('const assertion needs no justification'), []);
	assert.deepEqual(results.get('justified declaration'), []);
	assert.deepEqual(results.get('justified return'), []);

	const reported = results.get('unjustified assertion');

	assert.ok((reported?.length ?? 0) >= 1);
	assert.equal(reported?.[0]?.code, 'anti-slop(require-safety-comment-for-type-assertion)');
	assert.equal(reported?.[0]?.labels[0]?.span.line, 3);
});

/**
 * `hasSafetyComment` walks up until it reaches a `commentOwnerKinds` node, and
 * that set holds `VariableDeclaration` but not `ExportNamedDeclaration`. A
 * comment written above `export const` therefore attaches to the export, which
 * the walk never inspects, and the assertion reads as unjustified.
 *
 * This asserts the boundary as it stands rather than the behavior we would
 * want, so that widening the set — here or upstream — fails loudly instead of
 * passing unnoticed.
 */
test('require-safety-comment-for-type-assertion does not see a comment above an exported declaration', () => {
	const results = lintCases(
		'require-safety-comment-for-type-assertion',
		[
			{ name: 'justified export', code: 'declare const raw: string | number;\n\n// SAFETY: the caller has already parsed this as text.\nexport const text = raw as string;\n' },
		],
	);

	assert.equal(results.get('justified export')?.length, 1);
});
