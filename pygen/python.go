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

import (
	"path/filepath"
	"strings"
)

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

// Shadowed is the set of Python built-in names the generated code uses itself.
//
// An English name that matches one of these has to be escaped, because the
// generated Python calls the built-in: "Declare sum to be 5." emitted
// "sum = 5", and every later sum(...) in the file — the translation of
// English's own "sum of" — then failed with "int object is not callable".
// A user-defined function called "sum" broke the same way.
//
// The criterion is what the generated code emits, not the whole of Python's
// builtins: escaping a name the output never mentions would rename it for no
// reason. Every entry here appears in the two Python back-ends' output, in a
// helper definition, or in the module preamble.
var Shadowed = map[string]bool{
	"abs": true, "all": true, "any": true, "bool": true, "dict": true,
	"enumerate": true, "float": true, "hex": true, "id": true, "input": true,
	"int": true, "isinstance": true, "len": true, "list": true, "max": true,
	"min": true, "open": true, "print": true, "range": true, "reversed": true,
	"round": true, "sorted": true, "str": true, "sum": true, "zip": true,
	// Modules the preamble imports, which a name would shadow just as badly.
	"copy": true, "math": true, "os": true, "random": true, "sys": true,
	"time": true,
}

// SanitizeIdent escapes a name that Python cannot use as-is, by appending a
// trailing underscore following PEP 8 convention ("class" → "class_").
//
// It covers reserved words, which are a syntax error, and the built-ins the
// generated code relies on, which are a silent misbehaviour. A name beginning
// with a digit — which English allows nowhere, but bytecode could carry — also
// gets a leading underscore rather than producing unparseable Python.
func SanitizeIdent(name string) string {
	if name == "" {
		return "_"
	}
	if Keywords[name] || Shadowed[name] {
		return name + "_"
	}
	if name[0] >= '0' && name[0] <= '9' {
		return "_" + name
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

	// English rounds half away from zero, which is what math.Round does.
	// Python's round() rounds half to even, so round(2.5) was 3 in English and
	// 2 in the generated Python — the same program giving two answers.
	"_round": `def _round(x):
    return math.floor(x + 0.5) if x >= 0 else math.ceil(x - 0.5)`,

	// English takes the remainder of the truncated operands, with the sign of
	// the dividend. Python's % is floored and takes the sign of the divisor,
	// so -7 % 3 was -1 in English and 2 in the generated Python.
	"_remainder": `def _remainder(a, b):
    return math.fmod(math.trunc(a), math.trunc(b))`,

	// How a value is written out. English has one number type and prints a
	// whole one without a decimal point, writes true/false in lower case and
	// nothing for the absence of a value; Python writes 5.0, True and None. So
	// every program that printed a number, a boolean, an empty value or a list
	// of them produced different text once transpiled.
	"_show": `def _show(v):
    if v is None:
        return "nothing"
    if v is True:
        return "true"
    if v is False:
        return "false"
    if isinstance(v, float):
        if v != v or v in (float("inf"), float("-inf")):
            return repr(v)
        if v.is_integer() and abs(v) < 1e15:
            return str(int(v))
        text = repr(v)
        if "e" in text or "E" in text:
            text = format(v, "f").rstrip("0").rstrip(".")
        return text
    if isinstance(v, int):
        return str(v)
    if isinstance(v, str):
        return v
    if isinstance(v, dict):
        return "{" + ", ".join(_show(k) + ": " + _show(x) for k, x in v.items()) + "}"
    if isinstance(v, (list, tuple, range)):
        return "[" + ", ".join(_show(x) for x in v) + "]"
    return str(v)`,

	// English's "cast to boolean" reads the words a person would write and
	// refuses anything else. Python's bool() calls every non-empty string
	// true, so casting "no" gave False in English and True in the generated
	// Python.
	"_to_bool": `def _to_bool(v):
    if isinstance(v, bool):
        return v
    if v is None:
        return False
    if isinstance(v, str):
        text = v.strip().lower()
        if text in ("true", "yes", "1"):
            return True
        if text in ("false", "no", "0"):
            return False
        raise ValueError("cannot cast the text %r to a boolean" % v)
    return v != 0`,

	// The four below exist so that an argument is evaluated once. They were
	// inline expressions that named an argument two or three times — average
	// as "(sum(x) / len(x))", insert as "x[:i] + [v] + x[i:]" — so an
	// argument that did something as well as producing a value, such as
	// "the result of calling next_batch", did it twice.
	"_average": `def _average(lst):
    return sum(lst) / len(lst)`,

	"_insert": `def _insert(lst, index, item):
    index = int(index)
    return lst[:index] + [item] + lst[index:]`,

	"_remove_at": `def _remove_at(lst, index):
    index = int(index)
    return [v for i, v in enumerate(lst) if i != index]`,

	"_substring": `def _substring(text, start, length):
    start = int(start)
    return text[start:start + int(length)]`,

	// An English range includes its end, and runs downwards when the end is
	// below the start. Expressing that inline named start, end and step up to
	// three times each.
	"_range": `def _range(start, end, step=None):
    start = int(start)
    end = int(end)
    if step is None:
        step = 1 if start <= end else -1
    else:
        step = int(step)
    return range(start, end + (1 if step > 0 else -1), step)`,
}

// HelperOrder defines the deterministic emission order for helper functions.
//
// Every emitter walks this rather than its own set, so the same program always
// produces byte-identical Python. The bytecode back-end used to iterate a Go
// map, whose order is deliberately randomised, so its output differed run to
// run for any program needing two helpers.
var HelperOrder = []string{
	"_program_start",
	"_show",
	"_round",
	"_remainder",
	"_to_bool",
	"_range",
	"_table_remove",
	"_flatten",
	"_is_nan",
	"_is_infinite",
	"_sign",
	"_unique",
	"_product",
	"_zip_with",
	"_average",
	"_insert",
	"_remove_at",
	"_substring",
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

// ModuleName is the Python module name for an English source file, and also
// the base name of the .py file written for it.
//
// Both sides have to agree, and they did not: the file was named by stripping
// the last extension, while the import statement took everything before the
// *first* dot, so "my.lib.abc" was written as "my.lib.py" and imported as
// "my" — a module that does not exist. Neither is a legal module name anyway,
// since a dot in an import means a package boundary.
//
// Anything Python cannot spell in an identifier becomes an underscore, so one
// rule produces a name that is valid both as a file and as an import.
func ModuleName(path string) string {
	base := path
	if i := strings.LastIndexAny(base, `/\`); i >= 0 {
		base = base[i+1:]
	}
	if ext := filepath.Ext(base); ext != "" {
		base = strings.TrimSuffix(base, ext)
	}

	var b strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return SanitizeIdent(b.String())
}
