// Package pygen holds the parts of Python generation that are the same
// whichever English representation is being translated.
//
// Two back-ends emit Python: one from the AST and one by decompiling bytecode.
// They each had their own copy of the reserved-word list, the identifier
// escaping rule and the injected helper definitions, and the copies had
// already drifted — the decompiler's helper table was missing an entry the
// transpiler had, so a program using it decompiled to Python that referred to
// something never defined.
package pygen

// pythonKeywords is the set of Python reserved words that cannot be used as
// bare identifiers. Any English identifier that matches a keyword is suffixed
// with an underscore (PEP 8 convention, e.g. "class" → "class_").
var Keywords = map[string]bool{
	"False": true, "None": true, "True": true,
	"and": true, "as": true, "assert": true, "async": true, "await": true,
	"break": true, "class": true, "continue": true, "def": true, "del": true,
	"elif": true, "else": true, "except": true, "finally": true, "for": true,
	"from": true, "global": true, "if": true, "import": true, "in": true,
	"is": true, "lambda": true, "nonlocal": true, "not": true, "or": true,
	"pass": true, "raise": true, "return": true, "try": true, "type": true,
	"while": true, "with": true, "yield": true,
}

// sanitizeIdent escapes a Python reserved word used as a user-defined identifier
// by appending a trailing underscore, following PEP 8 conventions.
func SanitizeIdent(name string) string {
	if Keywords[name] {
		return name + "_"
	}
	return name
}

// ─── Python helper function definitions ──────────────────────────────────────
//
// These small Python functions are injected at the top of the generated file
// when the corresponding English stdlib call is used and there is no single
// Python expression that exactly reproduces the behaviour.

// HelperDefs maps a helper name to its Python source (no trailing newline).
var HelperDefs = map[string]string{
	"_program_start": "_program_start = time.time()",

	"_table_remove": `def _table_remove(d, k):
    result = dict(d)
    result.pop(k, None)
    return result`,

	"_flatten": `def _flatten(lst):
    return [item for sublist in lst for item in sublist]`,

	"_is_nan": `def _is_nan(x):
    try:
        return math.isnan(float(x))
    except (TypeError, ValueError):
        return True`,

	"_is_infinite": `def _is_infinite(x):
    try:
        return math.isinf(float(x))
    except (TypeError, ValueError):
        return False`,

	"_sign": `def _sign(x):
    if x > 0:
        return 1
    elif x < 0:
        return -1
    return 0`,

	"_unique": `def _unique(lst):
    seen = []
    for item in lst:
        if item not in seen:
            seen.append(item)
    return seen`,

	"_product": `def _product(lst):
    result = 1
    for item in lst:
        result *= item
    return result`,

	"_zip_with": `def _zip_with(a, b):
    return [[x, y] for x, y in zip(a, b)]`,
}

// HelperOrder defines the deterministic emission order for helper functions.
var HelperOrder = []string{
	"_program_start",
	"_table_remove",
	"_flatten",
	"_is_nan",
	"_is_infinite",
	"_sign",
	"_unique",
	"_product",
	"_zip_with",
}

// MathConstants maps the English math constants to their Python equivalents.
var MathConstants = map[string]string{
	"pi":       "math.pi",
	"e":        "math.e",
	"infinity": "math.inf",
}

// ReservedWord reports whether a name is a Python reserved word, so that a
// caller can decide to escape it.
func ReservedWord(name string) bool { return Keywords[name] }
