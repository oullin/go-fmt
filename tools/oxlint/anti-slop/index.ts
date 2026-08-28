import { eslintCompatPlugin } from '@oxlint/plugins';

import { noAmbientNondeterminismRule } from '#anti-slop/rules/no-ambient-nondeterminism';
import { noChainedTypeAssertionsRule } from '#anti-slop/rules/no-chained-type-assertions';
import { noConditionalEmptyObjectSpreadRule } from '#anti-slop/rules/no-conditional-empty-object-spread';
import { noKnownValueWideningRule } from '#anti-slop/rules/no-known-value-widening';
import { noModuleMockingRule } from '#anti-slop/rules/no-module-mocking';
import { noObjectParametersRule } from '#anti-slop/rules/no-object-parameters';
import { noReflectApplyRule } from '#anti-slop/rules/no-reflect-apply';
import { noReflectGetRule } from '#anti-slop/rules/no-reflect-get';
import { noRuntimeTypeofRule } from '#anti-slop/rules/no-runtime-typeof';
import { noForbiddenTermInSymbolNamesRule } from '#anti-slop/rules/no-shape-in-symbol-names';
import { noUnknownParametersRule } from '#anti-slop/rules/no-unknown-parameters';
import { noUnknownReturnsRule } from '#anti-slop/rules/no-unknown-returns';
import { noUnknownTypeAliasesRule } from '#anti-slop/rules/no-unknown-type-aliases';
import { noUnsafeDictionaryTypeRule } from '#anti-slop/rules/no-unsafe-dictionary-type';
import { noWidenThenAssertRule } from '#anti-slop/rules/no-widen-then-assert';
import { requireSafetyCommentForTypeAssertionRule } from '#anti-slop/rules/require-safety-comment-for-type-assertion';
import { requireSuppressionReasonRule } from '#anti-slop/rules/require-suppression-reason';

/** Generic Oxlint rules that reject low-evidence and low-signal implementation patterns. */
const antiSlopPlugin = eslintCompatPlugin(
	{
		meta: { name: 'anti-slop' },
		rules: {
			'no-ambient-nondeterminism': noAmbientNondeterminismRule,
			'no-chained-type-assertions': noChainedTypeAssertionsRule,
			'no-conditional-empty-object-spread': noConditionalEmptyObjectSpreadRule,
			'no-known-value-widening': noKnownValueWideningRule,
			'no-module-mocking': noModuleMockingRule,
			'no-object-parameters': noObjectParametersRule,
			'no-reflect-apply': noReflectApplyRule,
			'no-reflect-get': noReflectGetRule,
			'no-runtime-typeof': noRuntimeTypeofRule,
			'no-unsafe-dictionary-type': noUnsafeDictionaryTypeRule,
			'no-shape-in-symbol-names': noForbiddenTermInSymbolNamesRule,
			'no-unknown-parameters': noUnknownParametersRule,
			'no-unknown-returns': noUnknownReturnsRule,
			'no-unknown-type-aliases': noUnknownTypeAliasesRule,
			'no-widen-then-assert': noWidenThenAssertRule,
			'require-safety-comment-for-type-assertion': requireSafetyCommentForTypeAssertionRule,
			'require-suppression-reason': requireSuppressionReasonRule,
		},
	},
);

export default antiSlopPlugin;
