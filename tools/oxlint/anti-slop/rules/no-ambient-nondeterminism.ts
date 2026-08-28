import { defineRule } from '@oxlint/plugins';

import type { ESTree, SourceCode } from '@oxlint/plugins';

import { isGlobalMember, isGlobalNamed } from '#anti-slop/shared/global-member';

/** Global readings of the clock and the entropy source, by owner and property. */
const AMBIENT_MEMBERS: readonly (readonly [string, string, string])[] = [
	['Date', 'now', 'Date.now()'],
	['Math', 'random', 'Math.random()'],
	['performance', 'now', 'performance.now()'],
	['crypto', 'randomUUID', 'crypto.randomUUID()'],
	['crypto', 'getRandomValues', 'crypto.getRandomValues()'],
	['process', 'hrtime', 'process.hrtime()'],
];

/**
 * Name the ambient source a call reads, if it reads one.
 *
 * @param sourceCode - The source under analysis.
 * @param callee - The call's target.
 * @returns A display name for the source, or null when the call is deterministic.
 */
function ambientCallSource(sourceCode: SourceCode, callee: ESTree.Expression): string | null {
	for (const [owner, property, display] of AMBIENT_MEMBERS) {
		if (isGlobalMember(sourceCode, callee, owner, property)) {
			return display;
		}
	}

	return null;
}

/**
 * Require a formatter's output to depend on its input alone.
 *
 * fmtkit's contract is that the same bytes format to the same bytes, on any
 * machine, at any time. A pass that reads the clock or the entropy source
 * breaks that quietly: the output is still plausible, just no longer a
 * function of the input, and only an intermittent test failure reveals it.
 * Where a timestamp or an id is genuinely needed, the caller supplies it.
 */
export const noAmbientNondeterminismRule = defineRule(
	{
		meta: {
			type: 'problem',
			docs: {
				description: 'Disallow reading the ambient clock or entropy source; deterministic code takes time and randomness from its caller.',
			},
			messages: {
				ambientCall: '`{{source}}` makes this output depend on when and where it ran. Take the value as a parameter so the caller owns it.',
				ambientDate: '`new Date()` without an argument reads the ambient clock. Take the instant as a parameter so the caller owns it.',
			},
		},
		createOnce(context) {
			return {
				CallExpression(node) {
					if (node.callee.type === 'Super' || node.callee.type === 'V8IntrinsicExpression') {
						return;
					}

					const source = ambientCallSource(context.sourceCode, node.callee);

					if (source !== null) {
						context.report({ node, messageId: 'ambientCall', data: { source } });
					}
				},
				NewExpression(node) {
					if (node.arguments.length === 0 && isGlobalNamed(context.sourceCode, node.callee, 'Date')) {
						context.report({ node, messageId: 'ambientDate' });
					}
				},
			};
		},
	},
);
