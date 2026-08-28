import assert from 'node:assert/strict';
import { readdirSync } from 'node:fs';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

import antiSlopPlugin from '#anti-slop/index';

/**
 * `index.ts` maps `no-shape-in-symbol-names` to `noForbiddenTermInSymbolNamesRule`,
 * so a rule file and its registered id can drift apart without either the
 * typechecker or a per-rule test noticing. Only this comparison catches that.
 */
test('every rule file is registered, and every registered rule has a file', () => {
	const ruleFiles = readdirSync(
		fileURLToPath(new URL('../rules', import.meta.url)),
	)
		.filter((entry) => {
			return entry.endsWith('.ts');
		})
		.map((entry) => {
			return entry.slice(0, -'.ts'.length);
		})
		.sort();

	assert.deepEqual(Object.keys(antiSlopPlugin.rules).sort(), ruleFiles);
});

test('every registered rule has a test file', () => {
	const tested = readdirSync(
		fileURLToPath(new URL('.', import.meta.url)),
	)
		.filter((entry) => {
			return entry.endsWith('.test.ts') && entry !== 'registration.test.ts';
		})
		.map((entry) => {
			return entry.slice(0, -'.test.ts'.length);
		})
		.sort();

	assert.deepEqual(Object.keys(antiSlopPlugin.rules).sort(), tested);
});
