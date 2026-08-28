import type { ESTree, Scope, SourceCode, Variable } from '@oxlint/plugins';

function resolveVariable(sourceCode: SourceCode, identifier: ESTree.IdentifierReference): Variable | null {
	let scope: Scope | null = sourceCode.getScope(identifier);

	while (scope !== null) {
		const variable = scope.set.get(identifier.name);

		if (variable !== undefined) {
			return variable;
		}

		scope = scope.upper;
	}

	return null;
}

/**
 * Report whether an identifier names an unshadowed global.
 *
 * A local `const Math = ...` is a different object with different behavior, so
 * a rule about the global one must not claim it.
 *
 * @param sourceCode - The source under analysis.
 * @param expression - The expression in the object position.
 * @param name - The global's name.
 * @returns True when the expression resolves to that global.
 */
export function isGlobalNamed(sourceCode: SourceCode, expression: ESTree.Expression, name: string): boolean {
	if (expression.type !== 'Identifier' || expression.name !== name) {
		return false;
	}

	if (sourceCode.isGlobalReference(expression)) {
		return true;
	}

	const variable = resolveVariable(sourceCode, expression);

	return variable === null || variable.defs.length === 0;
}

/**
 * Report whether a member expression reads one property off an unshadowed global.
 *
 * Computed access is covered so that `Math["random"]` cannot slip past a rule
 * that only reads dotted names.
 *
 * @param sourceCode - The source under analysis.
 * @param expression - The member expression to inspect.
 * @param globalName - The global the member must be read from.
 * @param property - The property name to match.
 * @returns True when the expression reads that property off that global.
 */
export function isGlobalMember(sourceCode: SourceCode, expression: ESTree.Expression, globalName: string, property: string): boolean {
	if (!('property' in expression) || !('object' in expression) || !('computed' in expression)) {
		return false;
	}

	if (!isGlobalNamed(sourceCode, expression.object, globalName)) {
		return false;
	}

	const member = expression.property;

	return expression.computed ? member.type === 'Literal' && member.value === property : member.type === 'Identifier' && member.name === property;
}
