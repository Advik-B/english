package stdlib

import (
	"fmt"
	"sort"

	vm "github.com/Advik-B/english/astvm"
	"github.com/Advik-B/english/types"
)

// Param describes one parameter of a built-in function.
type Param struct {
	// Name is used in registration metadata and in usage hints.
	Name string
	// Kinds are the types accepted in this position. An empty list means any
	// type is accepted, which is how the genuinely polymorphic slots are
	// expressed: the default value of get_or_default, or the item of append.
	Kinds []types.TypeKind
}

// Accepts reports whether a value of the given kind may be passed here.
// An unknown actual kind is accepted, because the checker only rejects what it
// can prove wrong.
func (p Param) Accepts(actual types.TypeKind) bool {
	if len(p.Kinds) == 0 || actual == types.TypeUnknown {
		return true
	}
	for _, k := range p.Kinds {
		if types.Canonical(k) == types.Canonical(actual) {
			return true
		}
	}
	return false
}

// Describe renders the accepted types for a diagnostic, e.g. "list or array".
func (p Param) Describe() string {
	switch len(p.Kinds) {
	case 0:
		return "any value"
	case 1:
		return types.Name(p.Kinds[0])
	}
	out := ""
	for i, k := range p.Kinds {
		switch {
		case i == 0:
		case i == len(p.Kinds)-1:
			out += " or "
		default:
			out += ", "
		}
		out += types.Name(k)
	}
	return out
}

// Signature describes a built-in function's calling convention and types.
//
// This table is the single source of truth for which built-ins exist, how many
// arguments they take, and what types those arguments and the result have.
// Registration, dispatch, arity checking and the type checker all derive from
// it. The type checker used to keep its own copy of the argument types, which
// covered only 54 of the 84 built-ins, described only the first argument of
// each, and contradicted the implementations: it declared sum, first, last and
// count list-only while the code has always accepted arrays too, so summing an
// array was reported as a compile error even though it ran correctly.
type Signature struct {
	// Name is the function name as written in English source.
	Name string
	// Params are the parameters in order.
	Params []Param
	// Optional is the number of trailing parameters that may be omitted.
	Optional int
	// Returns is the result type, or types.TypeUnknown when it depends on the
	// arguments: first and last return an element of the collection, and
	// get_or_default returns whichever of its arguments applies.
	Returns types.TypeKind
	// Module selects the implementation dispatcher.
	Module string
}

// MinArgs is the smallest legal argument count.
func (s Signature) MinArgs() int { return len(s.Params) - s.Optional }

// MaxArgs is the largest legal argument count.
func (s Signature) MaxArgs() int { return len(s.Params) }

// ParamNames returns the parameter names in order.
func (s Signature) ParamNames() []string {
	names := make([]string, len(s.Params))
	for i, p := range s.Params {
		names[i] = p.Name
	}
	return names
}

// ─── Table construction helpers ──────────────────────────────────────────────

// param names a parameter and the kinds it accepts.
func param(name string, kinds ...types.TypeKind) Param {
	return Param{Name: name, Kinds: kinds}
}

// anyValue marks a parameter that accepts any type.
func anyValue(name string) Param { return Param{Name: name} }

const (
	kNum   = types.TypeF64
	kText  = types.TypeString
	kBool  = types.TypeBool
	kList  = types.TypeList
	kArray = types.TypeArray
	kTable = types.TypeLookup
	kAny   = types.TypeUnknown
	kNone  = types.TypeNull
)

func sig(module, name string, returns types.TypeKind, params ...Param) Signature {
	return Signature{Name: name, Params: params, Returns: returns, Module: module}
}

func optSig(module, name string, returns types.TypeKind, opt int, params ...Param) Signature {
	return Signature{Name: name, Params: params, Optional: opt, Returns: returns, Module: module}
}

// signatureList declares every built-in exactly once.
var signatureList = func() []Signature {
	var out []Signature

	// ── Math ──────────────────────────────────────────────────────────────
	for _, n := range []string{
		"sqrt", "abs", "floor", "ceil", "round",
		"sin", "cos", "tan", "log", "log10", "log2", "exp",
	} {
		out = append(out, sig("math", n, kNum, param("x", kNum)))
	}
	out = append(out,
		sig("math", "is_nan", kBool, param("x", kNum)),
		sig("math", "is_infinite", kBool, param("x", kNum)),
		sig("math", "pow", kNum, param("base", kNum), param("exponent", kNum)),
		sig("math", "min", kNum, param("a", kNum), param("b", kNum)),
		sig("math", "max", kNum, param("a", kNum), param("b", kNum)),
		sig("math", "random", kNum),
		sig("math", "random_between", kNum, param("min", kNum), param("max", kNum)),
	)

	// ── Strings ───────────────────────────────────────────────────────────
	for _, n := range []string{
		"uppercase", "lowercase", "casefold", "trim",
		"title", "capitalize", "swapcase", "trim_left", "trim_right",
	} {
		out = append(out, sig("string", n, kText, param("text", kText)))
	}
	for _, n := range []string{
		"is_digit", "is_alpha", "is_alnum", "is_space", "is_upper", "is_lower",
	} {
		out = append(out, sig("string", n, kBool, param("text", kText)))
	}
	out = append(out,
		sig("string", "to_number", kNum, param("text", kText)),
		// to_string and is_empty accept any value by design.
		sig("string", "to_string", kText, anyValue("value")),
		sig("string", "is_empty", kBool, anyValue("value")),
		sig("string", "split", kList, param("text", kText), param("separator", kText)),
		sig("string", "join", kText, param("list", kList, kArray), param("separator", kText)),
		sig("string", "replace", kText, param("text", kText), param("old", kText), param("new", kText)),
		sig("string", "contains", kBool, param("text", kText), param("substring", kText)),
		sig("string", "starts_with", kBool, param("text", kText), param("prefix", kText)),
		sig("string", "ends_with", kBool, param("text", kText), param("suffix", kText)),
		sig("string", "index_of", kNum, param("text", kText), param("search", kText)),
		sig("string", "substring", kText, param("text", kText), param("start", kNum), param("length", kNum)),
		sig("string", "str_repeat", kText, param("text", kText), param("n", kNum)),
		sig("string", "count_occurrences", kNum, param("text", kText), param("substring", kText)),
		optSig("string", "pad_left", kText, 1, param("text", kText), param("width", kNum), param("char", kText)),
		optSig("string", "pad_right", kText, 1, param("text", kText), param("width", kNum), param("char", kText)),
		optSig("string", "center", kText, 1, param("text", kText), param("width", kNum), param("char", kText)),
		sig("string", "zfill", kText, param("text", kText), param("width", kNum)),
	)

	// ── Number ────────────────────────────────────────────────────────────
	out = append(out,
		sig("number", "is_integer", kBool, param("x", kNum)),
		sig("number", "clamp", kNum, param("x", kNum), param("min", kNum), param("max", kNum)),
		sig("number", "sign", kNum, param("x", kNum)),
	)

	// ── List ──────────────────────────────────────────────────────────────
	// These accept an array as well as a list; the implementations always have.
	for _, n := range []string{"sort", "reverse", "unique", "flatten", "sorted_desc"} {
		out = append(out, sig("list", n, kList, param("list", kList, kArray)))
	}
	for _, n := range []string{"sum", "average", "min_value", "max_value", "product"} {
		out = append(out, sig("list", n, kNum, param("list", kList, kArray)))
	}
	out = append(out,
		sig("list", "any_true", kBool, param("list", kList, kArray)),
		sig("list", "all_true", kBool, param("list", kList, kArray)),
		// first and last return an element, whose type follows the collection.
		sig("list", "first", kAny, param("list", kList, kArray)),
		sig("list", "last", kAny, param("list", kList, kArray)),
		// count also accepts text and lookup tables.
		sig("list", "count", kNum, param("collection", kList, kArray, kTable, kText)),
		// append returns a list for a list and an array for an array.
		sig("list", "append", kAny, param("list", kList, kArray), anyValue("item")),
		sig("list", "remove", kList, param("list", kList, kArray), param("index", kNum)),
		sig("list", "insert", kList, param("list", kList, kArray), param("index", kNum), anyValue("item")),
		sig("list", "slice", kList, param("list", kList, kArray), param("start", kNum), param("end", kNum)),
		sig("list", "zip_with", kList, param("list", kList, kArray), param("other", kList, kArray)),
	)

	// ── I/O ───────────────────────────────────────────────────────────────
	out = append(out, optSig("io", "ask", kText, 1, anyValue("prompt")))

	// ── Lookup table ──────────────────────────────────────────────────────
	out = append(out,
		sig("lookup", "keys", kList, param("table", kTable)),
		sig("lookup", "values", kList, param("table", kTable)),
		sig("lookup", "table_remove", kTable, param("table", kTable), anyValue("key")),
		sig("lookup", "table_has", kBool, param("table", kTable), anyValue("key")),
		sig("lookup", "merge", kTable, param("table", kTable), param("other", kTable)),
		sig("lookup", "get_or_default", kAny, param("table", kTable), anyValue("key"), anyValue("default")),
	)

	// ── Time ──────────────────────────────────────────────────────────────
	out = append(out,
		sig("time", "current_time", kText),
		sig("time", "elapsed_time", kNum),
		sig("time", "sleep", kNone, param("seconds", kNum)),
	)

	return out
}()

// signatures indexes signatureList by function name.
var signatures = func() map[string]Signature {
	m := make(map[string]Signature, len(signatureList))
	for _, s := range signatureList {
		if _, dup := m[s.Name]; dup {
			panic("stdlib: duplicate builtin signature for " + s.Name)
		}
		m[s.Name] = s
	}
	return m
}()

// Lookup returns the signature of a built-in, and whether it exists.
func Lookup(name string) (Signature, bool) {
	s, ok := signatures[name]
	return s, ok
}

// Signatures returns every built-in signature, sorted by name. The type
// checker reads this rather than keeping its own copy.
func Signatures() []Signature {
	out := make([]Signature, len(signatureList))
	copy(out, signatureList)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Names returns every built-in name, sorted. Used by tooling that needs the
// authoritative list (completions, help, transpiler coverage tests).
func Names() []string {
	out := make([]string, 0, len(signatures))
	for n := range signatures {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// checkArity validates an argument count against a built-in's signature.
// It is called before dispatch so that no implementation ever indexes past the
// end of args, which previously produced an uncaught Go panic.
func checkArity(name string, got int) error {
	s, ok := signatures[name]
	if !ok {
		return nil // unknown name; the dispatcher reports it
	}
	minA, maxA := s.MinArgs(), s.MaxArgs()
	if got >= minA && got <= maxA {
		return nil
	}
	return vm.NewRuntimeError(fmt.Sprintf(
		"%s expects %s, got %s\n  Usage: %s",
		name, plural(minA, maxA), countWord(got), Usage(s),
	))
}

// plural renders an expected-argument-count range in English.
func plural(minA, maxA int) string {
	switch {
	case minA == maxA:
		return countWord(minA)
	case maxA == minA+1:
		return fmt.Sprintf("%s or %s", countWord(minA), countWord(maxA))
	default:
		return fmt.Sprintf("between %s and %s", countWord(minA), countWord(maxA))
	}
}

func countWord(n int) string {
	switch n {
	case 0:
		return "no arguments"
	case 1:
		return "1 argument"
	default:
		return fmt.Sprintf("%d arguments", n)
	}
}

// Usage renders a call template, e.g. `pad_left(text, width[, char])`.
func Usage(s Signature) string {
	out := s.Name + "("
	for i, p := range s.Params {
		optional := i >= s.MinArgs()
		switch {
		case i == 0 && optional:
			out += "["
		case i == 0:
			// first required parameter needs no separator
		case optional:
			out += "[, "
		default:
			out += ", "
		}
		out += p.Name
	}
	for i := 0; i < s.Optional; i++ {
		out += "]"
	}
	return out + ")"
}
