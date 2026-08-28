import { defineRule } from '@oxlint/plugins';

import type { ESTree } from '@oxlint/plugins';

/** Every directive form oxlint honors, whatever it suppresses and for how long. */
const DIRECTIVE = /^\s*oxlint-(disable|disable-next-line|disable-line)\b(?<body>[\s\S]*)$/u;

/**
 * A directive aimed at another engine. oxlint reads `eslint-disable` too, so
 * these silence real rules while claiming to configure a linter this repo does
 * not run.
 */
const FOREIGN_DIRECTIVE = /^\s*(eslint|biome|tslint|prettier)-(disable|ignore)\b/u;

/** Splits a directive body into the rules it names and the reason after `--`. */
const REASON_SEPARATOR = ' -- ';

/** A reason has to say something; a word or two is a placeholder, not a justification. */
const MINIMUM_REASON_LENGTH = 16;

type SuppressionFault = 'foreign' | 'noRules' | 'noReason' | 'thinReason';

/**
 * Judge one comment's suppression hygiene.
 *
 * @param comment - The comment to inspect.
 * @returns The fault to report, or null when the comment is fine.
 */
function faultFor(comment: ESTree.Comment): SuppressionFault | null {
	if (FOREIGN_DIRECTIVE.test(comment.value)) {
		return 'foreign';
	}

	const match = DIRECTIVE.exec(comment.value);

	if (match === null) {
		return null;
	}

	const body = match.groups?.['body'] ?? '';
	const separator = body.indexOf(REASON_SEPARATOR);

	if (separator === -1) {
		return body.trim() === '' ? 'noRules' : 'noReason';
	}

	if (body.slice(0, separator).trim() === '') {
		return 'noRules';
	}

	return body.slice(separator + REASON_SEPARATOR.length).trim().length < MINIMUM_REASON_LENGTH ? 'thinReason' : null;
}

/** Require every lint suppression to name what it silences and say why. */
export const requireSuppressionReasonRule = defineRule(
	{
		meta: {
			type: 'problem',
			docs: {
				description: 'Require oxlint suppressions to name their rules and carry a `--` justification, and reject directives for another linter.',
			},
			messages: {
				foreign: 'This directive targets another linter. Write `oxlint-disable-next-line <rule> -- <reason>`, or delete it.',
				noRules: 'This suppression names no rule, so it silences every future diagnostic on its target. Name the rules it is meant to silence.',
				noReason: 'This suppression has no justification. Append `-- <reason>` explaining the invariant that makes it safe.',
				thinReason: "This suppression's justification is too short to be one. State the invariant that makes silencing the rule correct.",
			},
		},
		createOnce(context) {
			return {
				Program() {
					for (const comment of context.sourceCode.getAllComments()) {
						const fault = faultFor(comment);

						if (fault !== null) {
							context.report({ node: comment, messageId: fault });
						}
					}
				},
			};
		},
	},
);
